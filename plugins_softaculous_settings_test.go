package directadmin

import "testing"

// The form is posted whole, so everything the page holds has to survive the round trip:
// dropping a field is how this silently resets someone's language or theme.
func TestSoftaculousSettingsFormPreservesThePage(t *testing.T) {
	form := softaculousSettingsForm([]byte(settingsFormFixture))

	for field, want := range map[string]string{
		"language":                "english",
		"timezone":                "0",
		"default_backup_location": "0",
		"left_panel_bg":           "#333333",
		"left_panel_cathead":      "#F2F2F2",
		"editsettings":            "Edit Settings",
	} {
		if got := form.Get(field); got != want {
			t.Errorf("%s = %q, want %q", field, got, want)
		}
	}

	// Unticked boxes stay absent, which is how the form expresses them.
	for _, name := range softaculousSettingsCheckboxes {
		if form.Has(name) {
			t.Errorf("%s was unticked and must not be sent", name)
		}
	}

	// The caller sets this one, so the rebuilt form must not carry it either way.
	if form.Has(softaculousBackupsOffField) {
		t.Error("the backup field is the caller's to set, not the form rebuilder's")
	}
}

// A page without the form, or without the one checkbox this drives, is refused rather than
// read: an absent field is indistinguishable from an unticked one once parsed.
func TestSoftaculousSettingsRejectsAPageItCannotDrive(t *testing.T) {
	if htmlHasTag([]byte(`<html><body>Please login</body></html>`), "form", softaculousSettingsFormName) {
		t.Error("a page with no form must not pass as the settings form")
	}

	if htmlHasTag([]byte(`<form name="editsettings"><input name="language" /></form>`), "input", softaculousBackupsOffField) {
		t.Error("a form without the backup checkbox must not pass as drivable")
	}

	if !htmlHasTag([]byte(settingsFormFixture), "form", softaculousSettingsFormName) ||
		!htmlHasTag([]byte(settingsFormFixture), "input", softaculousBackupsOffField) {
		t.Error("the real form must be recognised")
	}
}

// An absent field leaves a setting alone, an empty one overwrites it, so a field the page does
// not carry must be left out of the save entirely.
func TestSoftaculousSettingsFormOmitsFieldsThePageLacks(t *testing.T) {
	// A select whose default option carries no selected attribute, and no backup location group
	// at all, which is how Softaculous renders an account with no remote locations.
	page := []byte(`<form name="editsettings">
	<select name="language"><option value="english">English</option></select>
	<select name="timezone"><option value="0" selected>Default</option></select>
	<input type="checkbox" name="off_backup_au" />
	<input type="submit" name="editsettings" value="Instellingen Bewerken" />
	</form>`)

	form := softaculousSettingsForm(page)

	for _, field := range []string{"language", "default_backup_location"} {
		if form.Has(field) {
			t.Errorf("%s is not on the page and must not be sent as %q", field, form.Get(field))
		}
	}

	if got := form.Get("timezone"); got != "0" {
		t.Errorf("timezone = %q, want 0: a field the page does carry is still sent", got)
	}

	if got := form.Get(softaculousSettingsFormName); got != "Instellingen Bewerken" {
		t.Errorf("submit = %q, want the page's own localized caption", got)
	}
}
