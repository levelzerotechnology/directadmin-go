package directadmin

import (
	"bytes"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/goccy/go-json"
	"github.com/spf13/cast"
)

type httpDebug struct {
	Body     string
	Code     int
	Endpoint string
	Start    time.Time
}

// makeRequest is the underlying function for HTTP requests. It handles debugging statements, and simple error handling
func (c *UserContext) makeRequest(req *http.Request) ([]byte, error) {
	debug := httpDebug{
		Endpoint: req.URL.Path,
		Start:    time.Now(),
	}
	defer c.api.printDebugHTTP(&debug)

	resp, err := c.api.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if c.api.debug {
		debug.Code = resp.StatusCode
	}

	// exists solely for user session switching when logging in as a user under a reseller
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "session" {
			c.sessionID = cookie.Value
			break
		}
	}

	var responseBytes []byte

	if resp.Body != nil {
		responseBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}

		if c.api.debug {
			if len(responseBytes) > 32768 {
				debug.Body = "body too long for debug: " + cast.ToString(len(responseBytes))
			} else {
				debug.Body = string(responseBytes)
			}
		}
	}

	if resp.StatusCode/100 != 2 {
		return responseBytes, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return responseBytes, nil
}

// makeRequestNew supports DirectAdmin's new API
func (c *UserContext) makeRequestNew(method string, endpoint string, body any, object any) ([]byte, error) {
	var bodyBytes []byte

	if body != nil {
		var err error

		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error serializing body: %w", err)
		}
	}

	req, err := http.NewRequest(strings.ToUpper(method), c.api.url+"/api/"+endpoint, bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	query := req.URL.Query()
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", c.api.url)
	req.URL.RawQuery = query.Encode()

	if c.sessionID != "" {
		req.AddCookie(&http.Cookie{Name: "session", Value: c.sessionID})
	} else {
		req.SetBasicAuth(c.credentials.username, c.credentials.passkey)
	}

	resp, err := c.makeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	var genericResponse apiGenericResponseN

	if resp != nil {
		if object != nil {
			if err = json.Unmarshal(resp, &object); err != nil {
				return nil, fmt.Errorf("error unmarshalling response: %w", err)
			}
		} else if err = json.Unmarshal(resp, &genericResponse); err == nil {
			if genericResponse.Message != "" {
				return nil, errors.New(genericResponse.Type + ": " + genericResponse.Message)
			}
			return nil, errors.New(genericResponse.Type)
		}
	}

	return resp, nil
}

// makeRequestOld supports DirectAdmin's old API
func (c *UserContext) makeRequestOld(method string, endpoint string, body url.Values, object any) ([]byte, error) {
	req, err := http.NewRequest(strings.ToUpper(method), c.api.url+"/CMD_"+endpoint, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	query := req.URL.Query()
	query.Add("json", "yes")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", c.api.url)
	req.URL.RawQuery = query.Encode()

	if c.sessionID != "" {
		req.AddCookie(&http.Cookie{Name: "session", Value: c.sessionID})
	} else {
		req.SetBasicAuth(c.credentials.username, c.credentials.passkey)
	}

	var genericResponse apiGenericResponse

	resp, err := c.makeRequest(req)
	if err != nil {
		jsonErr := json.Unmarshal(resp, &genericResponse)
		if jsonErr != nil {
			return nil, fmt.Errorf("error making request: %w", err)
		}

		return nil, errors.New(genericResponse.Error + ": " + genericResponse.Result)
	}

	if resp != nil {
		if object != nil {
			if err = json.Unmarshal(resp, &object); err != nil {
				return nil, fmt.Errorf("error unmarshalling response: %w", err)
			}
		} else if err = json.Unmarshal(resp, &genericResponse); err == nil && genericResponse.Error != "" {
			return nil, errors.New(genericResponse.Error + ": " + genericResponse.Result)
		}
	}

	return resp, nil
}

// uploadFile functions for either the old or new DA API
func (c *UserContext) uploadFile(method string, endpoint string, data []byte, object any, contentType string) ([]byte, error) {
	req, err := http.NewRequest(strings.ToUpper(method), c.api.url+endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	query := req.URL.Query()
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", contentType)
	req.Header.Set("Referer", c.api.url)
	req.SetBasicAuth(c.credentials.username, c.credentials.passkey)
	req.URL.RawQuery = query.Encode()

	resp, err := c.makeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	if resp != nil && object != nil {
		if err = json.Unmarshal(resp, &object); err != nil {
			return nil, fmt.Errorf("error unmarshalling response: %w", err)
		}
	}

	return resp, nil
}

func (a *API) printDebugHTTP(debug *httpDebug) {
	if a.debug {
		fmt.Printf("\nENDPOINT: %v\nSTATUS CODE: %v\nTIME STARTED: %v\nTIME ENDED: %v\nTIME TAKEN: %v\nBODY: %v\n", debug.Endpoint, debug.Code, debug.Start, time.Now(), time.Since(debug.Start), debug.Body)
	}
}
