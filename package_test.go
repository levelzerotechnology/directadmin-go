package directadmin

import (
	"testing"

	"github.com/goccy/go-json"
)

const daPackage = `{"CPUQuota":"","IOReadBandwidthMax":"","IOReadIOPSMax":"","IOWriteBandwidthMax":"","IOWriteIOPSMax":"","MemoryHigh":"","MemoryMax":"","TasksMax":"","aftp":"OFF","bandwidth":"unlimited","catchall":"OFF","cgi":"OFF","cron":"ON","dnscontrol":"ON","domainptr":"unlimited","ftp":"9223372036854775807","git":"OFF","inode":"unlimited","jail":"ON","language":"en","login_keys":"ON","mysql":"unlimited","nemailf":"unlimited","nemailml":"unlimited","nemailr":"unlimited","nemails":"unlimited","nginx_unit":"ON","nsubdomains":"unlimited","php":"ON","quota":"unlimited","redis":"OFF","skin":"evolution","spam":"OFF","ssh":"OFF","ssl":"ON","suspend_at_limit":"ON","sysinfo":"OFF","vdomains":"unlimited","wordpress":"OFF"}`

func TestPackageTranslation(t *testing.T) {
	var rawPack rawPackage
	var pack Package

	if err := json.Unmarshal([]byte(daPackage), &rawPack); err != nil {
		t.Fatal(err)
	}

	pack = rawPack.translate()
	convertedPack := pack.translate()

	if convertedPack != rawPack {
		t.Fatalf("Expected %s\nGot %s", rawPack, convertedPack)
	}
}
