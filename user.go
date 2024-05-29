package directadmin

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"time"
)

type (
	User struct {
		Config UserConfig `json:"config"`
		Usage  UserUsage  `json:"usage"`
	}

	UserContext struct {
		api         *API
		credentials credentials
		sessionID   string
		User        User
	}

	UserConfig struct {
		AftpEnabled                    bool      `json:"aftpEnabled" yaml:"aftpEnabled"`
		ApiAccessWithPasswordEnabled   bool      `json:"apiAccessWithPasswordEnabled" yaml:"apiAccessWithPasswordEnabled"`
		CatchAllEnabled                bool      `json:"catchAllEnabled" yaml:"catchAllEnabled"`
		CgiEnabled                     bool      `json:"cgiEnabled" yaml:"cgiEnabled"`
		ClamAvEnabled                  bool      `json:"clamAvEnabled" yaml:"clamAvEnabled"`
		Created                        time.Time `json:"created" yaml:"created"`
		Creator                        string    `json:"creator" yaml:"creator"`
		CronEnabled                    bool      `json:"cronEnabled" yaml:"cronEnabled"`
		Domain                         string    `json:"domain" yaml:"domain"`
		DNSEnabled                     bool      `json:"dnsEnabled" yaml:"dnsEnabled"`
		Email                          string    `json:"email" yaml:"email"`
		GitEnabled                     bool      `json:"gitEnabled" yaml:"gitEnabled"`
		IpAddresses                    []string  `json:"ipAddresses" yaml:"ipAddresses"`
		JailEnabled                    bool      `json:"jailEnabled" yaml:"jailEnabled"`
		Language                       string    `json:"language" yaml:"language"`
		LoginKeysEnabled               bool      `json:"loginKeysEnabled" yaml:"loginKeysEnabled"`
		NginxEnabled                   bool      `json:"nginxEnabled" yaml:"nginxEnabled"`
		NotifyOnQuestionFailures       bool      `json:"notifyOnQuestionFailures" yaml:"notifyOnQuestionFailures"`
		NotifyOnTwoFactorFailures      bool      `json:"notifyOnTwoFactorFailures" yaml:"notifyOnTwoFactorFailures"`
		Ns1                            string    `json:"ns1" yaml:"ns1"`
		Ns2                            string    `json:"ns2" yaml:"ns2"`
		Package                        string    `json:"package" yaml:"package"`
		PhpEnabled                     bool      `json:"phpEnabled" yaml:"phpEnabled"`
		RedisEnabled                   bool      `json:"redisEnabled" yaml:"redisEnabled"`
		SecurityQuestionsEnabled       bool      `json:"securityQuestionsEnabled" yaml:"securityQuestionsEnabled"`
		Skin                           string    `json:"skin" yaml:"skin"`
		SpamEnabled                    bool      `json:"spamEnabled" yaml:"spamEnabled"`
		SshEnabled                     bool      `json:"sshEnabled" yaml:"sshEnabled"`
		SslEnabled                     bool      `json:"sslEnabled" yaml:"sslEnabled"`
		Suspended                      bool      `json:"suspended" yaml:"suspended"`
		SuspendAtLimitEnabled          bool      `json:"suspendAtLimitEnabled" yaml:"suspendAtLimitEnabled"`
		SysInfoEnabled                 bool      `json:"sysInfoEnabled" yaml:"sysInfoEnabled"`
		TwoFactorAuthenticationEnabled bool      `json:"twoFactorAuthenticationEnabled" yaml:"twoFactorAuthenticationEnabled"`
		Username                       string    `json:"username" yaml:"username"`
		UserType                       string    `json:"userType" yaml:"userType"`
		WordPressToolkitEnabled        bool      `json:"wordPressToolkitEnabled" yaml:"wordPressToolkitEnabled"`
	}

	UserUsage struct {
		BandwidthQuota        int           `json:"bandwidthQuota" yaml:"bandwidthQuota"`
		BandwidthUsage        int           `json:"bandwidthUsage" yaml:"bandwidthUsage"`
		DbQuota               int           `json:"dbQuota" yaml:"dbQuota"`
		DbUsage               int           `json:"dbUsage" yaml:"dbUsage"`
		DiskQuota             int           `json:"diskQuota" yaml:"diskQuota"`
		DiskUsage             int           `json:"diskUsage" yaml:"diskUsage"`
		Domains               []domainUsage `json:"domains" yaml:"domains"`
		DomainPointersQuota   int           `json:"domainPointersQuota" yaml:"domainPointersQuota"`
		DomainPointersUsage   int           `json:"domainPointersUsage" yaml:"domainPointersUsage"`
		DomainQuota           int           `json:"domainQuota" yaml:"domainQuota"`
		DomainUsage           int           `json:"domainUsage" yaml:"domainUsage"`
		EmailQuota            int           `json:"emailQuota" yaml:"emailQuota"`
		EmailUsage            int           `json:"emailUsage" yaml:"emailUsage"`
		EmailForwardersQuota  int           `json:"emailForwardersQuota" yaml:"emailForwardersQuota"`
		EmailForwardersUsage  int           `json:"emailForwardersUsage" yaml:"emailForwardersUsage"`
		EmailMailingListQuota int           `json:"emailMailingListQuota" yaml:"emailMailingListQuota"`
		EmailMailingListUsage int           `json:"emailMailingListUsage" yaml:"emailMailingListUsage"`
		FtpQuota              int           `json:"ftpQuota" yaml:"ftpQuota"`
		FtpUsage              int           `json:"ftpUsed" yaml:"ftpUsed"`
		InodeQuota            int           `json:"inodeQuota" yaml:"inodeQuota"`
		InodeUsage            int           `json:"inodeUsage" yaml:"inodeUsage"`
		SubdomainQuota        int           `json:"subdomainQuota" yaml:"subdomainQuota"`
		SubdomainUsage        int           `json:"subdomainUsage" yaml:"subdomainUsage"`
	}
)

// GetMyUserConfig (user) returns the session user's config
func (c *UserContext) GetMyUserConfig() (*UserConfig, error) {
	var config UserConfig
	var err error
	var rawConfig rawUserConfig

	if _, err = c.makeRequestOld(http.MethodGet, "API_SHOW_USER_CONFIG", nil, &rawConfig); err != nil {
		return nil, err
	}

	config, err = rawConfig.parse()
	if err != nil {
		return nil, err
	}

	return &config, nil
}

// GetMyUserUsage (user) returns the session user's usage
func (c *UserContext) GetMyUserUsage() (*UserUsage, error) {
	var rawUsage rawUserUsage
	var usage UserUsage

	if c.User.Config.Domain == "" {
		return nil, errors.New("user does not have a domain")
	}

	if _, err := c.makeRequestOld(http.MethodGet, "USER_STATS?bytes=yes&domain="+c.User.Config.Domain, nil, &rawUsage); err != nil {
		return nil, err
	}

	usage = rawUsage.parse()

	return &usage, nil
}

func (c *UserContext) addUsernamePrefix(check string) string {
	if !strings.HasPrefix(check, c.GetMyUsername()+"_") {
		check = c.GetMyUsername() + "_" + check
	}

	return check
}

func (c *UserContext) checkObjectExists(body url.Values) error {
	var response apiGenericResponse

	if _, err := c.makeRequestOld(http.MethodGet, "JSON_VALIDATE?"+body.Encode(), nil, &response); err != nil {
		return err
	}

	if response.Error != "" {
		return fmt.Errorf("object already exists: %v", response.Error)
	}

	return nil
}
