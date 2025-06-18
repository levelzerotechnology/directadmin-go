package directadmin

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cast"
)

type Subdomain struct {
	Domain     string `json:"domain" yaml:"domain"`
	PhpVersion string `json:"phpVersion" yaml:"phpVersion"`
	PublicHtml string `json:"publicHtml" yaml:"publicHtml"`
	Subdomain  string `json:"subdomain" yaml:"subdomain"`
}

// CreateSubdomain (user) creates the provided subdomain for the session user. This automatically gets called if
// subdomains are included in the CreateDomain call. You cannot specify a custom directory here
func (c *UserContext) CreateSubdomain(subdomain Subdomain) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("domain", subdomain.Domain)
	body.Set("subdomain", subdomain.Subdomain)

	if _, err := c.makeRequestOld(http.MethodPost, "API_SUBDOMAINS?action=create", body, &response); err != nil {
		return fmt.Errorf("failed to create subdomain: %v", err)
	}

	if response.Result != "Subdomain created" {
		return fmt.Errorf("failed to create subdomain: %v", response.Result)
	}

	return nil
}

// DeleteSubdomains (user) deletes all the specified subdomain for the session user
func (c *UserContext) DeleteSubdomains(deleteData bool, domain string, subdomains ...string) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("domain", domain)

	if deleteData {
		body.Set("contents", "yes")
	} else {
		body.Set("contents", "no")
	}

	for index, subdomain := range subdomains {
		body.Set("select"+cast.ToString(index), subdomain)
	}

	if _, err := c.makeRequestOld(http.MethodPost, "API_SUBDOMAINS?action=delete", body, &response); err != nil {
		return err
	}

	if response.Success != "Subdomains deleted" {
		return fmt.Errorf("failed to delete subdomain: %v", response.Result)
	}

	return nil
}

// ListSubdomains (user) returns an array of all subdomains for the given domain
func (c *UserContext) ListSubdomains(domain string) (subdomainList []string, err error) {
	if _, err = c.makeRequestOld(http.MethodGet, "API_SUBDOMAINS?bytes=yes&domain="+domain, nil, &subdomainList); err != nil {
		return nil, err
	}

	return subdomainList, nil
}

func (c *UserContext) UpdateSubdomainRoot(subdomain Subdomain) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("domain", subdomain.Domain)
	body.Set("subdomain", subdomain.Subdomain)
	body.Set("public_html", subdomain.PublicHtml)

	if _, err := c.makeRequestOld(http.MethodPost, "SUBDOMAIN?action=document_root_override", body, &response); err != nil {
		return fmt.Errorf("failed to update subdomain root: %v", err)
	}

	if response.Success != "Success" {
		return fmt.Errorf("failed to update subdomain root: %v", response.Result)
	}

	return nil
}
