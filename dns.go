package directadmin

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"

	"github.com/spf13/cast"
)

type DnsRecord struct {
	Name  string `json:"name"`
	Ttl   int    `json:"ttl"`
	Type  string `json:"type"`
	Value string `json:"value"`
}

// CheckDnsRecordExists (user) checks if the given dns record exists on the server
//
// checkField can be either "name" or "value"
func (c *UserContext) CheckDnsRecordExists(checkField string, domain string, dnsRecord DnsRecord) error {
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

// CreateDnsRecord (user) creates the provided dns record for the given domain
func (c *UserContext) CreateDnsRecord(domain string, dnsRecord DnsRecord) error {
	var response apiGenericResponse

	rawDnsRecordData := dnsRecord.translate()

	body := url.Values{}
	body.Set("domain", domain)
	body.Set("name", rawDnsRecordData.Name)
	body.Set("ttl", rawDnsRecordData.Ttl)
	body.Set("type", rawDnsRecordData.Type)
	body.Set("value", rawDnsRecordData.Value)

	if _, err := c.api.makeRequest(http.MethodPost, "API_DNS_CONTROL?action=add&action_pointers=yes", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "Record Added" {
		return fmt.Errorf("failed to create dns record: %v", response.Result)
	}

	return nil
}

// DeleteDnsRecords (user) deletes all the specified dnss for the session user
func (c *UserContext) DeleteDnsRecords(dnsRecords ...DnsRecord) error {
	var response apiGenericResponse

	body := url.Values{}
	dnsRecordMap := make(map[string][]string)

	for _, dnsRecord := range dnsRecords {
		dnsRecordMap[dnsRecord.Type] = append(dnsRecordMap[strings.ToLower(dnsRecord.Type)], fmt.Sprintf("name=%v&value=%v", dnsRecord.Name, dnsRecord.Value))
	}

	for dnsRecordType, dnsRecordData := range dnsRecordMap {
		for index, data := range dnsRecordData {
			body.Set(dnsRecordType+"recs"+cast.ToString(index), data)
		}
	}

	if _, err := c.api.makeRequest(http.MethodPost, "API_DNS_CONTROL?action=select&delete=yes", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "Records Deleted" {
		return fmt.Errorf("failed to delete dns records: %v", response.Result)
	}

	return nil
}

// GetDnsRecords (user) returns the given domain's dns records
func (c *UserContext) GetDnsRecords(domain string) ([]DnsRecord, error) {
	var dnsRecords []DnsRecord
	rawDnsRecords := struct {
		DnsRecords []rawDnsRecord `json:"records"`
	}{}

	if _, err := c.api.makeRequest(http.MethodGet, "API_DNS_CONTROL?domain="+domain, c.credentials, nil, &rawDnsRecords); err != nil {
		return nil, err
	}

	for _, dnsRecord := range rawDnsRecords.DnsRecords {
		dnsRecords = append(dnsRecords, dnsRecord.translate())
	}

	if len(dnsRecords) == 0 {
		return nil, errors.New("no dns records were found")
	}

	return dnsRecords, nil
}

// UpdateDnsRecord (user) updates the given dns record for the given domain
func (c *UserContext) UpdateDnsRecord(domain string, originalDnsRecord DnsRecord, updatedDnsRecord DnsRecord) error {
	var response apiGenericResponse

	rawDnsRecordData := updatedDnsRecord.translate()

	body := url.Values{}
	body.Set("domain", domain)
	body.Set("name", rawDnsRecordData.Name)
	body.Set("ttl", rawDnsRecordData.Ttl)
	body.Set("type", rawDnsRecordData.Type)
	body.Set("value", rawDnsRecordData.Value)
	body.Set(strings.ToLower(originalDnsRecord.Type)+"recs0", fmt.Sprintf("name=%v&value=%v", originalDnsRecord.Name, originalDnsRecord.Value))

	if _, err := c.api.makeRequest(http.MethodPost, "API_DNS_CONTROL?action=edit&action_pointers=yes", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "Record Added" {
		return fmt.Errorf("failed to create dns record: %v", response.Result)
	}

	return nil
}
