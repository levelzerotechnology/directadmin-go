package directadmin

import (
	"net/url"
	"testing"
)

// TestDNSRecordSelector checks that a record survives the round trip DirectAdmin performs:
// url.Values.Encode, then its own decode and re-split of the selector into name and value.
func TestDNSRecordSelector(t *testing.T) {
	records := []DNSRecord{
		{Name: "@", Value: "v=spf1 include:_spf.example.com ~all"},
		{Name: "test", Value: "a&b=c"},
		{Name: "with=equals", Value: "plain"},
	}

	for _, record := range records {
		body := url.Values{"txtrecs0": {dnsRecordSelector(record)}}

		decoded, err := url.ParseQuery(body.Encode())
		if err != nil {
			t.Fatal(err)
		}

		selector, err := url.ParseQuery(decoded.Get("txtrecs0"))
		if err != nil {
			t.Fatal(err)
		}

		if selector.Get("name") != record.Name || selector.Get("value") != record.Value {
			t.Fatalf("expected name=%q value=%q, got name=%q value=%q",
				record.Name, record.Value, selector.Get("name"), selector.Get("value"))
		}
	}
}
