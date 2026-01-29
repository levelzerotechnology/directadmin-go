package directadmin

import (
	"encoding/json"
	"strconv"
	"time"

	"github.com/spf13/cast"
)

type rawShownUsers struct {
	Users []*rawShownUser
}

func (r *rawShownUsers) UnmarshalJSON(data []byte) error {
	var raw map[string]json.RawMessage
	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}

	for key, value := range raw {
		if _, err := strconv.Atoi(key); err == nil {
			var user *rawShownUser
			if err = json.Unmarshal(value, &user); err != nil {
				return err
			}

			r.Users = append(r.Users, user)
		}
	}

	return nil
}

func (r *rawShownUsers) translate() []*User {
	users := make([]*User, len(r.Users))
	for i, u := range r.Users {
		users[i] = u.translate()
	}

	return users
}

type rawShownUser struct {
	Username  string `json:"username"`
	Bandwidth struct {
		Limit string `json:"limit"`
		Usage string `json:"usage"`
	} `json:"bandwidth"`
	Quota struct {
		Limit string `json:"limit"`
		Usage string `json:"usage"`
	} `json:"quota"`
	VDomains struct {
		Limit string `json:"limit"`
		Usage string `json:"usage"`
	} `json:"vdomains"`
	Package   string `json:"package"`
	Suspended struct {
		Reason string `json:"reason"`
		Value  string `json:"value"`
	} `json:"suspended"`
	IP          []string         `json:"ip"`
	Domains     map[string][]any `json:"domains"`
	DateCreated string           `json:"date_created"`
}

func (r *rawShownUser) translate() *User {
	domains := make([]string, 0, len(r.Domains))

	for domain := range r.Domains {
		domains = append(domains, domain)
	}

	return &User{
		Created:         time.Unix(cast.ToInt64(r.DateCreated), 0),
		Username:        r.Username,
		BandwidthLimit:  parseNum(r.Bandwidth.Limit),
		BandwidthUsage:  parseNum(r.Bandwidth.Usage),
		QuotaLimit:      parseNum(r.Quota.Limit),
		QuotaUsage:      parseNum(r.Quota.Usage),
		DomainsLimit:    parseNum(r.VDomains.Limit),
		DomainsUsage:    parseNum(r.VDomains.Usage),
		Package:         r.Package,
		Suspended:       r.Suspended.Value == "yes",
		SuspendedReason: r.Suspended.Reason,
		IPs:             r.IP,
		Domains:         domains,
	}
}
