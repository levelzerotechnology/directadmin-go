package directadmin

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/spf13/cast"
)

type Domain struct {
	Active             bool     `json:"active" yaml:"active"`
	BandwidthQuota     int      `json:"bandwidthQuota" yaml:"bandwidthQuota"`
	BandwidthUsage     int      `json:"bandwidthUsage" yaml:"bandwidthUsage"`
	CgiEnabled         bool     `json:"cgiEnabled" yaml:"cgiEnabled"`
	DefaultDomain      bool     `json:"defaultDomain" yaml:"defaultDomain"`
	DiskQuota          int      `json:"diskQuota" yaml:"diskQuota"`
	DiskUsage          int      `json:"diskUsage" yaml:"diskUsage"`
	Domain             string   `json:"domain" yaml:"domain"`
	IPAddresses        []string `json:"ipAddresses" yaml:"ipAddresses"`
	ModSecurityEnabled bool     `json:"modSecurityEnabled" yaml:"modSecurityEnabled"`
	OpenBaseDirEnabled bool     `json:"openBaseDirEnabled" yaml:"openBaseDirEnabled"`
	PhpEnabled         bool     `json:"phpEnabled" yaml:"phpEnabled"`
	PhpSelectorEnabled bool     `json:"phpSelectorEnabled" yaml:"phpSelectorEnabled"`
	PhpVersion         string   `json:"phpVersion" yaml:"phpVersion"`
	SafeMode           bool     `json:"safeMode" yaml:"safeMode"`
	SslEnabled         bool     `json:"sslEnabled" yaml:"sslEnabled"`
	Subdomains         []string `json:"subdomains" yaml:"subdomains"`
	SubdomainUsage     int      `json:"subdomainUsage" yaml:"subdomainUsage"`
	Suspended          bool     `json:"suspended" yaml:"suspended"`
	Username           string   `json:"username" yaml:"username"`
}

// AddDomainIP (user) adds an additional IP to a domain.
func (c *UserContext) AddDomainIP(domain string, ip string, createDNSRecords bool) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("action", "multi_ip")
	body.Set("add", "yes")
	body.Set("domain", domain)
	body.Set("ip", ip)

	if createDNSRecords {
		body.Set("dns", "yes")
	} else {
		body.Set("dns", "no")
	}

	if _, err := c.makeRequestOld(http.MethodPost, "DOMAIN", body, &response); err != nil {
		return fmt.Errorf("failed to add IP to domain: %v", err)
	}

	if response.Success != "IP Added" {
		return fmt.Errorf("failed to add IP to domain: %v", response.Result)
	}

	return nil
}

// CheckDomainExists (user) checks if the given domain exists on the server
func (c *UserContext) CheckDomainExists(domain string) error {
	return c.checkObjectExists(url.Values{
		"type":  {"domain"},
		"value": {domain},
	})
}

// CreateDomain (user) creates the provided domain for the session user
func (c *UserContext) CreateDomain(domain Domain) error {
	var response apiGenericResponse

	rawDomainData := domain.translate()

	body := url.Values{}
	body.Set("domain", rawDomainData.Domain)
	body.Set("ubandwidth", rawDomainData.BandwidthQuota)
	body.Set("uquota", rawDomainData.DiskQuota)
	body.Set("cgi", rawDomainData.CgiEnabled)
	body.Set("php", rawDomainData.PhpEnabled)
	body.Set("ssl", rawDomainData.SslEnabled)

	if _, err := c.makeRequestOld(http.MethodPost, "API_DOMAIN?action=create", body, &response); err != nil {
		return err
	}

	if response.Success != "Domain Created" {
		return fmt.Errorf("failed to create domain: %v", response.Result)
	}

	if len(domain.Subdomains) > 0 {
		for _, subdomain := range domain.Subdomains {
			if err := c.CreateSubdomain(Subdomain{
				Domain:    domain.Domain,
				Subdomain: subdomain,
			}); err != nil {
				return fmt.Errorf("successfully created domain, but failed to create subdomain %v: %v", subdomain, err)
			}
		}
	}

	// cache domain
	if c.api.cacheEnabled {
		go func(domainToCache Domain) {
			c.api.cache.domainsMutex.Lock()
			c.api.cache.domains[domain.Domain] = domain
			c.api.cache.domainsMutex.Unlock()
		}(domain)
	}

	return nil
}

// DeleteDomains (user) deletes all the specified domains for the session user
func (c *UserContext) DeleteDomains(deleteData bool, domains ...string) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("delete", "yes")
	body.Set("confirmed", "yes")
	body.Set("delete_data_aware", "yes")

	if deleteData {
		body.Set("delete_data", "yes")
	} else {
		body.Set("delete_data", "no")
	}

	for index, domain := range domains {
		body.Set("select"+cast.ToString(index), domain)
	}

	if _, err := c.makeRequestOld(http.MethodPost, "API_DOMAIN?action=select", body, &response); err != nil {
		return err
	}

	if response.Success != "Domain Deletion Results" {
		return fmt.Errorf("failed to delete domain: %v", response.Result)
	}

	// remove domains from cache
	if c.api.cacheEnabled {
		for _, domain := range domains {
			c.api.cache.domainsMutex.Lock()
			delete(c.api.cache.domains, domain)
			c.api.cache.domainsMutex.Unlock()
		}
	}

	return nil
}

// GetDomain (user) returns the single specified domain
func (c *UserContext) GetDomain(domainName string) (Domain, error) {
	// check if domain is in cache
	if c.api.cacheEnabled {
		if cachedDomain, ok := c.api.cache.domains[domainName]; ok {
			return cachedDomain, nil
		}
	}

	var rawDomains map[string]rawDomain

	if _, err := c.makeRequestOld(http.MethodGet, "API_ADDITIONAL_DOMAINS?bytes=yes&domain="+domainName, nil, &rawDomains); err != nil {
		return Domain{}, err
	}

	rawDomainData := rawDomains[domainName]

	if rawDomainData.SubdomainUsage != "0" {
		subdomains, err := c.ListSubdomains(rawDomainData.Domain)
		if err != nil {
			return Domain{}, err
		}

		rawDomainData.Subdomains = subdomains
	} else {
		rawDomainData.Subdomains = []string{}
	}

	if _, err := c.makeRequestOld(http.MethodGet, "API_ADDITIONAL_DOMAINS?bytes=yes&action=view&domain="+domainName, nil, &rawDomainData.ExtraData); err != nil {
		return Domain{}, err
	}

	return rawDomainData.translate(), nil
}

// GetDomains (user) returns the session user's domains
func (c *UserContext) GetDomains() ([]Domain, error) {
	var domains []Domain
	var rawDomains map[string]rawDomain

	if _, err := c.makeRequestOld(http.MethodGet, "API_ADDITIONAL_DOMAINS?bytes=yes", nil, &rawDomains); err != nil {
		return nil, fmt.Errorf("failed to get domains: %v", err)
	}

	if len(rawDomains) == 0 {
		return nil, errors.New("no domains were found")
	}

	// DA doesn't return the PHP version or mod security's status when returning all the domains, so we have to re-call
	// the endpoint for each domain. We can't call CMD_API_SHOW_DOMAINS instead in the call above because DA returns
	// different quota data for some reason when viewing a single domain
	var errs []error
	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Add(len(rawDomains))

	for _, rawDomainData := range rawDomains {
		// convert to local variable to prevent variable overwrite
		domainToProcess := rawDomainData

		// check if domainToProcess is in cache
		if c.api.cacheEnabled {
			if cachedDomain, ok := c.api.cache.domains[domainToProcess.Domain]; ok {
				mu.Lock()
				domains = append(domains, cachedDomain)
				mu.Unlock()
				wg.Done()

				continue
			}
		}

		go func(rawDomainData rawDomain) {
			defer wg.Done()

			if _, err := c.makeRequestOld(http.MethodGet, "API_ADDITIONAL_DOMAINS?action=view&bytes=yes&domain="+rawDomainData.Domain, nil, &rawDomainData.ExtraData); err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}

			if rawDomainData.SubdomainUsage != "0" {
				subdomains, err := c.ListSubdomains(rawDomainData.Domain)
				if err != nil {
					mu.Lock()
					errs = append(errs, err)
					mu.Unlock()
					return
				}

				rawDomainData.Subdomains = subdomains
			} else {
				rawDomainData.Subdomains = []string{}
			}

			mu.Lock()
			domains = append(domains, rawDomainData.translate())
			mu.Unlock()

			// cache domain
			if c.api.cacheEnabled {
				go func(domainToCache Domain) {
					c.api.cache.domainsMutex.Lock()
					c.api.cache.domains[domainToCache.Domain] = domainToCache
					c.api.cache.domainsMutex.Unlock()
				}(rawDomainData.translate())
			}
		}(domainToProcess)
	}

	wg.Wait()

	if len(errs) > 0 {
		counter := 0
		var errStrings []string

		for _, err := range errs {
			counter++
			errStrings = append(errStrings, "error "+cast.ToString(counter)+": "+err.Error())
		}

		return nil, errors.New(strings.Join(errStrings, "; "))
	}

	if len(domains) == 0 {
		return nil, errors.New("no domains were found")
	}

	return domains, nil
}

// ListDomains (user) returns an array of all domains for the session user
func (c *UserContext) ListDomains() (domainList []string, err error) {
	if _, err = c.makeRequestOld(http.MethodGet, "API_SHOW_DOMAINS?bytes=yes", nil, &domainList); err != nil {
		return nil, err
	}

	return domainList, nil
}

// SetDefaultDomain (user) sets the default domain for the session user
func (c *UserContext) SetDefaultDomain(domain string) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("select0", domain)
	body.Set("default", "yes")

	if _, err := c.makeRequestOld(http.MethodPost, "API_DOMAIN?action=select", body, &response); err != nil {
		return err
	}

	if response.Success != "Your default domain has been set" {
		return fmt.Errorf("failed to set default domain: %v", response.Result)
	}

	return nil
}

// UpdateDomain (user) accepts a Domain object and updates the version on DA with it
func (c *UserContext) UpdateDomain(domain Domain) error {
	var response apiGenericResponse

	rawDomainData := domain.translate()

	body := url.Values{}
	body.Set("domain", rawDomainData.Domain)
	body.Set("ubandwidth", rawDomainData.BandwidthQuota)
	body.Set("uquota", rawDomainData.DiskQuota)
	body.Set("cgi", rawDomainData.CgiEnabled)
	body.Set("php", rawDomainData.PhpEnabled)
	body.Set("ssl", rawDomainData.SslEnabled)

	if _, err := c.makeRequestOld(http.MethodPost, "API_DOMAIN?action=modify", body, &response); err != nil {
		return err
	}

	if response.Success != "The domain has been modified" {
		return fmt.Errorf("failed to update domain: %v", response.Result)
	}

	return nil
}
