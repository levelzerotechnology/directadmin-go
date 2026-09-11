package directadmin

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
)

// softaculousSettingsPath is Softaculous's per-user settings form.
//
// Its fields are not exposed through the remote API the rest of this file's neighbours use,
// so they are read off the rendered page and written back as the form post the panel itself
// makes. Everything here exists because of that.
const softaculousSettingsPath = "/CMD_PLUGINS/softaculous/index.raw?act=settings"

// softaculousBackupsOffField is the "Disable Backups before Auto Upgrade" checkbox. Its name
// reads backwards: set means the backups are off.
const softaculousBackupsOffField = "off_backup_au"

// softaculousSettingsCheckboxes are the other checkboxes on the form. They are replayed on
// every save because an absent checkbox is how the form says "unchecked", so a post carrying
// one field would silently clear the rest.
var softaculousSettingsCheckboxes = []string{
	"arrange_domain",
	"user_defined_color",
	"disable_suggestions",
}

// softaculousSettingsHidden are the hidden fields holding the custom theme colours. They
// only apply while user_defined_color is set, but they are replayed for the same reason.
var softaculousSettingsHidden = []string{
	"left_panel_bg",
	"left_panel_cathead_hover",
	"left_panel_cathead",
	"left_panel_scriptname",
}

// SoftaculousUpgradeBackupsEnabled (user) reports whether Softaculous takes a backup before
// it auto upgrades an installation, its plugins or its themes.
func (c *UserContext) SoftaculousUpgradeBackupsEnabled() (bool, error) {
	page, err := c.softaculousSettingsPage(nil)
	if err != nil {
		return false, err
	}

	// The field records the negative, so an unticked box is backups on.
	return !htmlChecked(page, softaculousBackupsOffField), nil
}

// SoftaculousSetUpgradeBackups (user) turns the pre-upgrade backup on or off, reporting
// whether it had to change anything.
//
// Off is worth having wherever something else already backs the account up: Softaculous
// takes a full copy before every core, plugin and theme upgrade it performs, and on an
// account with auto upgrades enabled that is the fastest way to fill a disk quota.
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

	// Read the answer back off the response, which is the form as it now stands. Softaculous
	// returns the page with a 200 whether or not it accepted the post, so the status proves
	// nothing on its own.
	if htmlChecked(saved, softaculousBackupsOffField) != !enabled {
		return false, errors.New("softaculous did not accept the settings save")
	}

	return true, nil
}

// softaculousSettingsPage fetches the settings form, or posts body to it and returns the
// page that comes back.
func (c *UserContext) softaculousSettingsPage(body io.Reader) ([]byte, error) {
	// Softaculous needs a real panel session rather than just the API credentials, as its
	// installation endpoints do.
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

	// Set explicitly, because this is the panel's own HTML form rather than an API endpoint
	// and Softaculous reads it as one.
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

	return page, nil
}

// softaculousSettingsForm rebuilds the settings form from the page as it currently stands,
// so that saving changes only the field the caller then overrides.
func softaculousSettingsForm(page []byte) url.Values {
	form := url.Values{}

	form.Set("language", htmlSelected(page, "language"))
	form.Set("timezone", htmlSelected(page, "timezone"))
	form.Set("default_backup_location", htmlCheckedRadio(page, "default_backup_location"))

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

	// The submit button's own name and value, which is what makes Softaculous treat the post
	// as a save rather than another page load.
	form.Set("editsettings", "Edit Settings")

	return form
}

// Reading values back out of a rendered HTML form.
//
// Regular expressions rather than a parser, deliberately: the only form this is used on is
// the Softaculous settings page, and the alternative is an HTML parser dependency to read
// four fields off one page. The cost of being wrong is bounded, because every caller reads
// the field back after writing it, so a mis-parse surfaces as a refused save.

// namedTag matches an element of the given kind carrying name="<name>".
func namedTag(element, name string) *regexp.Regexp {
	return regexp.MustCompile(`(?is)<` + element + `[^>]*\bname="` + regexp.QuoteMeta(name) + `"[^>]*>`)
}

// htmlAttr returns an attribute's value from a single tag.
func htmlAttr(tag []byte, attr string) string {
	match := regexp.MustCompile(`(?is)\b` + regexp.QuoteMeta(attr) + `="([^"]*)"`).FindSubmatch(tag)
	if len(match) < 2 {
		return ""
	}

	return string(match[1])
}

// htmlChecked reports whether the named checkbox is ticked.
func htmlChecked(page []byte, name string) bool {
	tag := namedTag("input", name).Find(page)

	return tag != nil && bytes.Contains(bytes.ToLower(tag), []byte("checked"))
}

// htmlCheckedRadio returns the value of the selected radio in a named group, or "" when
// nothing in the group is selected.
func htmlCheckedRadio(page []byte, name string) string {
	for _, tag := range namedTag("input", name).FindAll(page, -1) {
		if bytes.Contains(bytes.ToLower(tag), []byte("checked")) {
			return htmlAttr(tag, "value")
		}
	}

	return ""
}

// htmlInputValue returns the value of the named input, which is how the hidden fields
// carrying a form's real state are read.
func htmlInputValue(page []byte, name string) string {
	tag := namedTag("input", name).Find(page)
	if tag == nil {
		return ""
	}

	return htmlAttr(tag, "value")
}

// selectedOption matches the option carrying a selected attribute.
var selectedOption = regexp.MustCompile(`(?is)<option[^>]*\bselected\b[^>]*>`)

// htmlSelected returns the value of the chosen option in the named select.
//
// Scoped to that select's own block first, since a page carrying several selects would
// otherwise answer with whichever option came first in the document.
func htmlSelected(page []byte, name string) string {
	block := regexp.MustCompile(`(?is)<select[^>]*\bname="` + regexp.QuoteMeta(name) + `".*?</select>`).Find(page)
	if block == nil {
		return ""
	}

	option := selectedOption.Find(block)
	if option == nil {
		return ""
	}

	return htmlAttr(option, "value")
}
