package directadmin

import "github.com/spf13/cast"

type rawResellerPackage struct {
	rawPackage
	OversellEnabled    string `json:"oversell" url:"oversell"`
	UserQuota          string `json:"nuser,omitempty" url:"nuser,omitempty"`
	UserQuotaUnlimited string `json:"unusers,omitempty" url:"unusers,omitempty"` // unique unlimited field DA requires for an unlimited user quota
}

func (p *ResellerPackage) translate() (pack rawResellerPackage) {
	pack = rawResellerPackage{
		rawPackage:      p.Package.translate(),
		OversellEnabled: reverseParseOnOff(p.OversellEnabled, false),
	}

	if p.UserQuota == -1 {
		pack.UserQuota = ""
		pack.UserQuotaUnlimited = "yes"
	} else {
		pack.UserQuota = cast.ToString(p.UserQuota)
		pack.UserQuotaUnlimited = ""
	}

	return pack
}

func (p *rawResellerPackage) translate() (pack ResellerPackage) {
	pack = ResellerPackage{
		Package:         p.rawPackage.translate(),
		OversellEnabled: parseOnOff(p.OversellEnabled),
		UserQuota:       parseNum(p.UserQuota),
	}

	if p.UserQuotaUnlimited == "yes" {
		pack.UserQuota = -1
	} else if p.UserQuota == "" {
		pack.UserQuota = 0
	}

	return pack
}
