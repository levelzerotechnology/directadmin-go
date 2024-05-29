package directadmin

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cast"
)

const (
	SoftaculousProtocolHTTP      = "1"
	SoftaculousProtocolHTTPWWW   = "2"
	SoftaculousProtocolHTTPS     = "3"
	SoftaculousProtocolHTTPSWWW  = "4"
	SoftaculousScriptIDWordPress = 26
)

type SoftaculousScript struct {
	AdminEmail         string `url:"admin_email"`
	AdminPassword      string `url:"admin_pass"`
	AdminUsername      string `url:"admin_username"`
	AutoUpgrade        bool   `url:"en_auto_upgrade"`
	AutoUpgradePlugins bool   `url:"auto_upgrade_plugins"`
	AutoUpgradeThemes  bool   `url:"auto_upgrade_plugins"`
	DatabaseName       string `url:"softdb"`
	DatabasePrefix     string `url:"dbprefix"` // optional
	Directory          string `url:"softdirectory"`
	Domain             string `url:"softdomain"`
	Language           string `url:"language"`
	NotifyOnInstall    bool   `url:"noemail"`
	NotifyOnUpdate     bool   `url:"disable_notify_update"`
	OverwriteExisting  bool   `url:"overwrite_existing"`
	Protocol           string `url:"softproto"`
	SiteDescription    string `url:"site_desc"`
	SiteName           string `json:"site_name"`
}

func (s *SoftaculousScript) Parse() (url.Values, error) {
	if err := s.Validate(); err != nil {
		return nil, err
	}

	values := url.Values{}
	values.Add("admin_email", s.AdminEmail)
	values.Add("admin_pass", s.AdminPassword)
	values.Add("admin_username", s.AdminUsername)

	if s.AutoUpgrade {
		values.Add("en_auto_upgrade", "1")
	}

	if s.AutoUpgradePlugins {
		values.Add("auto_upgrade_plugins", "1")
	}

	if s.AutoUpgradeThemes {
		values.Add("auto_upgrade_themes", "1")
	}

	if s.DatabaseName != "" {
		values.Add("softdb", s.DatabaseName)

		if s.DatabasePrefix != "" {
			values.Add("dbprefix", s.DatabasePrefix)
		}
	}

	if s.Directory != "" {
		values.Add("softdirectory", s.Directory)
	}

	values.Add("softdomain", s.Domain)
	values.Add("language", s.Language)

	if !s.NotifyOnInstall {
		values.Add("noemail", "1")
	}

	if !s.NotifyOnUpdate {
		values.Add("disable_notify_update", "1")
	}

	if s.OverwriteExisting {
		values.Add("overwrite_existing", "1")
	}

	values.Add("softproto", s.Protocol)
	values.Add("site_desc", s.SiteDescription)
	values.Add("site_name", s.SiteName)

	return values, nil
}

func (s *SoftaculousScript) Validate() error {
	if s.AdminEmail == "" {
		return errors.New("admin email is required")
	}

	if s.AdminPassword == "" {
		return errors.New("admin password is required")
	}

	if s.AdminUsername == "" {
		return errors.New("admin username is required")
	}

	if s.DatabasePrefix != "" && !strings.HasSuffix(s.DatabasePrefix, "_") {
		return errors.New("database prefix missing trailing underscore")
	}

	if s.Language == "" {
		return errors.New("language is required")
	}

	switch s.Protocol {
	case SoftaculousProtocolHTTP, SoftaculousProtocolHTTPWWW, SoftaculousProtocolHTTPS, SoftaculousProtocolHTTPSWWW:
	default:
		return errors.New("invalid protocol")
	}

	if s.SiteDescription == "" {
		return errors.New("site description is required")
	}

	if s.SiteName == "" {
		return errors.New("site name is required")
	}

	return nil
}

func SoftaculousScriptWithDefaults() *SoftaculousScript {
	return &SoftaculousScript{
		Language: "en",
		Protocol: SoftaculousProtocolHTTPS,
	}
}

// SoftaculousInstallScript calls Softaculous's install script API endpoint.
//
// Docs: https://www.softaculous.com/docs/api/remote-api/#install-a-script
func (c *UserContext) SoftaculousInstallScript(script *SoftaculousScript, scriptID int) error {
	response := struct {
		Error map[string]string `json:"error"`
	}{
		Error: make(map[string]string),
	}

	if scriptID == 0 {
		return errors.New("missing script id")
	}

	body, err := script.Parse()
	if err != nil {
		return err
	}

	body.Set("softsubmit", "1")

	// Softaculous requires a genuine session ID
	if c.sessionID == "" {
		if err = c.CreateSession(); err != nil {
			return fmt.Errorf("failed to create user session: %w", err)
		}
	}

	if _, err = c.makeRequestOld(http.MethodPost, "PLUGINS/softaculous/index.raw?act=software&soft="+cast.ToString(scriptID)+"&multi_ver=1&api=json", body, &response); err != nil {
		return err
	}

	if len(response.Error) > 0 {
		return fmt.Errorf("failed to install script: %v", response.Error)
	}

	return nil
}
