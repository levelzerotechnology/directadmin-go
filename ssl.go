package directadmin

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cast"
)

// IssueSSL (user) requests a lets encrypt certificate for the given hostnames
func (c *UserContext) IssueSSL(domain string, hostnamesToCertify ...string) error {
	var response apiGenericResponse

	if len(hostnamesToCertify) == 0 {
		return errors.New("at least one hostname is required for the certificate")
	}

	body := url.Values{
		"type":          {"create"},
		"request":       {"letsencrypt"},
		"name":          {hostnamesToCertify[0]},
		"domain":        {domain},
		"keysize":       {"secp384r1"},
		"encryption":    {"sha256"},
		"wildcard":      {"no"},
		"background":    {"auto"},
		"action":        {"save"},
		"acme_provider": {"letsencrypt"},
	}

	for index, certDomain := range hostnamesToCertify {
		body.Set("le_select"+cast.ToString(index), certDomain)
	}

	if _, err := c.api.makeRequest(http.MethodPost, "API_SSL", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "Certificate and Key Saved." {
		return fmt.Errorf("failed to issue SSL certificate: %v", response.Result)
	}

	return nil
}
