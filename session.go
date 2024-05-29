package directadmin

import (
	"net/http"
	"strings"
)

type Session struct {
	AllowedCommands []string `json:"allowedCommands"`
	ConfigFeatures  struct {
		Auth2FA                         bool `json:"auth2FA"`
		BruteforceLogScanner            bool `json:"bruteforceLogScanner"`
		Cgroup                          bool `json:"cgroup"`
		Clamav                          bool `json:"clamav"`
		Composer                        bool `json:"composer"`
		Dnssec                          int  `json:"dnssec"`
		Git                             bool `json:"git"`
		Imapsync                        bool `json:"imapsync"`
		Inode                           bool `json:"inode"`
		Ipv6                            bool `json:"ipv6"`
		Jail                            int  `json:"jail"`
		MxWithoutDNSControl             bool `json:"mxWithoutDNSControl"`
		NetdataSock                     bool `json:"netdataSock"`
		Nginx                           bool `json:"nginx"`
		NginxProxy                      bool `json:"nginxProxy"`
		NginxTemplates                  bool `json:"nginxTemplates"`
		OneClickPMALogin                bool `json:"oneClickPMALogin"`
		Phpmyadmin                      bool `json:"phpmyadmin"`
		Redis                           bool `json:"redis"`
		ResellerCustomizeSkinConfigJson bool `json:"resellerCustomizeSkinConfigJson"`
		Roundcube                       bool `json:"roundcube"`
		RspamdSock                      bool `json:"rspamdSock"`
		SecurityQuestions               bool `json:"securityQuestions"`
		SquirrelMail                    bool `json:"squirrelMail"`
		Unit                            bool `json:"unit"`
		Webmail                         bool `json:"webmail"`
		Wordpress                       bool `json:"wordpress"`
	} `json:"configFeatures"`
	CustomDomainItems []struct {
		Checked     bool   `json:"checked"`
		Default     string `json:"default"`
		Description string `json:"description"`
		Hidden      bool   `json:"hidden"`
		Label       string `json:"label"`
		Name        string `json:"name"`
		Options     []struct {
			Text  string `json:"text"`
			Value string `json:"value"`
		} `json:"options"`
		ReadOnly bool   `json:"readOnly"`
		Type     string `json:"type"`
	} `json:"customDomainItems"`
	CustombuildOptions struct {
		ModSecurity bool `json:"modSecurity"`
	} `json:"custombuildOptions"`
	Demo              bool `json:"demo"`
	DirectadminConfig struct {
		AllowForwarderPipe                 bool     `json:"allowForwarderPipe"`
		FtpSeparator                       string   `json:"ftpSeparator"`
		HomeOverrides                      []string `json:"homeOverrides"`
		LoginKeys                          bool     `json:"loginKeys"`
		MaxFilesizeBytes                   int      `json:"maxFilesizeBytes"`
		ResellerWarningBandwidthPercentage int      `json:"resellerWarningBandwidthPercentage"`
		ShowPointersInList                 int      `json:"showPointersInList"`
		TableDefaultIPP                    int      `json:"tableDefaultIPP"`
		UserWarningBandwidthPercentage     int      `json:"userWarningBandwidthPercentage"`
		UserWarningInodePercentage         int      `json:"userWarningInodePercentage"`
		UserWarningQuotaPercentage         int      `json:"userWarningQuotaPercentage"`
		WebappsSSL                         bool     `json:"webappsSSL"`
		WebmailHideLinks                   bool     `json:"webmailHideLinks"`
		WebmailLink                        string   `json:"webmailLink"`
	} `json:"directadminConfig"`
	EffectiveRole           string `json:"effectiveRole"`
	EffectiveUsername       string `json:"effectiveUsername"`
	HavePluginHooksAdmin    bool   `json:"havePluginHooksAdmin"`
	HavePluginHooksReseller bool   `json:"havePluginHooksReseller"`
	HavePluginHooksUser     bool   `json:"havePluginHooksUser"`
	HomeDir                 string `json:"homeDir"`
	LoginAsDNSControl       bool   `json:"loginAsDNSControl"`
	PhpmyadminPublic        bool   `json:"phpmyadminPublic"`
	RealUsername            string `json:"realUsername"`
	SelectedDomain          string `json:"selectedDomain"`
	SessionID               string `json:"sessionID"`
	TicketsEnabled          bool   `json:"ticketsEnabled"`
}

func (c *UserContext) CreateSession() error {
	response := struct {
		SessionID string `json:"sessionID"`
	}{}

	request := struct {
		Username string `json:"username"`
		Password string `json:"password"`
	}{
		c.credentials.username,
		c.credentials.passkey,
	}

	// if we're a reseller logged in as a user, change the username to the reseller
	if c.GetMyUsername() != c.credentials.username {
		request.Username = strings.Split(c.credentials.username, "|")[0]
	}

	if _, err := c.makeRequestNew(http.MethodPost, "login", request, &response); err != nil {
		return err
	}

	c.sessionID = response.SessionID

	// if we're a reseller logged in as a user, switch the session to the user
	if c.GetMyUsername() != c.credentials.username {
		switchRequest := struct {
			Username string `json:"username"`
		}{
			strings.Split(c.credentials.username, "|")[1],
		}

		if _, err := c.makeRequestNew(http.MethodPost, "session/login-as/switch", switchRequest, nil); err != nil {
			return err
		}
	}

	return nil
}

func (c *UserContext) GetSessionInfo() (*Session, error) {
	var session Session

	if _, err := c.makeRequestNew(http.MethodGet, "session", nil, &session); err != nil {
		return nil, err
	}

	return &session, nil
}
