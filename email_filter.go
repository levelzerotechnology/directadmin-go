package directadmin

import (
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"sort"
	"strconv"
)

const (
	// EmailFilterTypeEmail blocks mail from a single sender address.
	EmailFilterTypeEmail = "email"

	// EmailFilterTypeDomain blocks mail from every sender at a domain.
	EmailFilterTypeDomain = "domain"
)

// EmailFilter is a domain-wide incoming mail filter. Filters belong to the domain rather
// than to an individual mailbox, so one covers every address on it.
type EmailFilter struct {
	Type  string `json:"type"`
	Value string `json:"value"`
}

// GetEmailFilters (user) returns the incoming mail filters for the given domain.
func (c *UserContext) GetEmailFilters(domain string) ([]EmailFilter, error) {
	var response struct {
		Filters json.RawMessage `json:"filters"`
	}

	if _, err := c.makeRequestOld(http.MethodGet, "EMAIL_FILTER?domain="+url.QueryEscape(domain), nil, &response); err != nil {
		return nil, fmt.Errorf("failed to get email filters: %w", err)
	}

	if len(response.Filters) == 0 {
		return []EmailFilter{}, nil
	}

	// DirectAdmin returns the filters as an object keyed by index rather than an array, but
	// emits an empty collection as [], so both shapes are accepted.
	var byIndex map[string]EmailFilter

	if err := json.Unmarshal(response.Filters, &byIndex); err != nil {
		var asList []EmailFilter

		if listErr := json.Unmarshal(response.Filters, &asList); listErr != nil {
			return nil, fmt.Errorf("failed to parse email filters: %w", err)
		}

		return asList, nil
	}

	keys := make([]string, 0, len(byIndex))
	for key := range byIndex {
		keys = append(keys, key)
	}

	// Keyed by a numeric index, so sort numerically to preserve DirectAdmin's order.
	sort.Slice(keys, func(i, j int) bool {
		left, leftErr := strconv.Atoi(keys[i])
		right, rightErr := strconv.Atoi(keys[j])

		if leftErr != nil || rightErr != nil {
			return keys[i] < keys[j]
		}

		return left < right
	})

	filters := make([]EmailFilter, 0, len(keys))
	for _, key := range keys {
		filters = append(filters, byIndex[key])
	}

	return filters, nil
}

// AddEmailFilter (user) adds an incoming mail filter to the given domain. filterType is
// one of the EmailFilterType constants.
//
// DirectAdmin accepts a duplicate without complaint and stores it twice, so callers that
// may run more than once should check GetEmailFilters first.
func (c *UserContext) AddEmailFilter(domain string, filterType string, value string) error {
	body := url.Values{}
	body.Set("action", "add")
	body.Set("domain", domain)
	body.Set("type", filterType)
	body.Set("value", value)

	var response apiGenericResponse

	if _, err := c.makeRequestOld(http.MethodPost, "EMAIL_FILTER", body, &response); err != nil {
		return fmt.Errorf("failed to add email filter: %w", err)
	}

	if response.Success == "" {
		return fmt.Errorf("failed to add email filter: %v", errorOrResult(response))
	}

	return nil
}

// BlockEmailSender (user) blocks incoming mail from a single address across every mailbox
// on the given domain.
func (c *UserContext) BlockEmailSender(domain string, address string) error {
	return c.AddEmailFilter(domain, EmailFilterTypeEmail, address)
}
