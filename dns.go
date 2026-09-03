package directadmin

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cast"
)

type DNSRecord struct {
	Name  string `json:"name"`
	TTL   int    `json:"ttl"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// CheckDNSRecordExists (user) checks if the given dns record exists on the server.
//
// checkField can be either "name" or "value".
func (c *UserContext) CheckDNSRecordExists(checkField string, domain string, dnsRecord DNSRecord) error {
	body := url.Values{
		"check":  {checkField},
		"domain": {domain},
		"name":   {dnsRecord.Name},
		"record": {dnsRecord.Type},
		"type":   {"dns"},
		"value":  {dnsRecord.Value},
	}

	if dnsRecord.Type == "MX" {
		body.Set("mx_value", dnsRecord.Value)
	}

	return c.checkObjectExists(body)
}

// CreateDNSRecord (user) creates the provided DNS record for the given domain.
func (c *UserContext) CreateDNSRecord(domain string, dnsRecord DNSRecord) error {
	var response apiGenericResponse

	rawDNSRecordData := dnsRecord.translate()

	body := url.Values{
		"domain": {domain},
		"name":   {rawDNSRecordData.Name},
		"ttl":    {cast.ToString(rawDNSRecordData.TTL)},
		"type":   {rawDNSRecordData.Type},
		"value":  {rawDNSRecordData.Value},
	}

	if _, err := c.makeRequestOld(http.MethodPost, "API_DNS_CONTROL?action=add&action_pointers=yes", body, &response); err != nil {
		return err
	}

	if response.Success != "Record Added" {
		return fmt.Errorf("failed to create dns record: %v", response.Result)
	}

	return nil
}

// DeleteDNSRecords (user) deletes all the specified DNS records for the given domain.
func (c *UserContext) DeleteDNSRecords(domain string, dnsRecords ...DNSRecord) error {
	var response apiGenericResponse

	body := url.Values{"domain": {domain}}
	dnsRecordMap := make(map[string][]string)

	for _, dnsRecord := range dnsRecords {
		dnsType := strings.ToLower(dnsRecord.Type)

		dnsRecordMap[dnsType] = append(dnsRecordMap[dnsType], dnsRecordSelector(dnsRecord))
	}

	for dnsRecordType, dnsRecordData := range dnsRecordMap {
		for index, data := range dnsRecordData {
			body.Set(dnsRecordType+"recs"+cast.ToString(index), data)
		}
	}

	if _, err := c.makeRequestOld(http.MethodPost, "API_DNS_CONTROL?action=select&delete=yes", body, &response); err != nil {
		return err
	}

	if response.Success != "Records Deleted" {
		return fmt.Errorf("failed to delete dns records: %v", response.Result)
	}

	return nil
}

// dnsRecordSelector builds the "name=X&value=Y" selector DirectAdmin expects in its recsN
// parameters. Both halves are escaped because DirectAdmin re-parses the value as a query
// string, so an unescaped & or = in a record would split into extra parameters and target
// the wrong record.
func dnsRecordSelector(dnsRecord DNSRecord) string {
	return "name=" + url.QueryEscape(dnsRecord.Name) + "&value=" + url.QueryEscape(dnsRecord.Value)
}

// GetDNSRecords (user) returns the given domain's DNS records.
func (c *UserContext) GetDNSRecords(domain string) ([]DNSRecord, error) {
	var dnsRecords []DNSRecord
	rawDNSRecords := struct {
		DNSRecords []rawDNSRecord `json:"records"`
	}{}

	if _, err := c.makeRequestOld(http.MethodGet, "API_DNS_CONTROL?domain="+domain, nil, &rawDNSRecords); err != nil {
		return nil, err
	}

	for _, dnsRecord := range rawDNSRecords.DNSRecords {
		dnsRecords = append(dnsRecords, dnsRecord.translate())
	}

	if len(dnsRecords) == 0 {
		return nil, errors.New("no dns records were found")
	}

	return dnsRecords, nil
}

// UpdateDNSRecord (user) updates the given DNS record for the given domain.
func (c *UserContext) UpdateDNSRecord(domain string, originalDNSRecord DNSRecord, updatedDNSRecord DNSRecord) error {
	var response apiGenericResponse

	rawDNSRecordData := updatedDNSRecord.translate()

	body := url.Values{
		"domain": {domain},
		"name":   {rawDNSRecordData.Name},
		"ttl":    {cast.ToString(rawDNSRecordData.TTL)},
		"type":   {rawDNSRecordData.Type},
		"value":  {rawDNSRecordData.Value},
		strings.ToLower(originalDNSRecord.Type) + "recs0": {dnsRecordSelector(originalDNSRecord)},
	}

	if _, err := c.makeRequestOld(http.MethodPost, "API_DNS_CONTROL?action=edit&action_pointers=yes", body, &response); err != nil {
		return err
	}

	if response.Success != "Record Edited" {
		return fmt.Errorf("failed to update dns record: %v", response.Result)
	}

	return nil
}
