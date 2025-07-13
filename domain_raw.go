package directadmin

import (
	"strconv"

	"github.com/spf13/cast"
)

type rawDomain struct {
	Active         string `json:"active"`
	BandwidthQuota string `json:"bandwidth_limit"`
	BandwidthUsage string `json:"bandwidth"`
	CGIEnabled     string `json:"cgi"`
	DefaultDomain  string `json:"defaultdomain"`
	DiskQuota      string `json:"quota_limit"`
	DiskUsage      string `json:"quota"`
	Domain         string `json:"domain"`
	ExtraData      struct {
		ModSecurityEnabled string `json:"modsecurity"`
		PHPOptions         map[string]struct {
			Index    string `json:"value"`
			Selected string `json:"selected"`
			Version  string `json:"text"`
		} `json:"php1_select"`
		PHPSelectorEnabled string `json:"has_php_selector"`
	} `json:"-"`
	IPAddresses        []string `json:"ips"`
	OpenBaseDirEnabled string   `json:"open_basedir"`
	PHPEnabled         string   `json:"php"`
	SafeMode           string   `json:"safemode"`
	SSLEnabled         string   `json:"ssl"`
	Subdomains         []string `json:"subdomains"`
	SubdomainUsage     string   `json:"subdomain"`
	Suspended          string   `json:"suspended"`
	Username           string   `json:"username"`
}

// translate returns an rawDomain object with DA's original JSON fields and data types.
func (d *Domain) translate() (domain rawDomain) {
	domain = rawDomain{
		Active:             reverseParseYesNo(d.Active),
		BandwidthQuota:     reverseParseNum(d.BandwidthQuota, false),
		BandwidthUsage:     reverseParseNum(d.BandwidthUsage, false),
		CGIEnabled:         reverseParseOnOff(d.CGIEnabled),
		DefaultDomain:      reverseParseYesNo(d.DefaultDomain),
		DiskQuota:          reverseParseNum(d.DiskQuota, false),
		DiskUsage:          reverseParseNum(d.DiskUsage, true),
		Domain:             d.Domain,
		IPAddresses:        d.IPAddresses,
		OpenBaseDirEnabled: reverseParseOnOff(d.OpenBaseDirEnabled),
		PHPEnabled:         reverseParseOnOff(d.PHPEnabled),
		SafeMode:           reverseParseOnOff(d.SafeMode),
		SSLEnabled:         reverseParseOnOff(d.SSLEnabled),
		Subdomains:         d.Subdomains,
		SubdomainUsage:     strconv.Itoa(d.SubdomainUsage),
		Suspended:          reverseParseYesNo(d.Suspended),
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

	domain.ExtraData.PHPSelectorEnabled = reverseParseYesNo(d.PHPSelectorEnabled)

	return domain
}

// translate returns a Domain object.
func (d *rawDomain) translate() (domain Domain) {
	domain = Domain{
		Active:             parseOnOff(d.Active),
		BandwidthQuota:     parseNum(d.BandwidthQuota),
		BandwidthUsage:     parseNum(d.BandwidthUsage),
		CGIEnabled:         parseOnOff(d.CGIEnabled),
		DefaultDomain:      parseOnOff(d.DefaultDomain),
		DiskQuota:          parseNum(d.DiskQuota),
		DiskUsage:          parseNum(d.DiskUsage),
		Domain:             d.Domain,
		IPAddresses:        d.IPAddresses,
		ModSecurityEnabled: parseOnOff(d.ExtraData.ModSecurityEnabled),
		OpenBaseDirEnabled: parseOnOff(d.OpenBaseDirEnabled),
		PHPEnabled:         parseOnOff(d.PHPEnabled),
		PHPSelectorEnabled: parseOnOff(d.ExtraData.PHPSelectorEnabled),
		SafeMode:           parseOnOff(d.SafeMode),
		SSLEnabled:         parseOnOff(d.SSLEnabled),
		Subdomains:         d.Subdomains,
		SubdomainUsage:     cast.ToInt(d.SubdomainUsage),
		Suspended:          parseOnOff(d.Suspended),
		Username:           d.Username,
	}

	for _, option := range d.ExtraData.PHPOptions {
		if option.Selected == "yes" {
			domain.PHPVersion = option.Version
		}
	}

	return domain
}
