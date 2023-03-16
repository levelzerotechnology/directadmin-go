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

func (c *UserContext) CreateEmailAccount(emailAccount EmailAccount) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("domain", emailAccount.Domain)
	body.Set("user", emailAccount.Username)
	body.Set("passwd", emailAccount.Password)
	body.Set("passwd2", emailAccount.Password)
	body.Set("quota", cast.ToString(emailAccount.DiskQuota))
	body.Set("limit", cast.ToString(emailAccount.SendQuota))

	if _, err := c.api.makeRequest(http.MethodPost, "API_POP?action=create", c.credentials, body, &response); err != nil {
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

	if _, err := c.api.makeRequest(http.MethodPost, "API_POP?action=delete", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "E-Mail Accounts Deleted" {
		return fmt.Errorf("failed to delete email account: %v", response.Result)
	}

	return nil
}

// GetEmailAccounts (user) returns an array of email accounts belonging to the provided domain
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

	if _, err := c.api.makeRequest(http.MethodGet, "EMAIL_POP?bytes=yes&domain="+domain, c.credentials, nil, &rawEmailAccounts); err != nil {
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

			// this is necessary because DA either returns a string, or a map with the disk usage in a "usage" field depending on which usage endpoint we hit
			switch emailAccount.Sent.(type) {
			case map[string]any:
				emailAccountSent := emailAccount.Sent.(map[string]any)
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

func (c *UserContext) UpdateEmailAccount(emailAccount EmailAccount) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("domain", emailAccount.Domain)
	body.Set("user", emailAccount.Username)
	body.Set("passwd", emailAccount.Password)
	body.Set("passwd2", emailAccount.Password)
	body.Set("quota", cast.ToString(emailAccount.DiskQuota))
	body.Set("limit", cast.ToString(emailAccount.SendQuota))

	if _, err := c.api.makeRequest(http.MethodPost, "API_POP?action=modify", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "Account password changed" {
		return fmt.Errorf("failed to update email account: %v", response.Result)
	}

	return nil
}

// VerifyEmailAccount (user) accepts the full email address as well as the password for the account. If the credentials aren't correct, an error will be returned.
func (c *UserContext) VerifyEmailAccount(address string, password string) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("email", address)
	body.Set("passwd", password)

	if _, err := c.api.makeRequest(http.MethodPost, "API_EMAIL_AUTH", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "Password Correct" {
		return errors.New("credentials incorrect")
	}

	return nil
}
