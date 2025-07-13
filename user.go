package directadmin

import (
	"fmt"
	"net/http"
	"net/url"
	"strings"
)

type (
	User struct {
		Config UserConfig `json:"config"`
		Usage  UserUsage  `json:"usage"`
	}

	UserContext struct {
		api         *API
		cookieJar   http.CookieJar
		credentials credentials
		User        User
	}

	UserConfig struct {
		AccountEnabled           bool     `json:"account"`
		AFTPEnabled              bool     `json:"aftp"`
		APIAllowPassword         bool     `json:"apiAllowPassword"`
		AutorespondersLimit      int      `json:"autorespondersLim"`
		BandwidthLimit           int      `json:"bandwidthLim"`
		CatchAllEnabled          bool     `json:"catchAll"`
		CGIEnabled               bool     `json:"cgi"`
		ClamAVEnabled            bool     `json:"clamav"`
		CPUQuota                 string   `json:"cpuQuota"`
		Creator                  string   `json:"creator"`
		CronEnabled              bool     `json:"cron"`
		DateCreated              string   `json:"dateCreated"`
		DNSControlEnabled        bool     `json:"dnsControl"`
		DocsRoot                 string   `json:"docsRoot"`
		Domain                   string   `json:"domain"`
		DomainPointersLimit      int      `json:"domainPointersLim"`
		Domains                  []string `json:"domains"`
		DomainsLimit             int      `json:"domainsLim"`
		Email                    string   `json:"email"`
		EmailAccountsLimit       int      `json:"emailAccountsLim"`
		EmailForwardersLimit     int      `json:"emailForwardersLim"`
		FeatureSets              []string `json:"featureSets"`
		FTPAccountsLimit         int      `json:"ftpAccountsLim"`
		GitEnabled               bool     `json:"git"`
		InodeLimit               int      `json:"inodeLim"`
		IOReadBandwidthMax       string   `json:"ioReadBandwidthMax"`
		IOReadIOPSMax            string   `json:"ioReadIOPSMax"`
		IOWriteBandwidthMax      string   `json:"ioWriteBandwidthMax"`
		IOWriteIOPSMax           string   `json:"ioWriteIOPSMax"`
		IP                       string   `json:"ip"`
		JailEnabled              bool     `json:"jail"`
		Language                 string   `json:"language"`
		LetsEncrypt              int      `json:"letsEncrypt"`
		LoginKeysEnabled         bool     `json:"loginKeys"`
		MailPartition            string   `json:"mailPartition"`
		MailingListsLimit        int      `json:"mailingListsLim"`
		MemoryHigh               string   `json:"memoryHigh"`
		MemoryMax                string   `json:"memoryMax"`
		MySQLConf                string   `json:"mySqlConf"`
		MySQLDatabasesLimit      int      `json:"mySqlDatabasesLim"`
		Name                     string   `json:"name"`
		NginxEnabled             bool     `json:"nginxUnit"`
		NS1                      string   `json:"ns1"`
		NS2                      string   `json:"ns2"`
		Package                  string   `json:"package"`
		PHPEnabled               bool     `json:"php"`
		PluginsBlacklist         []string `json:"pluginsBlacklist"`
		PluginsWhitelist         []string `json:"pluginsWhitelist"`
		QuotaLimit               int      `json:"quotaLim"`
		RedisEnabled             bool     `json:"redis"`
		SecurityQuestionsEnabled bool     `json:"securityQuestions"`
		Skin                     string   `json:"skin"`
		SpamEnabled              bool     `json:"spam"`
		SSHEnabled               bool     `json:"ssh"`
		SSLEnabled               bool     `json:"ssl"`
		SubdomainsLimit          int      `json:"subdomainsLim"`
		Suspended                bool     `json:"suspended"`
		SysInfoEnabled           bool     `json:"sysInfo"`
		TasksMax                 string   `json:"tasksMax"`
		TwoStepAuthEnabled       bool     `json:"twoStepAuth"`
		TwoStepAuthDescription   string   `json:"twoStepAuthDesc"`
		UserType                 string   `json:"userType"`
		Username                 string   `json:"username"`
		Users                    []string `json:"users"`
		UsersLimit               int      `json:"usersLim"`
		UsersManageDomains       int      `json:"usersManageDomains"`
		WordPressEnabled         bool     `json:"wordpress"`
		Zoom                     int      `json:"zoom"`
	}

	UserUsage struct {
		AutoResponders struct {
			Limit     int  `json:"limit"`
			Unlimited bool `json:"unlimited"`
			Usage     int  `json:"usage"`
		} `json:"autoresponders"`
		BandwidthBytes struct {
			Limit     int  `json:"limit"`
			Unlimited bool `json:"unlimited"`
			Usage     int  `json:"usage"`
		} `json:"bandwidthBytes"`
		DBQuotaBytes   int `json:"dbQuotaBytes"`
		DomainPointers struct {
			Limit     int  `json:"limit"`
			Unlimited bool `json:"unlimited"`
			Usage     int  `json:"usage"`
		} `json:"domainPointers"`
		Domains struct {
			Limit     int  `json:"limit"`
			Unlimited bool `json:"unlimited"`
			Usage     int  `json:"usage"`
		} `json:"domains"`
		EmailAccounts struct {
			Limit     int  `json:"limit"`
			Unlimited bool `json:"unlimited"`
			Usage     int  `json:"usage"`
		} `json:"emailAccounts"`
		EmailDeliveries         int `json:"emailDeliveries"`
		EmailDeliveriesIncoming int `json:"emailDeliveriesIncoming"`
		EmailDeliveriesOutgoing int `json:"emailDeliveriesOutgoing"`
		EmailForwarders         struct {
			Limit     int  `json:"limit"`
			Unlimited bool `json:"unlimited"`
			Usage     int  `json:"usage"`
		} `json:"emailForwarders"`
		EmailQuotaBytes int `json:"emailQuotaBytes"`
		FTPAccounts     struct {
			Limit     int  `json:"limit"`
			Unlimited bool `json:"unlimited"`
			Usage     int  `json:"usage"`
		} `json:"ftpAccounts"`
		Inode struct {
			Limit     int  `json:"limit"`
			Unlimited bool `json:"unlimited"`
			Usage     int  `json:"usage"`
		} `json:"inode"`
		MailingLists struct {
			Limit     int  `json:"limit"`
			Unlimited bool `json:"unlimited"`
			Usage     int  `json:"usage"`
		} `json:"mailingLists"`
		MySQLDatabases struct {
			Limit     int  `json:"limit"`
			Unlimited bool `json:"unlimited"`
			Usage     int  `json:"usage"`
		} `json:"mySqlDatabases"`
		OtherQuotaBytes int `json:"otherQuotaBytes"`
		QuotaBytes      struct {
			Limit     int  `json:"limit"`
			Unlimited bool `json:"unlimited"`
			Usage     int  `json:"usage"`
		} `json:"quotaBytes"`
		QuotaWithoutSystemBytes int `json:"quotaWithoutSystemBytes"`
		Subdomains              struct {
			Limit     int  `json:"limit"`
			Unlimited bool `json:"unlimited"`
			Usage     int  `json:"usage"`
		} `json:"subdomains"`
	}
)

// GetMyUserConfig (user) returns the session user's config.
func (c *UserContext) GetMyUserConfig() (*UserConfig, error) {
	var config UserConfig

	if _, err := c.makeRequestNew(http.MethodGet, "session/user-config", nil, &config); err != nil {
		return nil, err
	}

	return &config, nil
}

// GetMyUserUsage (user) returns the session user's usage.
func (c *UserContext) GetMyUserUsage() (*UserUsage, error) {
	var usage UserUsage

	if _, err := c.makeRequestNew(http.MethodGet, "session/user-usage", nil, &usage); err != nil {
		return nil, err
	}

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
