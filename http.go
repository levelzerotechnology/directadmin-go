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
)

type httpDebug struct {
	Body          string
	BodyTruncated bool
	Code          int
	Cookies       []string
	Endpoint      string
	Method        string
	Start         time.Time
}

func (c *UserContext) getRequestURLNew(endpoint string) string {
	return fmt.Sprintf("%s/api/%s", c.api.url, endpoint)
}

func (c *UserContext) getRequestURLOld(endpoint string) string {
	return fmt.Sprintf("%s/CMD_%s", c.api.url, endpoint)
}

// makeRequest is the underlying function for HTTP requests. It handles debugging statements, and simple error handling
func (c *UserContext) makeRequest(req *http.Request) ([]byte, error) {
	var debugCookies []string

	cookiesToSet := c.cookieJar.Cookies(req.URL)
	sessionCookieSet := false
	for _, cookie := range cookiesToSet {
		req.AddCookie(cookie)

		if cookie.Name == "csrftoken" {
			req.Header.Set("X-CSRFToken", cookie.Value)
		} else if cookie.Name == "session" {
			sessionCookieSet = true
		}

		if c.api.debug {
			debugCookies = append(debugCookies, cookie.String())
		}
	}

	debug := httpDebug{
		Cookies:  debugCookies,
		Endpoint: getPathWithQuery(req),
		Method:   req.Method,
		Start:    time.Now(),
	}
	defer c.api.printDebugHTTP(&debug)

	if !sessionCookieSet {
		req.SetBasicAuth(c.credentials.username, c.credentials.passkey)
	}

	resp, err := c.api.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()

	if c.api.debug {
		debug.Code = resp.StatusCode
	}

	// Required for plugin usage in particular (session and csrf token cookies).
	for _, cookie := range resp.Cookies() {
		c.cookieJar.SetCookies(req.URL, []*http.Cookie{cookie})
	}

	var responseBytes []byte

	if resp.Body != nil {
		responseBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}

		if c.api.debug {
			if len(responseBytes) > 32768 {
				debug.BodyTruncated = true
				debug.Body = string(responseBytes[:32768])
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

	req, err := http.NewRequest(strings.ToUpper(method), c.getRequestURLNew(endpoint), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	query := req.URL.Query()
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", c.api.url)
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

// makeRequestOld supports DirectAdmin's old API
func (c *UserContext) makeRequestOld(method string, endpoint string, body url.Values, object any) ([]byte, error) {
	req, err := http.NewRequest(strings.ToUpper(method), c.getRequestURLOld(endpoint), strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	query := req.URL.Query()
	query.Add("json", "yes")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.Header.Set("Referer", c.api.url)
	req.URL.RawQuery = query.Encode()

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
	req.URL.RawQuery = query.Encode()

	resp, err := c.makeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	if resp != nil && len(resp) > 0 && object != nil {
		if err = json.Unmarshal(resp, &object); err != nil {
			return nil, fmt.Errorf("error unmarshalling response: %w", err)
		}
	}

	return resp, nil
}

func (a *API) printDebugHTTP(debug *httpDebug) {
	if a.debug {
		var bodyTruncated string
		if debug.BodyTruncated {
			bodyTruncated = " (truncated)"
		}

		fmt.Printf("\nENDPOINT: %v %v\nSTATUS CODE: %v\nTIME STARTED: %v\nTIME ENDED: %v\nTIME TAKEN: %v\nCOOKIES: %s\nRESP BODY%s: %v\n", debug.Method, debug.Endpoint, debug.Code, debug.Start, time.Now(), time.Since(debug.Start), strings.Join(debug.Cookies, ";"), bodyTruncated, debug.Body)
	}
}

func getPathWithQuery(req *http.Request) string {
	if req == nil {
		return ""
	}

	if req.URL.RawQuery != "" {
		return req.URL.Path + "?" + req.URL.RawQuery
	}

	return req.URL.Path
}
