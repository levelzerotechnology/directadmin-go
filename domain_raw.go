package directadmin

import (
	"strconv"

	"github.com/spf13/cast"
)

type rawDomain struct {
	Active         string `json:"active"`
	BandwidthQuota string `json:"bandwidth_limit"`
	BandwidthUsage string `json:"bandwidth"`
	CgiEnabled     string `json:"cgi"`
	DefaultDomain  string `json:"defaultdomain"`
	DiskQuota      string `json:"quota_limit"`
	DiskUsage      string `json:"quota"`
	Domain         string `json:"domain"`
	ExtraData      struct {
		ModSecurityEnabled string `json:"modsecurity"`
		PhpOptions         map[string]struct {
			Index    string `json:"value"`
			Selected string `json:"selected"`
			Version  string `json:"text"`
		} `json:"php1_select"`
		PhpSelectorEnabled string `json:"has_php_selector"`
	} `json:"-"`
	IpAddresses        []string `json:"ips"`
	OpenBaseDirEnabled string   `json:"open_basedir"`
	PhpEnabled         string   `json:"php"`
	SafeMode           string   `json:"safemode"`
	SslEnabled         string   `json:"ssl"`
	Subdomains         []string `json:"subdomains"`
	SubdomainUsage     string   `json:"subdomain"`
	Suspended          string   `json:"suspended"`
	Username           string   `json:"username"`
}

// translate returns an rawDomain object with DA's original JSON fields and data types
func (d *Domain) translate() (domain rawDomain) {
	domain = rawDomain{
		Active:             reverseParseYesNo(d.Active, true),
		BandwidthQuota:     reverseParseNum(d.BandwidthQuota, false),
		BandwidthUsage:     reverseParseNum(d.BandwidthUsage, false),
		CgiEnabled:         reverseParseOnOff(d.CgiEnabled, false),
		DefaultDomain:      reverseParseYesNo(d.DefaultDomain, true),
		DiskQuota:          reverseParseNum(d.DiskQuota, false),
		DiskUsage:          reverseParseNum(d.DiskUsage, true),
		Domain:             d.Domain,
		IpAddresses:        d.IpAddresses,
		OpenBaseDirEnabled: reverseParseOnOff(d.OpenBaseDirEnabled, false),
		PhpEnabled:         reverseParseOnOff(d.PhpEnabled, false),
		SafeMode:           reverseParseOnOff(d.SafeMode, false),
		SslEnabled:         reverseParseOnOff(d.SslEnabled, false),
		Subdomains:         d.Subdomains,
		SubdomainUsage:     strconv.Itoa(d.SubdomainUsage),
		Suspended:          reverseParseYesNo(d.Suspended, true),
		Username:           d.Username,
	}

	// bug with the api where it returns 0 for unlimited bandwidth
	if domain.BandwidthUsage == "unlimited" {
		domain.BandwidthUsage = "0"
	}

	domain.ExtraData.ModSecurityEnabled = "no"
	if d.ModSecurityEnabled {
		domain.ExtraData.ModSecurityEnabled = "yes"
	}

	domain.ExtraData.PhpSelectorEnabled = reverseParseYesNo(d.PhpSelectorEnabled, true)

	return domain
}

// translate returns a Domain object
func (d *rawDomain) translate() (domain Domain) {
	domain = Domain{
		Active:             parseOnOff(d.Active),
		BandwidthQuota:     parseNum(d.BandwidthQuota),
		BandwidthUsage:     parseNum(d.BandwidthUsage),
		CgiEnabled:         parseOnOff(d.CgiEnabled),
		DefaultDomain:      parseOnOff(d.DefaultDomain),
		DiskQuota:          parseNum(d.DiskQuota),
		DiskUsage:          parseNum(d.DiskUsage),
		Domain:             d.Domain,
		IpAddresses:        d.IpAddresses,
		ModSecurityEnabled: parseOnOff(d.ExtraData.ModSecurityEnabled),
		OpenBaseDirEnabled: parseOnOff(d.OpenBaseDirEnabled),
		PhpEnabled:         parseOnOff(d.PhpEnabled),
		PhpSelectorEnabled: parseOnOff(d.ExtraData.PhpSelectorEnabled),
		SafeMode:           parseOnOff(d.SafeMode),
		SslEnabled:         parseOnOff(d.SslEnabled),
		Subdomains:         d.Subdomains,
		SubdomainUsage:     cast.ToInt(d.SubdomainUsage),
		Suspended:          parseOnOff(d.Suspended),
		Username:           d.Username,
	}

	for _, option := range d.ExtraData.PhpOptions {
		if option.Selected == "yes" {
			domain.PhpVersion = option.Version
		}
	}

	return domain
}
