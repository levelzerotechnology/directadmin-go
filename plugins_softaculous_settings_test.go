package directadmin

import "testing"

// Trimmed from a real Softaculous 6.3.9 settings page, keeping the shapes that matter: an
// unticked checkbox rendered with trailing whitespace rather than an attribute, a ticked one
// carrying checked="checked", two selects so the scoping is exercised, the hidden colour
// fields, and the backup location radio.
const softaculousSettingsFixture = `
<form accept-charset="UTF-8" name="editsettings" method="post" action="" id="editsettings">
<select name="language" id="choose_lang" class="form-control">
	<option class="form-control" value="dutch" >Dutch</option>
	<option class="form-control" value="english" selected="selected">English</option>
	<option class="form-control" value="french" >French</option>
</select>
<select name="timezone" id="timezone" class="form-control">
	<option value="0" selected="selected" >Default</option>
	<option class="form-control" value="Europe/London"  >(GMT +01:00) Europe/London</option>
</select>
<input type="checkbox" name="arrange_domain" id="arrange_domain"  />
<input type="checkbox" name="off_backup_au" id="off_backup_au" checked="checked" />
<input type="checkbox" name="user_defined_color" id="user_defined_color"  onchange="change_div(this.id);" />
<input type="checkbox" name="disable_suggestions" id="disable_suggestions"  />
<input type="hidden" name="left_panel_bg" value="#333333" id="hidden_left_panel_bg">
<input type="hidden" name="left_panel_cathead" value="#F2F2F2" id="hidden_left_panel_cathead">
<input type="radio" name="default_backup_location" value="7" />
<input type="radio" name="default_backup_location" value="0" checked="checked" />
<input type="submit" class="btn sai-button" name="editsettings" value="Edit Settings" />
</form>`

func TestSoftaculousSettingsParsing(t *testing.T) {
	page := []byte(softaculousSettingsFixture)

	if !htmlChecked(page, softaculousBackupsOffField) {
		t.Error("off_backup_au carries checked=\"checked\" and must read as ticked")
	}

	// The unticked boxes render with whitespace where the attribute would be, which is the
	// case a looser match would get wrong and silently clear on the next save.
	for _, name := range softaculousSettingsCheckboxes {
		if htmlChecked(page, name) {
			t.Errorf("%s has no checked attribute and must read as unticked", name)
		}
	}

	// Each select must be read within its own block. Answering from the first selected option
	// in the document would give the language for both.
	if got := htmlSelected(page, "language"); got != "english" {
		t.Errorf("language = %q, want english", got)
	}

	if got := htmlSelected(page, "timezone"); got != "0" {
		t.Errorf("timezone = %q, want 0", got)
	}

	// The selected radio, not simply the first in the group.
	if got := htmlCheckedRadio(page, "default_backup_location"); got != "0" {
		t.Errorf("default_backup_location = %q, want 0", got)
	}

	if got := htmlInputValue(page, "left_panel_bg"); got != "#333333" {
		t.Errorf("left_panel_bg = %q, want #333333", got)
	}

	// A field that is not on the page reads as absent rather than as some other field.
	if got := htmlInputValue(page, "not_a_field"); got != "" {
		t.Errorf("missing field = %q, want empty", got)
	}
}

// The form is posted whole, so everything the page holds has to survive the round trip:
// dropping a field is how this silently resets someone's language or theme.
func TestSoftaculousSettingsFormPreservesThePage(t *testing.T) {
	form := softaculousSettingsForm([]byte(softaculousSettingsFixture))

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
