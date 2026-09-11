package directadmin

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strings"
)

// Softaculous's per-user settings form. Its fields are absent from the remote API, so they are
// read off the rendered page and written back as the form post the panel itself makes.
const softaculousSettingsPath = "/CMD_PLUGINS/softaculous/index.raw?act=settings"

// The form's name, which doubles as the submit button's: posting it is what makes Softaculous
// treat the request as a save.
const softaculousSettingsFormName = "editsettings"

// The "Disable Backups before Auto Upgrade" checkbox. The name reads backwards: set means the
// backups are off.
const softaculousBackupsOffField = "off_backup_au"

// The remaining fields, replayed on every save because an absent checkbox is how the form says
// "unchecked", so a post carrying one field would clear the rest. This is the form as
// Softaculous 6.3.9 renders it; a version that adds a field would have it dropped, and dropped
// means reset.
var (
	softaculousSettingsCheckboxes = []string{
		"arrange_domain",
		"user_defined_color",
		"disable_suggestions",
	}

	// Custom theme colours, which apply while user_defined_color is set.
	softaculousSettingsHidden = []string{
		"left_panel_bg",
		"left_panel_cathead_hover",
		"left_panel_cathead",
		"left_panel_scriptname",
	}
)

// SoftaculousUpgradeBackupsEnabled (user) reports whether Softaculous takes a backup before it
// auto upgrades an installation, its plugins or its themes.
func (c *UserContext) SoftaculousUpgradeBackupsEnabled() (bool, error) {
	page, err := c.softaculousSettingsPage(nil)
	if err != nil {
		return false, err
	}

	return !htmlChecked(page, softaculousBackupsOffField), nil
}

// SoftaculousSetUpgradeBackups (user) turns the pre-upgrade backup on or off, reporting whether
// it had to change anything. Off is worth having wherever something else already backs the
// account up, since a full copy before every upgrade fills a disk quota quickly.
//
// The whole form is read and posted back with only this field changed, so the account's
// language, timezone, backup location and theme are carried through untouched.
func (c *UserContext) SoftaculousSetUpgradeBackups(enabled bool) (bool, error) {
	page, err := c.softaculousSettingsPage(nil)
	if err != nil {
		return false, err
	}

	if htmlChecked(page, softaculousBackupsOffField) == !enabled {
		return false, nil
	}

	form := softaculousSettingsForm(page)
	if !enabled {
		form.Set(softaculousBackupsOffField, "on")
	}

	saved, err := c.softaculousSettingsPage(strings.NewReader(form.Encode()))
	if err != nil {
		return false, err
	}

	// Softaculous returns the page with a 200 whether or not it accepted the post, so the answer
	// is read back off the form it returns.
	if htmlChecked(saved, softaculousBackupsOffField) != !enabled {
		return false, errors.New("softaculous did not accept the settings save")
	}

	return true, nil
}

// softaculousSettingsPage fetches the settings form, or posts body to it and returns the page
// that comes back.
func (c *UserContext) softaculousSettingsPage(body io.Reader) ([]byte, error) {
	// Softaculous needs a real panel session rather than just the API credentials.
	if err := c.CreateSession(); err != nil {
		return nil, fmt.Errorf("failed to create user session: %w", err)
	}

	method := http.MethodGet
	if body != nil {
		method = http.MethodPost
	}

	req, err := c.newRequest(context.Background(), method, c.api.url+softaculousSettingsPath, body)
	if err != nil {
		return nil, err
	}

	if body != nil {
		req.Header.Set("Content-Type", "application/x-www-form-urlencoded")
	}

	status, page, err := c.makeRequest(req)
	if err != nil {
		return nil, fmt.Errorf("softaculous settings: %w", err)
	}

	if status != http.StatusOK {
		return nil, fmt.Errorf("softaculous settings: unexpected status code: %d", status)
	}

	// A login page, an error page and a relocated form all arrive with a 200, and every field
	// then reads as absent rather than missing. Both checks refuse a page this cannot drive,
	// rather than rebuilding an empty form and posting the account's settings away.
	if !htmlHasTag(page, "form", softaculousSettingsFormName) {
		return nil, errors.New("softaculous settings: response is not the settings form")
	}

	if !htmlHasTag(page, "input", softaculousBackupsOffField) {
		return nil, fmt.Errorf("softaculous settings: form has no %s field", softaculousBackupsOffField)
	}

	return page, nil
}

// softaculousSettingsForm rebuilds the settings form from the page as it currently stands, so
// that saving changes only the field the caller then overrides.
func softaculousSettingsForm(page []byte) url.Values {
	form := url.Values{}

	// Empty is skipped rather than sent: a select rendering its default with no selected
	// attribute, or an absent backup location group, would otherwise reset the real setting.
	for field, value := range map[string]string{
		"language":                htmlSelected(page, "language"),
		"timezone":                htmlSelected(page, "timezone"),
		"default_backup_location": htmlCheckedRadio(page, "default_backup_location"),
	} {
		if value != "" {
			form.Set(field, value)
		}
	}

	for _, box := range softaculousSettingsCheckboxes {
		if htmlChecked(page, box) {
			form.Set(box, "on")
		}
	}

	for _, hidden := range softaculousSettingsHidden {
		if value := htmlInputValue(page, hidden); value != "" {
			form.Set(hidden, value)
		}
	}

	// The submit caption is rendered in the account's own language, so it is read off the page.
	submit := htmlInputValue(page, softaculousSettingsFormName)
	if submit == "" {
		submit = "Edit Settings"
	}

	form.Set(softaculousSettingsFormName, submit)

	return form
}
