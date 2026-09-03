package directadmin

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
	"time"
)

// debugBodyLimit caps the debug body bytes to 32KB.
const debugBodyLimit = 32768

const sessionTimeout = 1 * time.Hour

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

// makeRequest is the underlying function for HTTP requests. It handles debugging statements, and simple error handling.
// It returns the response status code alongside the body; a non-2xx code is also surfaced as an error.
func (c *UserContext) makeRequest(req *http.Request) (int, []byte, error) {
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
		return 0, nil, err
	}
	defer resp.Body.Close()

	if c.api.debug {
		debug.Code = resp.StatusCode
	}

	// Required for plugin usage in particular (session and csrf token cookies).
	for _, cookie := range resp.Cookies() {
		if cookie.Name == "session" {
			c.sessionExpires = time.Now().Add(sessionTimeout)
		}

		if cookie.Path == "" {
			cookie.Path = "/"
		}

		c.cookieJar.SetCookies(req.URL, []*http.Cookie{cookie})
	}

	var responseBytes []byte

	if resp.Body != nil {
		responseBytes, err = io.ReadAll(resp.Body)
		if err != nil {
			return resp.StatusCode, nil, fmt.Errorf("error reading response body: %w", err)
		}

		if c.api.debug {
			if len(responseBytes) > debugBodyLimit {
				debug.BodyTruncated = true
				debug.Body = string(responseBytes[:debugBodyLimit])
			} else {
				debug.Body = string(responseBytes)
			}
		}
	}

	if resp.StatusCode/100 != 2 {
		return resp.StatusCode, responseBytes, fmt.Errorf("unexpected status code: %d", resp.StatusCode)
	}

	return resp.StatusCode, responseBytes, nil
}

// makeRequestNew supports DirectAdmin's new API.
func (c *UserContext) makeRequestNew(method string, endpoint string, body any, object any) ([]byte, error) {
	var bodyBytes []byte

	if body != nil {
		var err error

		bodyBytes, err = json.Marshal(body)
		if err != nil {
			return nil, fmt.Errorf("error serializing body: %w", err)
		}
	}

	req, err := c.newRequest(context.Background(), method, c.getRequestURLNew(endpoint), bytes.NewBuffer(bodyBytes))
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.URL.RawQuery = query.Encode()

	_, resp, err := c.makeRequest(req)
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

// makeRequestOld supports DirectAdmin's old API.
func (c *UserContext) makeRequestOld(method string, endpoint string, body url.Values, object any) ([]byte, error) {
	req, err := c.newRequest(context.Background(), method, c.getRequestURLOld(endpoint), strings.NewReader(body.Encode()))
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	query.Add("json", "yes")
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", "application/json")
	req.URL.RawQuery = query.Encode()

	var genericResponse apiGenericResponse

	_, resp, err := c.makeRequest(req)
	if err != nil {
		jsonErr := json.Unmarshal(resp, &genericResponse)
		if jsonErr != nil {
			return nil, fmt.Errorf("error making request: %w", err)
		}

		return nil, errors.New(genericResponse.Error + ": " + genericResponse.Result)
	}

	if resp != nil {
		// The old API reports failures as an error field with a 200, so this is checked
		// before decoding into object. A caller's struct won't carry the error field and
		// would otherwise be handed back zeroed with no error at all.
		if err = json.Unmarshal(resp, &genericResponse); err == nil && genericResponse.Error != "" {
			return nil, errors.New(genericResponse.Error + ": " + genericResponse.Result)
		}

		if object != nil {
			if err = json.Unmarshal(resp, &object); err != nil {
				return nil, fmt.Errorf("error unmarshalling response: %w", err)
			}
		}
	}

	return resp, nil
}

// errorOrResult picks whichever part of a generic response explains a failure.
func errorOrResult(response apiGenericResponse) string {
	if response.Error != "" {
		return response.Error
	}

	if response.Result != "" {
		return response.Result
	}

	encoded, err := json.Marshal(response)
	if err != nil {
		return "unknown error"
	}

	return string(encoded)
}

// newRequest builds an *http.Request with the headers common to every DirectAdmin
// request (Referer, User-Agent) already set. Callers add any method-specific headers.
func (c *UserContext) newRequest(ctx context.Context, method string, url string, body io.Reader) (*http.Request, error) {
	req, err := http.NewRequestWithContext(ctx, strings.ToUpper(method), url, body)
	if err != nil {
		return nil, fmt.Errorf("error creating request: %w", err)
	}

	req.Header.Set("Referer", c.api.url)
	req.Header.Set("User-Agent", "DirectAdmin-Go-SDK")

	return req, nil
}

// uploadFile functions for either the old or new DA API.
func (c *UserContext) uploadFile(method string, endpoint string, data []byte, object any, contentType string) ([]byte, error) {
	req, err := c.newRequest(context.Background(), method, c.api.url+endpoint, bytes.NewBuffer(data))
	if err != nil {
		return nil, err
	}

	query := req.URL.Query()
	req.Header.Set("Accept", "application/json")
	req.Header.Set("Content-Type", contentType)
	req.URL.RawQuery = query.Encode()

	_, resp, err := c.makeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("error making request: %w", err)
	}

	if len(resp) > 0 && object != nil {
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

// CustomHTTP is a thin escape hatch for making authenticated requests against
// server paths the SDK doesn't wrap directly, such as third-party plugin
// endpoints (e.g. /CMD_PLUGINS/...). Obtain one via UserContext.HTTP.
type CustomHTTP struct {
	ctx *UserContext
}

// HTTP returns a CustomHTTP bound to the user's session.
func (c *UserContext) HTTP() *CustomHTTP {
	return &CustomHTTP{ctx: c}
}

// do performs an authenticated request against the given server path (including
// any query string, e.g. "/CMD_PLUGINS/redis/ajax.raw?a=start"). It returns the
// response status code and raw body; a non-2xx code is also surfaced as an error.
func (h *CustomHTTP) do(ctx context.Context, method, path string, body io.Reader) (int, []byte, error) {
	req, err := h.ctx.newRequest(ctx, method, h.ctx.api.url+path, body)
	if err != nil {
		return 0, nil, err
	}

	return h.ctx.makeRequest(req)
}

// GET performs an authenticated GET.
func (h *CustomHTTP) GET(ctx context.Context, path string) (int, []byte, error) {
	return h.do(ctx, http.MethodGet, path, nil)
}

// POST performs an authenticated POST. No Content-Type is set by default.
func (h *CustomHTTP) POST(ctx context.Context, path string, body io.Reader) (int, []byte, error) {
	return h.do(ctx, http.MethodPost, path, body)
}

// PUT performs an authenticated PUT.
func (h *CustomHTTP) PUT(ctx context.Context, path string, body io.Reader) (int, []byte, error) {
	return h.do(ctx, http.MethodPut, path, body)
}

// DELETE performs an authenticated DELETE.
func (h *CustomHTTP) DELETE(ctx context.Context, path string) (int, []byte, error) {
	return h.do(ctx, http.MethodDelete, path, nil)
}
