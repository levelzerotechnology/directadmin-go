package directadmin

import (
	"encoding/json"
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

type (
	SoftaculousInstallation struct {
		ID                string `json:"insid"`
		ScriptID          int    `json:"sid"`
		Ver               string `json:"ver"`
		ITime             int    `json:"itime"`
		Path              string `json:"softpath"`
		URL               string `json:"softurl"`
		Domain            string `json:"softdomain"`
		FileIndex         any    `json:"fileindex"` // Sometimes a string slice, other times a map.
		SiteName          string `json:"site_name"`
		SoftDB            string `json:"softdb"`
		SoftDBuser        string `json:"softdbuser"`
		SoftDBhost        string `json:"softdbhost"`
		SoftDBpass        string `json:"softdbpass"`
		DBCreated         bool   `json:"dbcreated"`
		DBPrefix          string `json:"dbprefix"`
		ImportSrc         string `json:"import_src"`
		DisplaySoftDBPass string `json:"display_softdbpass"`
		ScriptName        string `json:"script_name"`
	}

	SoftaculousScript struct {
		AdminEmail         string `url:"admin_email"`
		AdminPassword      string `url:"admin_pass"`
		AdminUsername      string `url:"admin_username"`
		AutoUpgrade        bool   `url:"en_auto_upgrade"`
		AutoUpgradePlugins bool   `url:"auto_upgrade_plugins"`
		AutoUpgradeThemes  bool   `url:"auto_upgrade_plugins"`
		DatabaseName       string `url:"softdb"`
		DatabasePrefix     string `url:"dbprefix"` // Optional.
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
)

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

	// Softaculous requires a genuine session ID.
	if err = c.CreateSession(); err != nil {
		return fmt.Errorf("failed to create user session: %w", err)
	}

	if _, err = c.makeRequestOld(http.MethodPost, "PLUGINS/softaculous/index.raw?act=software&soft="+cast.ToString(scriptID)+"&multi_ver=1&api=json", body, &response); err != nil {
		return err
	}

	if len(response.Error) > 0 {
		return fmt.Errorf("failed to install script: %v", response.Error)
	}

	return nil
}

// SoftaculousListInstallations lists all installations accessible to the authenticated user.
func (c *UserContext) SoftaculousListInstallations() ([]*SoftaculousInstallation, error) {
	type rawResponse struct {
		Error         map[string]string `json:"error"`
		Installations json.RawMessage   `json:"installations"`
	}

	var raw rawResponse

	if err := c.CreateSession(); err != nil {
		return nil, fmt.Errorf("failed to create user session: %w", err)
	}

	if _, err := c.makeRequestOld(http.MethodPost, "PLUGINS/softaculous/index.raw?act=installations&api=json", nil, &raw); err != nil {
		return nil, err
	}

	if len(raw.Error) > 0 {
		return nil, fmt.Errorf("failed to list installations: %v", raw.Error)
	}

	// Try unmarshalling as a map first.
	var installationsMap map[string]map[string]*SoftaculousInstallation
	if err := json.Unmarshal(raw.Installations, &installationsMap); err == nil {
		var installations []*SoftaculousInstallation
		for _, userInstalls := range installationsMap {
			for _, install := range userInstalls {
				installations = append(installations, install)
			}
		}
		return installations, nil
	}

	// Fallback: Check if it's an empty array.
	var installationsArray []any
	if err := json.Unmarshal(raw.Installations, &installationsArray); err == nil && len(installationsArray) == 0 {
		return []*SoftaculousInstallation{}, nil
	}

	return nil, errors.New("unexpected format for installations field")
}

// SoftaculousUninstallScript calls Softaculous's install script API endpoint.
//
// Docs: https://www.softaculous.com/docs/api/remote-api/#remove-an-installed-script
func (c *UserContext) SoftaculousUninstallScript(installID string, deleteFiles bool, deleteDB bool) error {
	if installID == "" {
		return errors.New("missing install id")
	}

	response := struct {
		Error map[string]string `json:"error"`
	}{
		Error: make(map[string]string),
	}

	body := url.Values{}
	body.Set("noemail", "1")
	body.Set("removeins", "1")

	if deleteFiles {
		body.Set("remove_dir", "1")
		body.Set("remove_datadir", "1")
	}

	if deleteDB {
		body.Set("remove_db", "1")
		body.Set("remove_dbuser", "1")
	}

	// Softaculous requires a genuine session ID
	if err := c.CreateSession(); err != nil {
		return fmt.Errorf("failed to create user session: %w", err)
	}

	if _, err := c.makeRequestOld(http.MethodPost, "PLUGINS/softaculous/index.raw?act=remove&insid="+installID+"&api=json", body, &response); err != nil {
		return err
	}

	if len(response.Error) > 0 {
		return fmt.Errorf("failed to uninstall script: %v", response.Error)
	}

	return nil
}
