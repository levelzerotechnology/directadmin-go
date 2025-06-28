package directadmin

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
)

const (
	PHPSelectorExtensionStateBuildIn  = "build-in"
	PHPSelectorExtensionStateDisabled = "disabled"
	PHPSelectorExtensionStateEnabled  = "enabled"
)

type (
	// PHPSelectorExtension represents a PHP extension configuration.
	PHPSelectorExtension struct {
		Description string `json:"description"`
		Name        string `json:"name"`
		State       string `json:"state"`
	}

	PHPSelectorList struct {
		AvailableVersions  []*PHPSelectorVersion `json:"available_versions"`
		DefaultVersion     string                `json:"default_version"`
		DomainsTabIsHidden bool                  `json:"domains_tab_is_hidden"`
		ExtensionsIsHidden bool                  `json:"extensions_is_hidden"`
		Result             string                `json:"result"`
		SelectedVersion    string                `json:"selected_version"`
		SelectorEnabled    bool                  `json:"selector_enabled"`
		Timestamp          float64               `json:"timestamp"`
	}

	// PHPSelectorOption represents a PHP configuration option.
	PHPSelectorOption struct {
		Comment string `json:"comment"`
		Default string `json:"default"`
		Range   string `json:"range,omitempty"`
		Type    string `json:"type"`
	}

	// PHPSelectorVersion represents a single PHP version configuration.
	PHPSelectorVersion struct {
		Extensions   []*PHPSelectorExtension       `json:"extensions"`
		NameModifier string                        `json:"name_modifier"`
		Options      map[string]*PHPSelectorOption `json:"options"`
		Status       string                        `json:"status"`
		Version      string                        `json:"version"`
	}
)

// GetVersion retrieves the given PHP version with its extensions.
func (l *PHPSelectorList) GetVersion(version string) (*PHPSelectorVersion, error) {
	for _, availableVersion := range l.AvailableVersions {
		if availableVersion.Version == version {
			return availableVersion, nil
		}
	}

	return nil, fmt.Errorf("failed to find version: %s", version)
}

// GetEnabledExtensions returns a slice of enabled and built-in extensions for the given PHP version.
func (v *PHPSelectorVersion) GetEnabledExtensions() []string {
	var enabledExtensions []string

	for _, extension := range v.Extensions {
		if extension.State == PHPSelectorExtensionStateBuildIn || extension.State == PHPSelectorExtensionStateEnabled {
			enabledExtensions = append(enabledExtensions, extension.Name)
		}
	}

	return enabledExtensions
}

// PHPSelectorDisableExtension disables the given extension for the given PHP version if it is not already disabled.
func (c *UserContext) PHPSelectorDisableExtension(version string, extension string) error {
	selectedVersion, err := c.PHPSelectorGetVersion(version)
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	enabledExtensions := selectedVersion.GetEnabledExtensions()
	setExtensions := make([]string, 0, len(enabledExtensions))

	for _, enabledExtension := range enabledExtensions {
		if enabledExtension != extension {
			setExtensions = append(setExtensions, enabledExtension)
		}
	}

	// Extension is already disabled or doesn't exist..
	if len(setExtensions) == len(enabledExtensions) {
		return nil
	}

	return c.PHPSelectorSetExtensions(version, setExtensions...)
}

// PHPSelectorEnableExtension enables the given extension for the given PHP version if it is not already enabled.
func (c *UserContext) PHPSelectorEnableExtension(version string, extension string) error {
	selectedVersion, err := c.PHPSelectorGetVersion(version)
	if err != nil {
		return fmt.Errorf("failed to get version: %w", err)
	}

	enabledExtensions := selectedVersion.GetEnabledExtensions()

	for _, enabledExtension := range enabledExtensions {
		// No need to continue as the extension is already enabled.
		if enabledExtension == extension {
			return nil
		}
	}

	enabledExtensions = append(enabledExtensions, extension)

	return c.PHPSelectorSetExtensions(version, enabledExtensions...)
}

// PHPSelectorGetDefaultVersion retrieves the server's default PHP version with its extensions.
func (c *UserContext) PHPSelectorGetDefaultVersion() (*PHPSelectorVersion, error) {
	versions, err := c.PHPSelectorListVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to list PHP versions: %w", err)
	}

	for _, availableVersion := range versions.AvailableVersions {
		if availableVersion.Version == versions.DefaultVersion {
			return availableVersion, nil
		}
	}

	return nil, fmt.Errorf("failed to find default version: %s", versions.DefaultVersion)
}

// PHPSelectorGetSelectedVersion retrieves the selected PHP version with its extensions.
func (c *UserContext) PHPSelectorGetSelectedVersion() (*PHPSelectorVersion, error) {
	versions, err := c.PHPSelectorListVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to list PHP versions: %w", err)
	}

	for _, availableVersion := range versions.AvailableVersions {
		if availableVersion.Version == versions.SelectedVersion {
			return availableVersion, nil
		}
	}

	return nil, fmt.Errorf("failed to find selected version: %s", versions.SelectedVersion)
}

// PHPSelectorGetVersion retrieves the given PHP version with its extensions.
func (c *UserContext) PHPSelectorGetVersion(version string) (*PHPSelectorVersion, error) {
	versions, err := c.PHPSelectorListVersions()
	if err != nil {
		return nil, fmt.Errorf("failed to list PHP versions: %w", err)
	}

	return versions.GetVersion(version)
}

// PHPSelectorListVersions lists all PHP versions accessible to the authenticated user.
func (c *UserContext) PHPSelectorListVersions() (*PHPSelectorList, error) {
	if err := c.CreateSession(); err != nil {
		return nil, fmt.Errorf("failed to create user session: %w", err)
	}

	csrfToken, err := c.phpSelectorCreateCSRFToken()
	if err != nil {
		return nil, fmt.Errorf("failed to create CSRF token: %w", err)
	}

	body := url.Values{}
	body.Set("command", "cloudlinux-selector")
	body.Set("csrftoken", csrfToken)
	body.Set("method", "get")
	body.Set("params[interpreter]", "php")

	var resp PHPSelectorList

	if _, err = c.makeRequestOld(http.MethodPost, "PLUGINS/phpselector/index.raw?c=send-request", body, &resp); err != nil {
		return nil, err
	}

	return &resp, nil
}

// PHPSelectorSetExtensions sets the given extensions for the given PHP version.
func (c *UserContext) PHPSelectorSetExtensions(version string, extensions ...string) error {
	if err := c.CreateSession(); err != nil {
		return fmt.Errorf("failed to create user session: %w", err)
	}

	csrfToken, err := c.phpSelectorCreateCSRFToken()
	if err != nil {
		return fmt.Errorf("failed to create CSRF token: %w", err)
	}

	formattedExtensions := make(map[string]string, len(extensions))
	for _, extension := range extensions {
		formattedExtensions[extension] = "enabled"
	}

	jsonExtensions, err := json.Marshal(formattedExtensions)
	if err != nil {
		return fmt.Errorf("failed to marshal extensions: %w", err)
	}

	body := url.Values{}
	body.Set("command", "cloudlinux-selector")
	body.Set("csrftoken", csrfToken)
	body.Set("method", "set")
	body.Set("params[extensions]", string(jsonExtensions))
	body.Set("params[interpreter]", "php")
	body.Set("params[version]", version)

	resp := struct {
		Result string `json:"result"`
	}{}

	if _, err = c.makeRequestOld(http.MethodPost, "PLUGINS/phpselector/index.raw?c=send-request", body, &resp); err != nil {
		return err
	}

	if resp.Result != "success" {
		return errors.New("failed to set extensions")
	}

	return nil
}

// phpSelectorCreateCSRFToken creates a CSRF token for the PHP Selector plugin if one doesn't already exist.
//
// This is a helper function used to retrieve the CSRF token for the PHP Selector plugin.
func (c *UserContext) phpSelectorCreateCSRFToken() (string, error) {
	const endpoint = "PLUGINS/phpselector/index.raw"

	csrfToken := c.getCSRFToken(c.getRequestURLOld(endpoint))
	if csrfToken != "" {
		return csrfToken, nil
	}

	// Retrieve CSRF token.
	if _, err := c.makeRequestOld(http.MethodGet, "PLUGINS/phpselector/index.raw?a=cookie", nil, nil); err != nil {
		return "", err
	}

	return c.getCSRFToken(c.getRequestURLOld(endpoint)), nil
}
