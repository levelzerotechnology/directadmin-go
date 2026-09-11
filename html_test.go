package directadmin

import "testing"

// Trimmed from a real Softaculous 6.3.9 settings page, keeping the shapes that matter: an
// unticked checkbox rendered with whitespace rather than an attribute, a ticked one, two
// selects so the scoping is exercised, the hidden colour fields, and the backup radio group.
const settingsFormFixture = `
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
<input type="checkbox" name="disable_suggestions" class="checked-style" id="disable_suggestions"  />
<input type="hidden" name="left_panel_bg" value="#333333" id="hidden_left_panel_bg">
<input type="hidden" name="left_panel_cathead" value="#F2F2F2" id="hidden_left_panel_cathead">
<input type="radio" name="default_backup_location" value="7" />
<input type="radio" name="default_backup_location" value="0" checked="checked" />
<input type="submit" class="btn sai-button" name="editsettings" value="Edit Settings" />
</form>`

func TestHTMLFormParsing(t *testing.T) {
	page := []byte(settingsFormFixture)

	if !htmlChecked(page, "off_backup_au") {
		t.Error(`off_backup_au carries checked="checked" and must read as ticked`)
	}

	// Unticked boxes render with whitespace where the attribute would be, and one carries a
	// class containing the word: both are what a looser match clears on the next save.
	for _, name := range []string{"arrange_domain", "user_defined_color", "disable_suggestions"} {
		if htmlChecked(page, name) {
			t.Errorf("%s has no checked attribute and must read as unticked", name)
		}
	}

	// Each select is read within its own block, or both answer with the language.
	if got := htmlSelected(page, "language"); got != "english" {
		t.Errorf("language = %q, want english", got)
	}

	if got := htmlSelected(page, "timezone"); got != "0" {
		t.Errorf("timezone = %q, want 0", got)
	}

	// The selected radio, not the first in the group.
	if got := htmlCheckedRadio(page, "default_backup_location"); got != "0" {
		t.Errorf("default_backup_location = %q, want 0", got)
	}

	if got := htmlInputValue(page, "left_panel_bg"); got != "#333333" {
		t.Errorf("left_panel_bg = %q, want #333333", got)
	}

	if got := htmlInputValue(page, "not_a_field"); got != "" {
		t.Errorf("missing field = %q, want empty", got)
	}
}

// data- prefixed lookalikes and hidden companion inputs are how a loose match gives the wrong
// answer for the right field.
func TestHTMLAttributeLookalikes(t *testing.T) {
	if htmlChecked([]byte(`<input type="checkbox" name="x" data-checked="1" />`), "x") {
		t.Error("data-checked is not the checked attribute")
	}

	both := []byte(`<input type="hidden" name="x" value="0" /><input type="checkbox" name="x" checked />`)
	if !htmlChecked(both, "x") {
		t.Error("a checkbox preceded by a hidden input of the same name must read as ticked")
	}

	selects := []byte(`<select name="language"><option data-selected="x" value="dutch">D</option>` +
		`<option value="english" selected="selected">E</option></select>`)
	if got := htmlSelected(selects, "language"); got != "english" {
		t.Errorf("language = %q, want english: data-selected is not the selected attribute", got)
	}

	if got := htmlInputValue([]byte(`<input name="c" data-value="junk" value="#333" />`), "c"); got != "#333" {
		t.Errorf("c = %q, want #333: data-value is not the value attribute", got)
	}

	placeholder := []byte(`<input type="hidden" name="c" value="" /><input name="c" value="#333" />`)
	if got := htmlInputValue(placeholder, "c"); got != "#333" {
		t.Errorf("c = %q, want #333: an empty placeholder must not shadow the real value", got)
	}
}
