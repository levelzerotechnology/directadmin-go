package directadmin

import (
	"fmt"
	"github.com/spf13/cast"
	"net/http"
	"net/url"
	"strings"
)

func (c *UserContext) CreateEmailForwarder(domain string, user string, emails ...string) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("domain", domain)
	body.Set("email", strings.Join(emails, ","))
	body.Set("user", user)

	if _, err := c.api.makeRequest(http.MethodPost, "API_EMAIL_FORWARDERS?action=create", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "Forwarder created" {
		return fmt.Errorf("failed to create email account: %v", response.Result)
	}

	return nil
}

// GetEmailForwarders (user) returns an array of email forwarders belonging to the provided domain
func (c *UserContext) GetEmailForwarders(domain string) (map[string][]string, error) {
	var emailForwarders = make(map[string][]string)

	if _, err := c.api.makeRequest(http.MethodGet, "API_EMAIL_FORWARDERS?domain="+domain, c.credentials, nil, &emailForwarders); err != nil {
		return nil, err
	}

	return emailForwarders, nil
}

func (c *UserContext) DeleteEmailForwarders(domain string, names ...string) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("domain", domain)

	for index, name := range names {
		body.Set("select"+cast.ToString(index), name)
	}

	if _, err := c.api.makeRequest(http.MethodPost, "API_EMAIL_FORWARDERS?action=delete", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "Forwarders deleted" {
		return fmt.Errorf("failed to delete email forwarder: %v", response.Result)
	}

	return nil
}

func (c *UserContext) UpdateEmailForwarder(domain string, user string, emails ...string) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("domain", domain)
	body.Set("email", strings.Join(emails, ","))
	body.Set("user", user)

	if _, err := c.api.makeRequest(http.MethodPost, "API_EMAIL_FORWARDERS?action=modify", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "Forwarder updated" {
		return fmt.Errorf("failed to update email forwarder: %v", response.Result)
	}

	return nil
}
