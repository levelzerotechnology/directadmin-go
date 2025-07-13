package directadmin

import (
	"net/http"
	"net/url"
	"strings"
)

type Session struct {
	AllowedCommands []string `json:"allowedCommands"`
	ConfigFeatures  struct {
		Auth2FA                         bool `json:"auth2FA"`
		BruteforceLogScanner            bool `json:"bruteforceLogScanner"`
		CGroup                          bool `json:"cgroup"`
		ClamAV                          bool `json:"clamav"`
		Composer                        bool `json:"composer"`
		DNSSEC                          int  `json:"dnssec"`
		Git                             bool `json:"git"`
		IMAPSync                        bool `json:"imapsync"`
		Inode                           bool `json:"inode"`
		IPv6                            bool `json:"IPv6"`
		Jail                            int  `json:"jail"`
		MXWithoutDNSControl             bool `json:"mxWithoutDNSControl"`
		NetdataSock                     bool `json:"netdataSock"`
		Nginx                           bool `json:"nginx"`
		NginxProxy                      bool `json:"nginxProxy"`
		NginxTemplates                  bool `json:"nginxTemplates"`
		OneClickPMALogin                bool `json:"oneClickPMALogin"`
		PHPMyAdmin                      bool `json:"phpmyadmin"`
		Redis                           bool `json:"redis"`
		ResellerCustomizeSkinConfigJSON bool `json:"resellerCustomizeSkinConfigJSON"`
		Roundcube                       bool `json:"roundcube"`
		RspamdSock                      bool `json:"rspamdSock"`
		SecurityQuestions               bool `json:"securityQuestions"`
		SquirrelMail                    bool `json:"squirrelMail"`
		Unit                            bool `json:"unit"`
		Webmail                         bool `json:"webmail"`
		WordPress                       bool `json:"wordpress"`
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
		FTPSeparator                       string   `json:"ftpSeparator"`
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
	PHPMyAdminPublic        bool   `json:"phpmyadminPublic"`
	RealUsername            string `json:"realUsername"`
	SelectedDomain          string `json:"selectedDomain"`
	SessionID               string `json:"sessionID"`
	TicketsEnabled          bool   `json:"ticketsEnabled"`
}

// CreateSession (user) creates a session for the provided credentials if one does not already exist.
func (c *UserContext) CreateSession() error {
	// Avoid creating a session if we already have one.
	apiCookies := c.cookieJar.Cookies(c.api.parsedURL)
	for _, cookie := range apiCookies {
		if cookie.Name == "session" {
			return nil
		}
	}

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

	// If we're a reseller logged in as a user, change the username to the reseller.
	if c.GetMyUsername() != c.credentials.username {
		request.Username = strings.Split(c.credentials.username, "|")[0]
	}

	if _, err := c.makeRequestNew(http.MethodPost, "login", request, &response); err != nil {
		return err
	}

	c.cookieJar.SetCookies(c.api.parsedURL, []*http.Cookie{{Name: "session", Value: response.SessionID}})

	// If we're a reseller logged in as a user, switch the session to the user.
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

// getCSRFToken retrieves the CSRF token from the cookie jar, for the provided URL.
func (c *UserContext) getCSRFToken(rawURL string) string {
	endpointURL, err := url.Parse(rawURL)
	if err != nil {
		return ""
	}

	cookies := c.cookieJar.Cookies(endpointURL)
	for _, cookie := range cookies {
		if cookie.Name == "csrftoken" {
			return cookie.Value
		}
	}

	return ""
}
