package directadmin

import (
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"sync"

	"github.com/google/go-querystring/query"
	"github.com/spf13/cast"
)

type Package struct {
	AnonymousFtpEnabled     bool   `json:"anonymousFtpEnabled" yaml:"anonymousFtpEnabled"`
	BandwidthQuota          int    `json:"bandwidthQuota" yaml:"bandwidthQuota"`
	CpuQuota                int    `json:"cpuQuota" yaml:"cpuQuota"`
	CatchallEnabled         bool   `json:"catchallEnabled" yaml:"catchallEnabled"`
	CgiEnabled              bool   `json:"cgiEnabled" yaml:"cgiEnabled"`
	CronEnabled             bool   `json:"cronEnabled" yaml:"cronEnabled"`
	DnsControlEnabled       bool   `json:"dnsControlEnabled" yaml:"dnsControlEnabled"`
	DomainPointerQuota      int    `json:"domainPointerQuota" yaml:"domainPointerQuota"`
	DomainQuota             int    `json:"domainQuota" yaml:"domainQuota"`
	EmailAutoresponderQuota int    `json:"emailAutoresponderQuota" yaml:"emailAutoresponderQuota"`
	EmailForwarderQuota     int    `json:"emailForwarderQuota" yaml:"emailForwarderQuota"`
	EmailMailingListQuota   int    `json:"emailMailingListQuota" yaml:"emailMailingListQuota"`
	EmailQuota              int    `json:"emailQuota" yaml:"emailQuota"`
	FtpQuota                int    `json:"ftp" yaml:"ftpQuota"`
	GitEnabled              bool   `json:"gitEnabled" yaml:"gitEnabled"`
	IoReadBandwidthMax      int    `json:"ioReadBandwidthMax" yaml:"ioReadBandwidthMax"`
	IoReadIopsMax           int    `json:"ioReadIopsMax" yaml:"ioReadIopsMax"`
	IoWriteBandwidthMax     int    `json:"ioWriteBandwidthMax" yaml:"ioWriteBandwidthMax"`
	IoWriteIopsMax          int    `json:"ioWriteIopsMax" yaml:"ioWriteIopsMax"`
	InodeQuota              int    `json:"inodeQuota" yaml:"inodeQuota"`
	JailEnabled             bool   `json:"jailEnabled" yaml:"jailEnabled"`
	Language                string `json:"language" yaml:"language"`
	LoginKeysEnabled        bool   `json:"loginKeysEnabled" yaml:"loginKeysEnabled"`
	MemoryHigh              int    `json:"memoryHigh" yaml:"memoryHigh"`
	MemoryMax               int    `json:"memoryMax" yaml:"memoryMax"`
	MysqlQuota              int    `json:"mysqlQuota" yaml:"mysqlQuota"`
	Name                    string `json:"name" yaml:"name"`
	NginxEnabled            bool   `json:"nginxEnabled" yaml:"nginxEnabled"`
	PhpEnabled              bool   `json:"phpEnabled" yaml:"phpEnabled"`
	Quota                   int    `json:"quota" yaml:"quota"`
	RedisEnabled            bool   `json:"redisEnabled" yaml:"redisEnabled"`
	SshEnabled              bool   `json:"sshEnabled" yaml:"sshEnabled"`
	Skin                    string `json:"skin" yaml:"skin"`
	SpamAssassinEnabled     bool   `json:"spamAssassinEnabled" yaml:"spamAssassinEnabled"`
	SslEnabled              bool   `json:"sslEnabled" yaml:"sslEnabled"`
	SubdomainQuota          int    `json:"subdomainQuota" yaml:"subdomainQuota"`
	SuspendAtLimitEnabled   bool   `json:"suspendAtLimitEnabled" yaml:"suspendAtLimitEnabled"`
	SysInfoEnabled          bool   `json:"sysInfoEnabled" yaml:"sysInfoEnabled"`
	TasksMax                int    `json:"tasksMax" yaml:"tasksMax"`
	WordpressEnabled        bool   `json:"wordpressEnabled" yaml:"wordpressEnabled"`
}

// CreatePackage (reseller) creates the provided package
func (c *ResellerContext) CreatePackage(pack Package) error {
	var response apiGenericResponse

	body, err := query.Values(pack.translate())
	if err != nil {
		return err
	}

	if _, err = c.api.makeRequest(http.MethodPost, "MANAGE_USER_PACKAGES?add=yes", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "Saved" {
		return fmt.Errorf("failed to create package: %v", response.Result)
	}

	return nil
}

// DeletePackages (reseller) deletes all the specified packs for the session user
func (c *ResellerContext) DeletePackages(packs ...string) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("delete", "yes")

	for index, pack := range packs {
		body.Set("delete"+cast.ToString(index), pack)
		body.Set("select"+cast.ToString(index), pack)
	}

	if _, err := c.api.makeRequest(http.MethodPost, "MANAGE_USER_PACKAGES", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "Deleted" {
		return fmt.Errorf("failed to delete pack: %v", response.Result)
	}

	return nil
}

// GetPackage (reseller) returns the single specified package
func (c *ResellerContext) GetPackage(packageName string) (Package, error) {
	var rawPack rawPackage

	if _, err := c.api.makeRequest(http.MethodGet, "API_PACKAGES_USER?package="+packageName, c.credentials, nil, &rawPack); err != nil {
		return Package{}, fmt.Errorf("failed to get package info for %v: %v", packageName, err)
	}

	rawPack.Name = packageName

	return rawPack.translate(), nil
}

// GetPackages (reseller) returns all packages belonging to the session user
func (c *ResellerContext) GetPackages() ([]Package, error) {
	var packageList []string
	var packages []Package

	if _, err := c.api.makeRequest(http.MethodGet, "API_PACKAGES_USER", c.credentials, nil, &packageList); err != nil {
		return nil, err
	}

	var errs []error
	var wg sync.WaitGroup
	var mu sync.Mutex
	wg.Add(len(packageList))

	for _, packageName := range packageList {
		// convert to local variable to prevent variable overwrite
		packageToProcess := packageName

		go func(packageName string) {
			defer wg.Done()

			pack, err := c.GetPackage(packageName)
			if err != nil {
				mu.Lock()
				errs = append(errs, err)
				mu.Unlock()
				return
			}

			mu.Lock()
			packages = append(packages, pack)
			mu.Unlock()
		}(packageToProcess)
	}

	wg.Wait()

	if len(errs) > 0 {
		counter := 0
		var errStrings []string

		for _, err := range errs {
			counter++
			errStrings = append(errStrings, "error "+cast.ToString(counter)+": "+err.Error())
		}

		return nil, errors.New(strings.Join(errStrings, "; "))
	}

	if len(packages) == 0 {
		return nil, errors.New("no packages could be found")
	}

	return packages, nil
}

// RenamePackage (reseller) renames the provided package
func (c *ResellerContext) RenamePackage(oldPackName string, newPackName string) error {
	var response apiGenericResponse

	body := url.Values{}
	body.Set("old_package", oldPackName)
	body.Set("new_package", newPackName)

	if _, err := c.api.makeRequest(http.MethodPost, "MANAGE_USER_PACKAGES?action=rename", c.credentials, body, &response); err != nil {
		return err
	}

	if response.Success != "Saved" {
		return fmt.Errorf("failed to rename package: %v", response.Result)
	}

	return nil
}

// UpdatePackage (reseller) accepts a Package object and updates the version on DA with it
func (c *ResellerContext) UpdatePackage(pack Package) error {
	// DA's update functionality is virtually identical to create, so we'll just use that
	return c.CreatePackage(pack)
}
