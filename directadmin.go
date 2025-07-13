package directadmin

import (
	"errors"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	AccountRoleAdmin    = "admin"
	AccountRoleReseller = "reseller"
	AccountRoleUser     = "user"
)

type API struct {
	cache struct {
		domains            map[string]Domain // Domain name is key.
		domainsMutex       *sync.Mutex
		emailAccounts      map[string]EmailAccount // Domain name is key.
		emailAccountsMutex *sync.Mutex
		packages           map[string]Package // Package name is key.
		packagesMutex      *sync.Mutex
		users              map[string]User // Username is key.
		usersMutex         *sync.Mutex
	}
	cacheEnabled bool
	debug        bool
	httpClient   *http.Client
	parsedURL    *url.URL
	url          string
}

type (
	apiGenericResponse struct {
		Error   string `json:"error,omitempty"`
		Result  string `json:"result,omitempty"`
		Success string `json:"success,omitempty"`
	}

	apiGenericResponseNew struct {
		Message string `json:"message"`
		Type    string `json:"type"`
	}
)

func New(serverURL string, timeout time.Duration, cacheEnabled bool, debug bool) (*API, error) {
	parsedURL, err := url.ParseRequestURI(serverURL)
	if err != nil {
		return nil, err
	}

	if parsedURL.Host == "" {
		return nil, errors.New("invalid host provided, ensure that the host is a full URL e.g. https://your-ip-address:2222")
	}

	api := API{
		cacheEnabled: cacheEnabled,
		debug:        debug,
		parsedURL:    parsedURL,
		url:          parsedURL.String(),
	}

	if cacheEnabled {
		api.cache.domains = make(map[string]Domain)
		api.cache.emailAccounts = make(map[string]EmailAccount)
		api.cache.packages = make(map[string]Package)
		api.cache.users = make(map[string]User)
	}

	api.httpClient = &http.Client{Timeout: timeout}

	return &api, nil
}

func (a *API) GetURL() string {
	return a.url
}
