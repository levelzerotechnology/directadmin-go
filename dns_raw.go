package directadmin

import "github.com/spf13/cast"

type rawDnsRecord struct {
	Name  string `json:"name"`
	Ttl   string `json:"ttl"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (d *DnsRecord) translate() rawDnsRecord {
	return rawDnsRecord{
		Name:  d.Name,
		Ttl:   cast.ToString(d.Ttl),
		Type:  d.Type,
		Value: d.Value,
	}
}

func (d *rawDnsRecord) translate() DnsRecord {
	return DnsRecord{
		Name:  d.Name,
		Ttl:   cast.ToInt(d.Ttl),
		Type:  d.Type,
		Value: d.Value,
	}
}
