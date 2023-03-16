package directadmin

import "github.com/spf13/cast"

type rawSysInfo struct {
	CPUs map[string]struct {
		MHz       string `json:"mhz"`
		ModelName string `json:"model_name"`
		VendorID  string `json:"vendor_id"`
	} `json:"cpus"`
	Load struct {
		Load1  string `json:"load_1"`
		Load5  string `json:"load_5"`
		Load15 string `json:"load_15"`
	} `json:"load"`
	MemInfo struct {
		Active            string `json:"Active"`
		ActiveAnon        string `json:"Active(anon)"`
		ActiveFile        string `json:"Active(file)"`
		AnonHugePages     string `json:"AnonHugePages"`
		AnonPages         string `json:"AnonPages"`
		Bounce            string `json:"Bounce"`
		Buffers           string `json:"Buffers"`
		Cached            string `json:"Cached"`
		CommitLimit       string `json:"CommitLimit"`
		CommittedAS       string `json:"Committed_AS"`
		DirectMap1G       string `json:"DirectMap1G"`
		DirectMap2M       string `json:"DirectMap2M"`
		DirectMap4K       string `json:"DirectMap4k"`
		Dirty             string `json:"Dirty"`
		FileHugePages     string `json:"FileHugePages"`
		FilePmdMapped     string `json:"FilePmdMapped"`
		HardwareCorrupted string `json:"HardwareCorrupted"`
		HugePagesFree     string `json:"HugePages_Free"`
		HugePagesRsvd     string `json:"HugePages_Rsvd"`
		HugePagesSurp     string `json:"HugePages_Surp"`
		HugePagesTotal    string `json:"HugePages_Total"`
		HugePageSize      string `json:"Hugepagesize"`
		HugeTlb           string `json:"Hugetlb"`
		Inactive          string `json:"Inactive"`
		InactiveAnon      string `json:"Inactive(anon)"`
		InactiveFile      string `json:"Inactive(file)"`
		KReclaimable      string `json:"KReclaimable"`
		KernelStack       string `json:"KernelStack"`
		Mapped            string `json:"Mapped"`
		MemAvailable      string `json:"MemAvailable"`
		MemFree           string `json:"MemFree"`
		MemTotal          string `json:"MemTotal"`
		MLocked           string `json:"Mlocked"`
		NfsUnstable       string `json:"NFS_Unstable"`
		PageTables        string `json:"PageTables"`
		PerCpu            string `json:"Percpu"`
		SReclaimable      string `json:"SReclaimable"`
		SUnreclaim        string `json:"SUnreclaim"`
		Shmem             string `json:"Shmem"`
		ShmemHugePages    string `json:"ShmemHugePages"`
		ShmemPmdMapped    string `json:"ShmemPmdMapped"`
		Slab              string `json:"Slab"`
		SwapCached        string `json:"SwapCached"`
		SwapFree          string `json:"SwapFree"`
		SwapTotal         string `json:"SwapTotal"`
		Unevictable       string `json:"Unevictable"`
		VmallocChunk      string `json:"VmallocChunk"`
		VmallocTotal      string `json:"VmallocTotal"`
		VmallocUsed       string `json:"VmallocUsed"`
		Writeback         string `json:"Writeback"`
		WritebackTmp      string `json:"WritebackTmp"`
	} `json:"mem_info"`
	NumberOfCPUs string `json:"numcpus"`
	Services     map[string]struct {
		InfoStr string `json:"info_str"`
		Name    string `json:"name"`
		Version string `json:"version"`
	} `json:"services"`
	UptimeInfo struct {
		Days         string `json:"days"`
		Hours        string `json:"hours"`
		Minutes      string `json:"minutes"`
		TotalSeconds string `json:"total_seconds"`
		Uptime       string `json:"uptime"`
	} `json:"uptime_info"`
}

func (r *rawSysInfo) parse() SysInfo {
	sysInfo := SysInfo{}

	sysInfo.CpuCount = cast.ToInt(r.NumberOfCPUs)

	counter := 0
	for _, cpu := range r.CPUs {
		sysInfo.Cpus[cast.ToString(counter)] = struct {
			MHz    float64 `json:"mhz"`
			Model  string  `json:"model"`
			Vendor string  `json:"vendor"`
		}{
			MHz:    cast.ToFloat64(cpu.MHz),
			Model:  cpu.ModelName,
			Vendor: cpu.VendorID,
		}
		counter++
	}

	sysInfo.SystemLoad.Last1Minute = r.Load.Load1
	sysInfo.SystemLoad.Last5Minutes = r.Load.Load5
	sysInfo.SystemLoad.Last15Minutes = r.Load.Load15

	sysInfo.MemInfo.Active = cast.ToInt(r.MemInfo.Active)
	sysInfo.MemInfo.ActiveAnon = cast.ToInt(r.MemInfo.ActiveAnon)
	sysInfo.MemInfo.ActiveFile = cast.ToInt(r.MemInfo.ActiveFile)
	sysInfo.MemInfo.AnonHugePages = cast.ToInt(r.MemInfo.AnonHugePages)
	sysInfo.MemInfo.AnonPages = cast.ToInt(r.MemInfo.AnonPages)
	sysInfo.MemInfo.Bounce = cast.ToInt(r.MemInfo.Bounce)
	sysInfo.MemInfo.Buffers = cast.ToInt(r.MemInfo.Buffers)
	sysInfo.MemInfo.Cached = cast.ToInt(r.MemInfo.Cached)
	sysInfo.MemInfo.CommitLimit = cast.ToInt(r.MemInfo.CommitLimit)
	sysInfo.MemInfo.CommittedAS = cast.ToInt(r.MemInfo.CommittedAS)
	sysInfo.MemInfo.DirectMap1G = cast.ToInt(r.MemInfo.DirectMap1G)
	sysInfo.MemInfo.DirectMap2M = cast.ToInt(r.MemInfo.DirectMap2M)
	sysInfo.MemInfo.DirectMap4K = cast.ToInt(r.MemInfo.DirectMap4K)
	sysInfo.MemInfo.Dirty = cast.ToInt(r.MemInfo.Dirty)
	sysInfo.MemInfo.FileHugePages = cast.ToInt(r.MemInfo.FileHugePages)
	sysInfo.MemInfo.FilePmdMapped = cast.ToInt(r.MemInfo.FilePmdMapped)
	sysInfo.MemInfo.HardwareCorrupted = cast.ToInt(r.MemInfo.HardwareCorrupted)
	sysInfo.MemInfo.HugePagesFree = cast.ToInt(r.MemInfo.HugePagesFree)
	sysInfo.MemInfo.HugePagesRsvd = cast.ToInt(r.MemInfo.HugePagesRsvd)
	sysInfo.MemInfo.HugePagesSurp = cast.ToInt(r.MemInfo.HugePagesSurp)
	sysInfo.MemInfo.HugePagesTotal = cast.ToInt(r.MemInfo.HugePagesTotal)
	sysInfo.MemInfo.HugePageSize = cast.ToInt(r.MemInfo.HugePageSize)
	sysInfo.MemInfo.HugeTlb = cast.ToInt(r.MemInfo.HugeTlb)
	sysInfo.MemInfo.Inactive = cast.ToInt(r.MemInfo.Inactive)
	sysInfo.MemInfo.InactiveAnon = cast.ToInt(r.MemInfo.InactiveAnon)
	sysInfo.MemInfo.InactiveFile = cast.ToInt(r.MemInfo.InactiveFile)
	sysInfo.MemInfo.KReclaimable = cast.ToInt(r.MemInfo.KReclaimable)
	sysInfo.MemInfo.KernelStack = cast.ToInt(r.MemInfo.KernelStack)
	sysInfo.MemInfo.Mapped = cast.ToInt(r.MemInfo.Mapped)
	sysInfo.MemInfo.MemAvailable = cast.ToInt(r.MemInfo.MemAvailable)
	sysInfo.MemInfo.MemFree = cast.ToInt(r.MemInfo.MemFree)
	sysInfo.MemInfo.MemTotal = cast.ToInt(r.MemInfo.MemTotal)
	sysInfo.MemInfo.MLocked = cast.ToInt(r.MemInfo.MLocked)
	sysInfo.MemInfo.NfsUnstable = cast.ToInt(r.MemInfo.NfsUnstable)
	sysInfo.MemInfo.PageTables = cast.ToInt(r.MemInfo.PageTables)
	sysInfo.MemInfo.PerCpu = cast.ToInt(r.MemInfo.PerCpu)
	sysInfo.MemInfo.SReclaimable = cast.ToInt(r.MemInfo.SReclaimable)
	sysInfo.MemInfo.SUnreclaim = cast.ToInt(r.MemInfo.SUnreclaim)
	sysInfo.MemInfo.Shmem = cast.ToInt(r.MemInfo.Shmem)
	sysInfo.MemInfo.ShmemHugePages = cast.ToInt(r.MemInfo.ShmemHugePages)
	sysInfo.MemInfo.ShmemPmdMapped = cast.ToInt(r.MemInfo.ShmemPmdMapped)
	sysInfo.MemInfo.Slab = cast.ToInt(r.MemInfo.Slab)
	sysInfo.MemInfo.SwapCached = cast.ToInt(r.MemInfo.SwapCached)
	sysInfo.MemInfo.SwapFree = cast.ToInt(r.MemInfo.SwapFree)
	sysInfo.MemInfo.SwapTotal = cast.ToInt(r.MemInfo.SwapTotal)
	sysInfo.MemInfo.Unevictable = cast.ToInt(r.MemInfo.Unevictable)
	sysInfo.MemInfo.VmallocChunk = cast.ToInt(r.MemInfo.VmallocChunk)
	sysInfo.MemInfo.VmallocTotal = cast.ToInt(r.MemInfo.VmallocTotal)
	sysInfo.MemInfo.VmallocUsed = cast.ToInt(r.MemInfo.VmallocUsed)
	sysInfo.MemInfo.Writeback = cast.ToInt(r.MemInfo.Writeback)
	sysInfo.MemInfo.WritebackTmp = cast.ToInt(r.MemInfo.WritebackTmp)

	for serviceId, service := range r.Services {
		sysInfo.Services[serviceId] = struct {
			Name    string `json:"name"`
			Status  string `json:"status"`
			Version string `json:"version"`
		}{
			Name:    service.Name,
			Status:  service.InfoStr,
			Version: service.Version,
		}
	}

	sysInfo.Uptime.Days = r.UptimeInfo.Days
	sysInfo.Uptime.Hours = r.UptimeInfo.Hours
	sysInfo.Uptime.Minutes = r.UptimeInfo.Minutes
	sysInfo.Uptime.TotalSeconds = r.UptimeInfo.TotalSeconds
	sysInfo.Uptime.Uptime = r.UptimeInfo.Uptime

	return sysInfo
}
