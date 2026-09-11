package directadmin

import (
	"regexp"
)

// Reading values off a rendered HTML form, for the panel settings with no API endpoint behind
// them. Regexes rather than an HTML parser dependency, deliberately. Known limit: an attribute
// value containing '>' ends a tag match early; no form read here carries one.

var (
	// Matched from the whitespace in front, so data-checked and the like are not mistaken for
	// the real attribute.
	checkedAttr  = regexp.MustCompile(`(?is)\schecked\s*(=|/?>|\s)`)
	selectedAttr = regexp.MustCompile(`(?is)\sselected\s*(=|/?>|\s)`)
	valueAttr    = regexp.MustCompile(`(?is)\svalue="([^"]*)"`)

	optionTag = regexp.MustCompile(`(?is)<option[^>]*>`)
)

// namedTag matches an element of the given kind carrying name="<name>".
func namedTag(element, name string) *regexp.Regexp {
	return regexp.MustCompile(`(?is)<` + element + `[^>]*\sname="` + regexp.QuoteMeta(name) + `"[^>]*>`)
}

// htmlValue returns a single tag's value attribute.
func htmlValue(tag []byte) string {
	match := valueAttr.FindSubmatch(tag)
	if len(match) < 2 {
		return ""
	}

	return string(match[1])
}

// htmlHasTag reports whether the page carries the named element at all, telling a page that
// isn't the expected form apart from one whose fields are merely unset.
func htmlHasTag(page []byte, element, name string) bool {
	return namedTag(element, name).Match(page)
}

// htmlChecked reports whether the named checkbox is ticked. Every input of that name is
// considered, since a hidden companion input may precede the checkbox.
func htmlChecked(page []byte, name string) bool {
	for _, tag := range namedTag("input", name).FindAll(page, -1) {
		if checkedAttr.Match(tag) {
			return true
		}
	}

	return false
}

// htmlCheckedRadio returns the value of the selected radio in a named group, or "" when
// nothing in the group is selected.
func htmlCheckedRadio(page []byte, name string) string {
	for _, tag := range namedTag("input", name).FindAll(page, -1) {
		if checkedAttr.Match(tag) {
			return htmlValue(tag)
		}
	}

	return ""
}

// htmlInputValue returns the first non-empty value among the inputs of that name, so an empty
// placeholder does not shadow the real one.
func htmlInputValue(page []byte, name string) string {
	for _, tag := range namedTag("input", name).FindAll(page, -1) {
		if value := htmlValue(tag); value != "" {
			return value
		}
	}

	return ""
}

// htmlSelected returns the value of the chosen option in the named select, scoped to that
// select's own block so several selects on a page don't answer for each other.
func htmlSelected(page []byte, name string) string {
	block := regexp.MustCompile(`(?is)<select[^>]*\sname="` + regexp.QuoteMeta(name) + `".*?</select>`).Find(page)
	if block == nil {
		return ""
	}

	for _, option := range optionTag.FindAll(block, -1) {
		if selectedAttr.Match(option) {
			return htmlValue(option)
		}
	}

	return ""
}
