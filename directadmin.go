package directadmin

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"sync"
	"time"
)

const (
	AccountRoleAdmin    = "Admin"
	AccountRoleReseller = "Reseller"
	AccountRoleUser     = "User"
)

type API struct {
	cache struct {
		domains            map[string]Domain // domain name is key
		domainsMutex       *sync.Mutex
		emailAccounts      map[string]EmailAccount // domain name is key
		emailAccountsMutex *sync.Mutex
		packages           map[string]Package // package name is key
		packagesMutex      *sync.Mutex
		users              map[string]User // username is key
		usersMutex         *sync.Mutex
	}
	cacheEnabled bool
	debug        bool
	httpClient   *http.Client
	ssl          bool
	timeout      time.Duration
	url          string
}

type apiGenericResponse struct {
	Error   string `json:"error,omitempty"`
	Result  string `json:"result,omitempty"`
	Success string `json:"success,omitempty"`
}

type apiGenericResponseN struct {
	Message string `json:"message"`
	Type    string `json:"type"`
}

// TODO: implement caching layer which can be enabled/disabled on New()
// essentially, for domains it'd have map[string]Domain at the API level
// then if any user called from the API, it would check the cache first
// would either need a cache lifetime field added to domains, or
// add an additional map for lifetime checks

func New(serverUrl string, timeout time.Duration, cacheEnabled bool, debug bool) (*API, error) {
	parsedUrl, err := url.ParseRequestURI(serverUrl)
	if err != nil {
		return nil, err
	}

	if parsedUrl.Host == "" {
		return nil, errors.New("invalid host provided, ensure that the host is a full URL e.g. https://your-ip-address:2222")
	}

	api := API{
		cacheEnabled: cacheEnabled,
		debug:        debug,
		url:          parsedUrl.String(),
	}

	if cacheEnabled {
		api.cache.domains = make(map[string]Domain)
		api.cache.emailAccounts = make(map[string]EmailAccount)
		api.cache.packages = make(map[string]Package)
		api.cache.users = make(map[string]User)
	}

	switch parsedUrl.Scheme {
	case "http":
		api.ssl = false
	case "https":
		api.ssl = true
	default:
		return nil, fmt.Errorf("invalid url scheme: %v", parsedUrl.Scheme)
	}

	api.httpClient = &http.Client{Timeout: timeout}

	return &api, nil
}

func (a *API) GetURL() string {
	return a.url
}
