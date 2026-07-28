package directadmin

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cast"
)

// ACMEConfig is a domain's automatic TLS (ACME) configuration.
type ACMEConfig struct {
	DNSEnvironment map[string]string `json:"dnsEnvironment" yaml:"dnsEnvironment"`
	DNSProvider    string            `json:"dnsProvider" yaml:"dnsProvider"`
	Enabled        bool              `json:"enabled" yaml:"enabled"`
	KeyType        string            `json:"keyType" yaml:"keyType"` // e.g. "ec256", "rsa2048".
	PreferWildcard bool              `json:"preferWildcard" yaml:"preferWildcard"`
	Provider       string            `json:"provider" yaml:"provider"` // e.g. "letsencrypt".
	SkipDNSNames   []string          `json:"skipDNSNames" yaml:"skipDNSNames"`
}

// normalize fills nil maps/slices, which DA rejects when sent as null.
func (a ACMEConfig) normalize() ACMEConfig {
	if a.DNSEnvironment == nil {
		a.DNSEnvironment = map[string]string{}
	}

	if a.SkipDNSNames == nil {
		a.SkipDNSNames = []string{}
	}

	return a
}

// IssueSSL (user) requests a lets encrypt certificate for the given hostnames.
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

	if _, err := c.makeRequestOld(http.MethodPost, "API_SSL", body, &response); err != nil {
		return err
	}

	if response.Success != "Certificate and Key Saved." {
		return fmt.Errorf("failed to issue SSL certificate: %v", response.Result)
	}

	return nil
}

// ProvisionCerts (user) triggers ACME certificate provisioning for the given domain
// using its saved ACMEConfig. DirectAdmin does this synchronously and it can take
// several minutes, so the API needs a suitably long timeout. The raw response body
// is returned as its shape varies by provider.
func (c *UserContext) ProvisionCerts(domain string) ([]byte, error) {
	resp, err := c.makeRequestNew(http.MethodPost, "domain-tls/"+domain+"/provision-certs", nil, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to provision certificates for %v: %w", domain, err)
	}

	return resp, nil
}

// SetACMEConfig (user) replaces the ACME config for the given domain.
func (c *UserContext) SetACMEConfig(domain string, config ACMEConfig) error {
	if _, err := c.makeRequestNew(http.MethodPut, "domain-tls/"+domain+"/acme-config", config.normalize(), nil); err != nil {
		return fmt.Errorf("failed to set ACME config for %v: %w", domain, err)
	}

	return nil
}
