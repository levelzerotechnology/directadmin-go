package directadmin

import (
	"fmt"
	"net/http"
	"time"
)

type SysInfo struct {
	CpuCount int `json:"cpuCount" yaml:"cpuCount"`
	Cpus     map[string]struct {
		MHz    float64 `json:"mhz"`
		Model  string  `json:"model"`
		Vendor string  `json:"vendor"`
	}
	SystemLoad struct {
		Last1Minute   string `json:"last1Minute"`
		Last5Minutes  string `json:"last5Minutes"`
		Last15Minutes string `json:"last15Minutes"`
	} `json:"systemLoad"`
	MemInfo struct {
		Active            int `json:"active"`
		ActiveAnon        int `json:"activeAnon"`
		ActiveFile        int `json:"activeFile"`
		AnonHugePages     int `json:"anonHugePages"`
		AnonPages         int `json:"anonPages"`
		Bounce            int `json:"bounce"`
		Buffers           int `json:"buffers"`
		Cached            int `json:"cached"`
		CommitLimit       int `json:"commitLimit"`
		CommittedAS       int `json:"committedAs"`
		DirectMap1G       int `json:"directMap1G"`
		DirectMap2M       int `json:"directMap2M"`
		DirectMap4K       int `json:"directMap4K"`
		Dirty             int `json:"Dirty"`
		FileHugePages     int `json:"fileHugePages"`
		FilePmdMapped     int `json:"filePmdMapped"`
		HardwareCorrupted int `json:"hardwareCorrupted"`
		HugePagesFree     int `json:"hugePagesFree"`
		HugePagesRsvd     int `json:"hugePagesRsvd"`
		HugePagesSurp     int `json:"hugePagesSurp"`
		HugePagesTotal    int `json:"hugePagesTotal"`
		HugePageSize      int `json:"hugePageSize"`
		HugeTlb           int `json:"hugeTlb"`
		Inactive          int `json:"inactive"`
		InactiveAnon      int `json:"inactiveAnon"`
		InactiveFile      int `json:"inactiveFile"`
		KReclaimable      int `json:"kReclaimable"`
		KernelStack       int `json:"kernelStack"`
		Mapped            int `json:"mapped"`
		MemAvailable      int `json:"memAvailable"`
		MemFree           int `json:"memFree"`
		MemTotal          int `json:"memTotal"`
		MLocked           int `json:"mLocked"`
		NfsUnstable       int `json:"nfsUnstable"`
		PageTables        int `json:"pageTables"`
		PerCpu            int `json:"perCpu"`
		SReclaimable      int `json:"sReclaimable"`
		SUnreclaim        int `json:"sUnreclaim"`
		Shmem             int `json:"shmem"`
		ShmemHugePages    int `json:"shmemHugePages"`
		ShmemPmdMapped    int `json:"shmemPmdMapped"`
		Slab              int `json:"slab"`
		SwapCached        int `json:"swapCached"`
		SwapFree          int `json:"swapFree"`
		SwapTotal         int `json:"swapTotal"`
		Unevictable       int `json:"snevictable"`
		VmallocChunk      int `json:"vmallocChunk"`
		VmallocTotal      int `json:"vmallocTotal"`
		VmallocUsed       int `json:"vmallocUsed"`
		Writeback         int `json:"writeback"`
		WritebackTmp      int `json:"writebackTmp"`
	} `json:"memory"`
	Services map[string]struct {
		Name    string `json:"name"`
		Status  string `json:"status"`
		Version string `json:"version"`
	} `json:"services"`
	Uptime struct {
		Days         string `json:"days"`
		Hours        string `json:"hours"`
		Minutes      string `json:"minutes"`
		TotalSeconds string `json:"totalSeconds"`
		Uptime       string `json:"uptime"`
	} `json:"uptime"`
}

type BasicSysInfo struct {
	AllowPasswordReset bool      `json:"allowPasswordReset"`
	Hostname           string    `json:"hostname"`
	Languages          []string  `json:"languages"`
	LicenseError       string    `json:"licenseError"`
	LicenseTrial       bool      `json:"licenseTrial"`
	LicenseValid       bool      `json:"licenseValid"`
	OtpTrustDays       int       `json:"otpTrustDays"`
	Time               time.Time `json:"time"`
}

func (c *UserContext) GetBasicSysInfo() (*BasicSysInfo, error) {
	var basicSysInfo BasicSysInfo

	if _, err := c.api.makeRequestN(http.MethodGet, "info", c.credentials, nil, &basicSysInfo); err != nil {
		return nil, fmt.Errorf("failed to get basic sys info: %v", err)
	}

	return &basicSysInfo, nil
}

//func (c *UserContext) GetPhpVersions() (*SysInfo, error) {
//	var rawSys rawSysInfo
//	var sys SysInfo
//
//	if _, err := c.api.makeRequest(http.MethodGet, "API_SYSTEM_INFO", c.credentials, &rawSys); err != nil {
//		return nil, err
//	}
//
//	sys = rawSys.parse()
//
//	return &sys, nil
//}

func (c *UserContext) GetSysInfo() (*SysInfo, error) {
	var rawSys rawSysInfo
	var sys SysInfo

	if _, err := c.api.makeRequest(http.MethodGet, "API_SYSTEM_INFO", c.credentials, nil, &rawSys); err != nil {
		return nil, err
	}

	sys = rawSys.parse()

	return &sys, nil
}
