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
		AutoUpgradeThemes  bool   `url:"auto_upgrade_themes"`
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

// SoftaculousCreateLoginURL (user) returns a one-time URL that logs directly into the given installation's admin panel.
//
// Docs: https://www.softaculous.com/docs/api/remote-api/#auto-sign-on
func (c *UserContext) SoftaculousCreateLoginURL(installID string) (string, error) {
	var response struct {
		Error map[string]string `json:"error"`
		URL   string            `json:"sign_on_url"`
	}

	if installID == "" {
		return "", errors.New("missing install id")
	}

	if err := c.CreateSession(); err != nil {
		return "", fmt.Errorf("failed to create user session: %w", err)
	}

	if _, err := c.makeRequestOld(http.MethodPost, "PLUGINS/softaculous/index.raw?act=sign_on&insid="+installID+"&api=json", nil, &response); err != nil {
		return "", err
	}

	if len(response.Error) > 0 {
		return "", fmt.Errorf("failed to get login url: %v", response.Error)
	}

	return response.URL, nil
}

// SoftaculousInstallScript (user) calls Softaculous's install script API endpoint.
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

// SoftaculousListInstallations (user) lists all installations accessible to the authenticated user.
//
// Docs: https://www.softaculous.com/docs/api/remote-api/#list-installed-script
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

// SoftaculousUninstallScript (user) calls Softaculous's install script API endpoint.
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

// Core auto-upgrade levels for an existing installation, sent as auto_upgrade_core.
const (
	SoftaculousAutoUpgradeOff   = "0"
	SoftaculousAutoUpgradeMajor = "1"
	SoftaculousAutoUpgradeMinor = "2"
)

// SoftaculousInstallationSettings is the settings block returned when an installation is
// saved.
//
// The boolean-ish fields arrive inconsistently: an unset flag comes back as an empty
// string while a set one comes back as the number 1 or the boolean true, so they are held
// as any and read through the accessors rather than decoded straight into bool.
type SoftaculousInstallationSettings struct {
	ID       string `json:"insid"`
	ScriptID int    `json:"sid"`
	Domain   string `json:"softdomain"`
	Path     string `json:"softpath"`

	DisableWPCron      any `json:"disable_wp_cron"`
	AutoUpgradePlugins any `json:"auto_upgrade_plugins"`
	AutoUpgradeThemes  any `json:"auto_upgrade_themes"`

	// AutoUpgrade is the core upgrade level. Note the asymmetry: it is written as
	// auto_upgrade_core and read back as eu_auto_upgrade. Writing eu_auto_upgrade is
	// accepted and silently ignored.
	AutoUpgrade any `json:"eu_auto_upgrade"`

	// Set by Softaculous once wp-cron is disabled and it has installed a system cron in
	// its place. Empty while WordPress is still running its own pseudo-cron, so it doubles
	// as proof the change took.
	CronCommand string `json:"cron_command"`
	CronMinute  string `json:"cron_min"`
	CronHour    string `json:"cron_hour"`
}

// WPCronDisabled reports whether WordPress's own pseudo-cron is turned off, which is what
// makes Softaculous install a real system cron in its place.
func (s SoftaculousInstallationSettings) WPCronDisabled() bool {
	return cast.ToBool(s.DisableWPCron)
}

// PluginsAutoUpgrade reports whether plugin auto-upgrades are on.
func (s SoftaculousInstallationSettings) PluginsAutoUpgrade() bool {
	return cast.ToBool(s.AutoUpgradePlugins)
}

// ThemesAutoUpgrade reports whether theme auto-upgrades are on.
func (s SoftaculousInstallationSettings) ThemesAutoUpgrade() bool {
	return cast.ToBool(s.AutoUpgradeThemes)
}

// AutoUpgradeLevel reports the core upgrade level as one of the SoftaculousAutoUpgrade
// constants.
func (s SoftaculousInstallationSettings) AutoUpgradeLevel() string {
	// Off arrives as an empty string or false rather than as "0".
	if level := cast.ToString(s.AutoUpgrade); level != "" && level != "false" {
		return level
	}

	return SoftaculousAutoUpgradeOff
}

// SystemCronInstalled reports whether Softaculous has put a real cron in place of
// WordPress's pseudo-cron.
func (s SoftaculousInstallationSettings) SystemCronInstalled() bool {
	return s.CronCommand != ""
}

// SoftaculousInstallationDefaults are the settings a managed installation should carry.
type SoftaculousInstallationDefaults struct {
	// AutoUpgradeCore is one of the SoftaculousAutoUpgrade constants. Empty leaves it
	// alone.
	AutoUpgradeCore string

	AutoUpgradePlugins bool
	AutoUpgradeThemes  bool

	// DisableWPCron swaps WordPress's pseudo-cron for a system cron Softaculous installs,
	// so scheduled work stops depending on a visitor arriving and stops costing a PHP
	// request on every page view.
	DisableWPCron bool
}

// softaculousSave posts to the installation endpoint and returns the settings it reports.
//
// Only the supplied fields are written; anything omitted keeps its current value, so a
// save with nothing set is a safe way to read the current settings.
func (c *UserContext) softaculousSave(installID string, scriptName string, fields url.Values) (*SoftaculousInstallationSettings, error) {
	if installID == "" {
		return nil, errors.New("missing installation id")
	}

	if scriptName == "" {
		return nil, errors.New("missing script name")
	}

	var response struct {
		Done     any                             `json:"done"`
		Error    map[string]string               `json:"error"`
		Settings SoftaculousInstallationSettings `json:"settings"`
	}

	body := url.Values{}
	for key, values := range fields {
		for _, value := range values {
			body.Add(key, value)
		}
	}

	body.Set("insid", installID)
	body.Set("save", "1")

	// Softaculous requires a genuine session ID.
	if err := c.CreateSession(); err != nil {
		return nil, fmt.Errorf("failed to create user session: %w", err)
	}

	endpoint := "PLUGINS/softaculous/index.raw?act=" + url.QueryEscape(strings.ToLower(scriptName)) + "&api=json"

	if _, err := c.makeRequestOld(http.MethodPost, endpoint, body, &response); err != nil {
		return nil, err
	}

	if len(response.Error) > 0 {
		return nil, fmt.Errorf("softaculous rejected the save: %v", response.Error)
	}

	// done is only present on a save that had something to change. A read, which is a save
	// with no fields, omits it, so requiring it would fail every read.
	if len(fields) > 0 && !cast.ToBool(response.Done) {
		return nil, errors.New("softaculous did not report success")
	}

	return &response.Settings, nil
}

// SoftaculousReadSettings (user) returns an installation's current settings.
//
// Softaculous offers no read-only view of these fields, so this is a save with nothing to
// change. Omitted fields keep their values, which is what makes that safe.
func (c *UserContext) SoftaculousReadSettings(installID string, scriptName string) (*SoftaculousInstallationSettings, error) {
	return c.softaculousSave(installID, scriptName, nil)
}

// SoftaculousApplyDefaults (user) writes an installation's upgrade and cron settings.
//
// All of them go in one call. The response describes the installation as it stood
// beforehand, so it is not proof the write landed; call SoftaculousReadSettings after if
// confirmation is needed.
func (c *UserContext) SoftaculousApplyDefaults(installID string, scriptName string, defaults SoftaculousInstallationDefaults) (*SoftaculousInstallationSettings, error) {
	fields := url.Values{}

	if defaults.AutoUpgradeCore != "" {
		fields.Set("auto_upgrade_core", defaults.AutoUpgradeCore)
	}

	// Written whether on or off, otherwise a setting could never be turned back off.
	fields.Set("auto_upgrade_plugins", boolField(defaults.AutoUpgradePlugins))
	fields.Set("auto_upgrade_themes", boolField(defaults.AutoUpgradeThemes))
	fields.Set("disable_wp_cron", boolField(defaults.DisableWPCron))

	return c.softaculousSave(installID, scriptName, fields)
}

func boolField(on bool) string {
	if on {
		return "1"
	}

	return "0"
}
