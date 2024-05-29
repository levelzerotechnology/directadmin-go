package directadmin

import (
	"fmt"
	"net/http"
	"net/url"

	"github.com/spf13/cast"
)

type PHPVersion struct {
	ID       string `json:"value"`
	Selected bool   `json:"selected"`
	Text     string `json:"text"`
	Version  string `json:"version"`
}

// GetPHPVersions (user) returns an array of the available PHP versions
func (c *UserContext) GetPHPVersions(domainName string) ([]*PHPVersion, error) {
	var rawPHPVersions struct {
		PHPSelect map[string]struct {
			Selected string `json:"selected"`
			Text     string `json:"text"`
			Value    string `json:"value"`
		} `json:"php1_select"`
		PHPVersion1 string `json:"php1_ver"`
		PHPVersion2 string `json:"php2_ver"`
		PHPVersion3 string `json:"php3_ver"`
		PHPVersion4 string `json:"php4_ver"`
	}

	if _, err := c.makeRequestOld(http.MethodGet, "API_ADDITIONAL_DOMAINS?domain="+domainName+"&action=view", nil, &rawPHPVersions); err != nil {
		return nil, err
	}

	versionMap := map[string]string{
		"0": rawPHPVersions.PHPVersion1,
		"1": rawPHPVersions.PHPVersion2,
		"2": rawPHPVersions.PHPVersion3,
		"3": rawPHPVersions.PHPVersion4,
	}

	versions := make([]*PHPVersion, 0, len(rawPHPVersions.PHPSelect))

	for index, rawVersion := range rawPHPVersions.PHPSelect {
		versions = append(versions, &PHPVersion{
			ID:       cast.ToString(cast.ToInt(index) + 1), // TODO: refactor this
			Text:     rawVersion.Text,
			Selected: rawVersion.Selected == "yes",
			Version:  versionMap[index],
		})
	}

	return versions, nil
}

// SetPHPVersion (user) sets the PHP version for the given domain to the given version ID
func (c *UserContext) SetPHPVersion(domain string, versionID string) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("action", "php_selector")
	body.Set("domain", domain)
	body.Set("php1_select", versionID)
	body.Set("save", "yes")

	if _, err := c.makeRequestOld(http.MethodPost, "API_DOMAIN", body, &response); err != nil {
		return err
	}

	if response.Success != "PHP versions saved" {
		return fmt.Errorf("failed to set PHP version: %v", response.Result)
	}

	return nil
}
