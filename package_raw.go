package directadmin

type rawPackage struct {
	AnonymousFtpEnabled     string `json:"aftp" url:"aftp"`
	BandwidthQuota          string `json:"bandwidth" url:"bandwidth"`
	CpuQuota                string `json:"CPUQuota" url:"CPUQuota"`
	CatchallEnabled         string `json:"catchall" url:"catchall"`
	CgiEnabled              string `json:"cgi" url:"cgi"`
	CronEnabled             string `json:"cron" url:"cron"`
	DNSControlEnabled       string `json:"dnscontrol" url:"dnscontrol"`
	DomainPointerQuota      string `json:"domainptr" url:"domainptr"`
	DomainQuota             string `json:"vdomains" url:"vdomains"`
	EmailAutoresponderQuota string `json:"nemailr" url:"nemailr"`
	EmailForwarderQuota     string `json:"nemailf" url:"nemailf"`
	EmailMailingListQuota   string `json:"nemailml" url:"nemailml"`
	EmailQuota              string `json:"nemails" url:"nemails"`
	FtpQuota                string `json:"ftp" url:"ftp"`
	GitEnabled              string `json:"git" url:"git"`
	IoReadBandwidthMax      string `json:"IOReadBandwidthMax" url:"IOReadBandwidthMax"`
	IoReadIopsMax           string `json:"IOReadIOPSMax" url:"IOReadIOPSMax"`
	IoWriteBandwidthMax     string `json:"IOWriteBandwidthMax" url:"IOWriteBandwidthMax"`
	IoWriteIopsMax          string `json:"IOWriteIOPSMax" url:"IOWriteIOPSMax"`
	InodeQuota              string `json:"inode" url:"inode"`
	JailEnabled             string `json:"jail" url:"jail"`
	Language                string `json:"language" url:"language"`
	LoginKeysEnabled        string `json:"login_keys" url:"login_keys"`
	MemoryHigh              string `json:"MemoryHigh" url:"MemoryHigh"`
	MemoryMax               string `json:"MemoryMax" url:"MemoryMax"`
	MysqlQuota              string `json:"mysql" url:"mysql"`
	Name                    string `json:"packagename" url:"packagename"`
	NginxEnabled            string `json:"nginx_unit" url:"nginx_unit"`
	PhpEnabled              string `json:"php" url:"php"`
	Quota                   string `json:"quota" url:"quota"`
	RedisEnabled            string `json:"redis" url:"redis"`
	SshEnabled              string `json:"ssh" url:"ssh"`
	Skin                    string `json:"skin" url:"skin"`
	SpamAssassinEnabled     string `json:"spam" url:"spam"`
	SslEnabled              string `json:"ssl" url:"ssl"`
	SubdomainQuota          string `json:"nsubdomains" url:"nsubdomains"`
	SuspendAtLimitEnabled   string `json:"suspend_at_limit" url:"suspend_at_limit"`
	SysInfoEnabled          string `json:"sysinfo" url:"sysinfo"`
	TasksMax                string `json:"TasksMax" url:"TasksMax"`
	WordpressEnabled        string `json:"wordpress" url:"wordpress"`
}

func (p *Package) translate() (pack rawPackage) {
	return rawPackage{
		AnonymousFtpEnabled:     reverseParseOnOff(p.AnonymousFtpEnabled, false),
		BandwidthQuota:          reverseParseNum(p.BandwidthQuota, false),
		CatchallEnabled:         reverseParseOnOff(p.AnonymousFtpEnabled, false),
		CgiEnabled:              reverseParseOnOff(p.CgiEnabled, false),
		CpuQuota:                reverseParseNum(p.CpuQuota, true),
		CronEnabled:             reverseParseOnOff(p.CronEnabled, false),
		DNSControlEnabled:       reverseParseOnOff(p.DNSControlEnabled, false),
		DomainPointerQuota:      reverseParseNum(p.DomainPointerQuota, false),
		DomainQuota:             reverseParseNum(p.DomainQuota, false),
		EmailAutoresponderQuota: reverseParseNum(p.EmailAutoresponderQuota, false),
		EmailForwarderQuota:     reverseParseNum(p.EmailForwarderQuota, false),
		EmailMailingListQuota:   reverseParseNum(p.EmailMailingListQuota, false),
		EmailQuota:              reverseParseNum(p.EmailQuota, false),
		FtpQuota:                reverseParseNum(p.FtpQuota, false),
		GitEnabled:              reverseParseOnOff(p.GitEnabled, false),
		IoReadBandwidthMax:      reverseParseNum(p.IoReadBandwidthMax, true),
		IoReadIopsMax:           reverseParseNum(p.IoReadIopsMax, true),
		IoWriteBandwidthMax:     reverseParseNum(p.IoWriteBandwidthMax, true),
		IoWriteIopsMax:          reverseParseNum(p.IoWriteIopsMax, true),
		InodeQuota:              reverseParseNum(p.InodeQuota, false),
		JailEnabled:             reverseParseOnOff(p.JailEnabled, false),
		Language:                p.Language,
		LoginKeysEnabled:        reverseParseOnOff(p.LoginKeysEnabled, false),
		MemoryHigh:              reverseParseNum(p.MemoryHigh, true),
		MemoryMax:               reverseParseNum(p.MemoryMax, true),
		MysqlQuota:              reverseParseNum(p.MysqlQuota, false),
		Name:                    p.Name,
		NginxEnabled:            reverseParseOnOff(p.NginxEnabled, false),
		PhpEnabled:              reverseParseOnOff(p.PhpEnabled, false),
		Quota:                   reverseParseNum(p.Quota, false),
		RedisEnabled:            reverseParseOnOff(p.RedisEnabled, false),
		SshEnabled:              reverseParseOnOff(p.SshEnabled, false),
		Skin:                    p.Skin,
		SpamAssassinEnabled:     reverseParseOnOff(p.SpamAssassinEnabled, false),
		SslEnabled:              reverseParseOnOff(p.SslEnabled, false),
		SubdomainQuota:          reverseParseNum(p.SubdomainQuota, false),
		SuspendAtLimitEnabled:   reverseParseOnOff(p.SuspendAtLimitEnabled, false),
		SysInfoEnabled:          reverseParseOnOff(p.SysInfoEnabled, false),
		TasksMax:                reverseParseNum(p.TasksMax, true),
		WordpressEnabled:        reverseParseOnOff(p.WordpressEnabled, false),
	}
}

func (p *rawPackage) translate() Package {
	return Package{
		AnonymousFtpEnabled:     parseOnOff(p.AnonymousFtpEnabled),
		BandwidthQuota:          parseNum(p.BandwidthQuota),
		CpuQuota:                parseNum(p.CpuQuota),
		CatchallEnabled:         parseOnOff(p.CatchallEnabled),
		CgiEnabled:              parseOnOff(p.CgiEnabled),
		CronEnabled:             parseOnOff(p.CronEnabled),
		DNSControlEnabled:       parseOnOff(p.DNSControlEnabled),
		DomainPointerQuota:      parseNum(p.DomainPointerQuota),
		DomainQuota:             parseNum(p.DomainQuota),
		EmailAutoresponderQuota: parseNum(p.EmailAutoresponderQuota),
		EmailForwarderQuota:     parseNum(p.EmailForwarderQuota),
		EmailMailingListQuota:   parseNum(p.EmailMailingListQuota),
		EmailQuota:              parseNum(p.EmailQuota),
		FtpQuota:                parseNum(p.FtpQuota),
		GitEnabled:              parseOnOff(p.GitEnabled),
		IoReadBandwidthMax:      parseNum(p.IoReadBandwidthMax),
		IoReadIopsMax:           parseNum(p.IoReadIopsMax),
		IoWriteBandwidthMax:     parseNum(p.IoWriteBandwidthMax),
		IoWriteIopsMax:          parseNum(p.IoWriteIopsMax),
		InodeQuota:              parseNum(p.InodeQuota),
		JailEnabled:             parseOnOff(p.JailEnabled),
		Language:                p.Language,
		LoginKeysEnabled:        parseOnOff(p.LoginKeysEnabled),
		MemoryHigh:              parseNum(p.MemoryHigh),
		MemoryMax:               parseNum(p.MemoryMax),
		MysqlQuota:              parseNum(p.MysqlQuota),
		Name:                    p.Name,
		NginxEnabled:            parseOnOff(p.NginxEnabled),
		PhpEnabled:              parseOnOff(p.PhpEnabled),
		Quota:                   parseNum(p.Quota),
		RedisEnabled:            parseOnOff(p.RedisEnabled),
		SshEnabled:              parseOnOff(p.SshEnabled),
		Skin:                    p.Skin,
		SpamAssassinEnabled:     parseOnOff(p.SpamAssassinEnabled),
		SslEnabled:              parseOnOff(p.SslEnabled),
		SubdomainQuota:          parseNum(p.SubdomainQuota),
		SuspendAtLimitEnabled:   parseOnOff(p.SuspendAtLimitEnabled),
		SysInfoEnabled:          parseOnOff(p.SysInfoEnabled),
		TasksMax:                parseNum(p.TasksMax),
		WordpressEnabled:        parseOnOff(p.WordpressEnabled),
	}
}
