package directadmin

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cast"
)

type EmailAccount struct {
	DiskQuota int    `json:"diskQuota" yaml:"diskQuota"`
	DiskUsage int    `json:"diskUsage" yaml:"diskUsage"`
	Domain    string `json:"domain" yaml:"domain"`
	Password  string `json:"password" yaml:"password"`
	SendQuota int    `json:"sendQuota" yaml:"sendQuota"`
	SendUsage int    `json:"sendUsage" yaml:"sendUsage"`
	Suspended bool   `json:"suspended" yaml:"suspended"`
	Username  string `json:"username" yaml:"username"`
}

// CreateEmailAccount (user) creates the given email account.
func (c *UserContext) CreateEmailAccount(emailAccount EmailAccount) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("domain", emailAccount.Domain)
	body.Set("user", emailAccount.Username)
	body.Set("passwd", emailAccount.Password)
	body.Set("passwd2", emailAccount.Password)
	body.Set("quota", cast.ToString(emailAccount.DiskQuota))
	body.Set("limit", cast.ToString(emailAccount.SendQuota))

	if _, err := c.makeRequestOld(http.MethodPost, "API_POP?action=create", body, &response); err != nil {
		return err
	}

	if response.Success != "Email account created" {
		return fmt.Errorf("failed to create email account: %v", response.Result)
	}

	return nil
}

func (c *UserContext) DeleteEmailAccount(domain string, name string) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("domain", domain)
	body.Set("user", name)

	if _, err := c.makeRequestOld(http.MethodPost, "API_POP?action=delete", body, &response); err != nil {
		return err
	}

	if response.Success != "E-Mail Accounts Deleted" {
		return fmt.Errorf("failed to delete email account: %v", response.Result)
	}

	return nil
}

// GetEmailAccounts (user) returns an array of email accounts belonging to the provided domain.
func (c *UserContext) GetEmailAccounts(domain string) ([]EmailAccount, error) {
	var emailAccounts []EmailAccount
	rawEmailAccounts := struct {
		EmailAccounts map[string]struct {
			Sent      any    `json:"sent"`
			Suspended string `json:"suspended"`
			Usage     struct {
				DiskQuota string `json:"quota"`
				DiskUsage string `json:"usage"`
			} `json:"usage"`
			Username string `json:"account"`
		} `json:"emails"`
	}{}

	if _, err := c.makeRequestOld(http.MethodGet, "EMAIL_POP?bytes=yes&domain="+domain, nil, &rawEmailAccounts); err != nil {
		return nil, err
	}

	for id, emailAccount := range rawEmailAccounts.EmailAccounts {
		if id != "info" {
			email := EmailAccount{
				DiskQuota: cast.ToInt(emailAccount.Usage.DiskQuota),
				DiskUsage: cast.ToInt(emailAccount.Usage.DiskUsage),
				Domain:    domain,
				Suspended: parseOnOff(emailAccount.Suspended),
				Username:  emailAccount.Username,
			}

			// This is necessary because DA either returns a string, or a map with the disk usage in a "usage" field
			// depending on which usage endpoint we hit.
			if emailAccountSent, ok := emailAccount.Sent.(map[string]any); ok {
				email.SendQuota = cast.ToInt(emailAccountSent["send_limit"])
				email.SendUsage = cast.ToInt(emailAccountSent["sent"])
			}

			emailAccounts = append(emailAccounts, email)
		}
	}

	if len(emailAccounts) == 0 {
		return nil, errors.New("no email accounts were found")
	}

	return emailAccounts, nil
}

// ToggleDKIM (user) sets DKIM for the given domain.
func (c *UserContext) ToggleDKIM(domain string, status bool) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("action", "set_dkim")
	body.Set("domain", domain)

	if status {
		body.Set("enable", "yes")
	} else {
		body.Set("disable", "yes")
	}

	if _, err := c.makeRequestOld(http.MethodPost, "API_EMAIL_POP", body, &response); err != nil {
		return err
	}

	if response.Success != "Success" {
		return fmt.Errorf("failed to toggle DKIM state: %v", response.Result)
	}

	return nil
}

// UpdateEmailAccount (user) updates/overwrites the given email account.
func (c *UserContext) UpdateEmailAccount(emailAccount EmailAccount) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("domain", emailAccount.Domain)
	body.Set("user", emailAccount.Username)
	body.Set("passwd", emailAccount.Password)
	body.Set("passwd2", emailAccount.Password)
	body.Set("quota", cast.ToString(emailAccount.DiskQuota))
	body.Set("limit", cast.ToString(emailAccount.SendQuota))

	if _, err := c.makeRequestOld(http.MethodPost, "API_POP?action=modify", body, &response); err != nil {
		return err
	}

	if response.Success != "Account password changed" {
		return fmt.Errorf("failed to update email account: %v", response.Result)
	}

	return nil
}

// UseInternalMailHandler tells the server to use the local mail handler for the given domain. If this is enabled, other
// domains on the server that email this domain will use the server's local mail handler to deliver the email, rather
// than looking up the domain's MX records. This is fine if your email is being hosted on the same server, but not
// otherwise.
func (c *UserContext) UseInternalMailHandler(domain string, enable bool) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("domain", domain)

	if enable {
		body.Set("internal", "yes")
	} else {
		body.Set("internal", "no")
	}

	if _, err := c.makeRequestOld(http.MethodPost, "API_DNS_MX?action=internal", body, &response); err != nil {
		return err
	}

	if response.Success != "Option Saved" {
		return fmt.Errorf("failed to set internal mail handler for %v: %v", domain, response.Result)
	}

	return nil
}

// VerifyEmailAccount (user) accepts the full email address as well as the password for the account. If the credentials aren't correct, an error will be returned.
func (c *UserContext) VerifyEmailAccount(address string, password string) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("email", address)
	body.Set("passwd", password)

	if _, err := c.makeRequestOld(http.MethodPost, "API_EMAIL_AUTH", body, &response); err != nil {
		return err
	}

	if response.Success != "Password Correct" {
		return errors.New("credentials incorrect")
	}

	return nil
}
