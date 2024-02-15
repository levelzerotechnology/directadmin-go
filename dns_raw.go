package directadmin

import "github.com/spf13/cast"

type rawDNSRecord struct {
	Name  string `json:"name"`
	Ttl   string `json:"ttl"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

func (d *DNSRecord) translate() rawDNSRecord {
	return rawDNSRecord{
		Name:  d.Name,
		Ttl:   cast.ToString(d.Ttl),
		Type:  d.Type,
		Value: d.Value,
	}
}

func (d *rawDNSRecord) translate() DNSRecord {
	return DNSRecord{
		Name:  d.Name,
		Ttl:   cast.ToInt(d.Ttl),
		Type:  d.Type,
		Value: d.Value,
	}
}
