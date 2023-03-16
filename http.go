package directadmin

import (
	"bytes"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"os"
	"strings"
	"time"
)

func (a *API) makeRequest(method string, endpoint string, accountCredentials credentials, body url.Values, object any, writeToFile ...*os.File) ([]byte, error) {
	defer a.queryTime(endpoint, time.Now())

	var responseBytes []byte

	req, err := http.NewRequest(strings.ToUpper(method), a.url+"/CMD_"+endpoint, strings.NewReader(body.Encode()))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	query := req.URL.Query()
	query.Add("json", "yes")

	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(accountCredentials.username, accountCredentials.passkey)
	req.URL.RawQuery = query.Encode()

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if a.debug {
		fmt.Printf("HTTP STATUS: %v\n", resp.Status)
	}

	if resp.Body != nil {
		defer resp.Body.Close()

		responseBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}

		if a.debug {
			print(string(responseBytes))
		}
	}

	// unmarshal object into generic response object to check if there's an error
	var genericResponse apiGenericResponse

	if err = json.Unmarshal(responseBytes, &genericResponse); err == nil {
		if genericResponse.Error != "" {
			return nil, errors.New(genericResponse.Error + ": " + genericResponse.Result)
		}
	}

	if len(responseBytes) > 0 {
		if len(writeToFile) > 0 {
			if _, err = writeToFile[0].Write(responseBytes); err != nil {
				return nil, fmt.Errorf("error writing to file: %w", err)
			}
		} else if object != nil {
			if err = json.Unmarshal(responseBytes, &object); err != nil {
				return nil, fmt.Errorf("error unmarshaling response: %w", err)
			}
		}
	}

	return responseBytes, nil
}

// makeRequestN supports DA's new API
func (a *API) makeRequestN(method string, endpoint string, accountCredentials credentials, body any, object any) ([]byte, error) {
	defer a.queryTime(endpoint, time.Now())

	var err error
	var requestBytes, responseBytes []byte

	if body != nil {
		requestBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error marshalling body: %w", err)
		}
	}

	req, err := http.NewRequest(strings.ToUpper(method), a.url+"/api/"+endpoint, bytes.NewBuffer(requestBytes))
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.SetBasicAuth(accountCredentials.username, accountCredentials.passkey)

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if a.debug {
		fmt.Printf("HTTP STATUS: %v\n", resp.Status)
	}

	if resp.Body != nil {
		defer resp.Body.Close()

		responseBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}

		if a.debug {
			print(string(responseBytes))
		}
	}

	// ignore unmarshal error, as DA's responses are consistent across endpoints
	if resp.StatusCode/100 == 2 {
		// unmarshal to object if one is provided
		if object != nil && len(responseBytes) > 0 {
			if err = json.Unmarshal(responseBytes, &object); err != nil {
				return nil, fmt.Errorf("error unmarshaling response: %w", err)
			}
		}
	} else {
		// unmarshal object into generic response object to check if there's an error
		var genericResponse apiGenericResponseN

		if err = json.Unmarshal(responseBytes, &genericResponse); err == nil {
			if genericResponse.Message != "" {
				return nil, errors.New(genericResponse.Type + ": " + genericResponse.Message)
			}
			return nil, errors.New(genericResponse.Type)
		}
	}

	return responseBytes, nil
}

func (a *API) uploadFile(endpoint string, accountCredentials credentials, body *bytes.Buffer, object any, contentType string) ([]byte, error) {
	defer a.queryTime(endpoint, time.Now())

	var responseBytes []byte

	req, err := http.NewRequest(http.MethodPost, a.url+"/CMD_"+endpoint, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	query := req.URL.Query()
	query.Add("json", "yes")

	req.Header.Set("Content-Type", contentType)
	req.SetBasicAuth(accountCredentials.username, accountCredentials.passkey)
	req.URL.RawQuery = query.Encode()

	resp, err := a.httpClient.Do(req)
	if err != nil {
		return nil, err
	}

	defer resp.Body.Close()

	if a.debug {
		fmt.Printf("HTTP STATUS: %v\n", resp.Status)
	}

	if resp.Body != nil {
		defer resp.Body.Close()

		responseBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return nil, fmt.Errorf("error reading response body: %w", err)
		}

		if a.debug {
			print(string(responseBytes))
		}
	}

	// unmarshal to object if one is provided
	if object != nil && len(responseBytes) > 0 {
		if err = json.Unmarshal(responseBytes, &object); err != nil {
			return nil, fmt.Errorf("error unmarshaling response: %w", err)
		}
	}
	// unmarshal object into generic response object to check if there's an error
	var genericResponse apiGenericResponse

	if err = json.Unmarshal(responseBytes, &genericResponse); err == nil {
		if genericResponse.Error != "" {
			return nil, errors.New(genericResponse.Error + ": " + genericResponse.Result)
		}
	}

	return responseBytes, nil
}

func (a *API) queryTime(endpoint string, startTime time.Time) {
	if a.debug {
		fmt.Printf("ENDPOINT: %v\nTIME STARTED: %v\nTIME ENDED: %v\nTIME TAKEN: %v\n\n", endpoint, startTime, time.Now(), time.Since(startTime))
	}
}
