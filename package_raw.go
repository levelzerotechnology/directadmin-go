package directadmin

type rawPackage struct {
	AnonymousFtpEnabled     string `json:"aftp" url:"aftp"`
	BandwidthQuota          string `json:"bandwidth" url:"bandwidth"`
	CPUQuota                string `json:"CPUQuota" url:"CPUQuota"`
	CatchallEnabled         string `json:"catchall" url:"catchall"`
	CGIEnabled              string `json:"cgi" url:"cgi"`
	CronEnabled             string `json:"cron" url:"cron"`
	DNSControlEnabled       string `json:"dnscontrol" url:"dnscontrol"`
	DomainPointerQuota      string `json:"domainptr" url:"domainptr"`
	DomainQuota             string `json:"vdomains" url:"vdomains"`
	EmailAutoresponderQuota string `json:"nemailr" url:"nemailr"`
	EmailForwarderQuota     string `json:"nemailf" url:"nemailf"`
	EmailMailingListQuota   string `json:"nemailml" url:"nemailml"`
	EmailQuota              string `json:"nemails" url:"nemails"`
	FTPQuota                string `json:"ftp" url:"ftp"`
	GitEnabled              string `json:"git" url:"git"`
	IOReadBandwidthMax      string `json:"IOReadBandwidthMax" url:"IOReadBandwidthMax"`
	IOReadIopsMax           string `json:"IOReadIOPSMax" url:"IOReadIOPSMax"`
	IOWriteBandwidthMax     string `json:"IOWriteBandwidthMax" url:"IOWriteBandwidthMax"`
	IOWriteIopsMax          string `json:"IOWriteIOPSMax" url:"IOWriteIOPSMax"`
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
	SSHEnabled              string `json:"ssh" url:"ssh"`
	Skin                    string `json:"skin" url:"skin"`
	SpamAssassinEnabled     string `json:"spam" url:"spam"`
	SSLEnabled              string `json:"ssl" url:"ssl"`
	SubdomainQuota          string `json:"nsubdomains" url:"nsubdomains"`
	SuspendAtLimitEnabled   string `json:"suspend_at_limit" url:"suspend_at_limit"`
	SysInfoEnabled          string `json:"sysinfo" url:"sysinfo"`
	TasksMax                string `json:"TasksMax" url:"TasksMax"`
	WordPressEnabled        string `json:"wordpress" url:"wordpress"`
}

func (p *Package) translate() (pack rawPackage) {
	return rawPackage{
		AnonymousFtpEnabled:     reverseParseOnOff(p.AnonymousFTPEnabled),
		BandwidthQuota:          reverseParseNum(p.BandwidthQuota, false),
		CatchallEnabled:         reverseParseOnOff(p.AnonymousFTPEnabled),
		CGIEnabled:              reverseParseOnOff(p.CGIEnabled),
		CPUQuota:                reverseParseNum(p.CPUQuota, true),
		CronEnabled:             reverseParseOnOff(p.CronEnabled),
		DNSControlEnabled:       reverseParseOnOff(p.DNSControlEnabled),
		DomainPointerQuota:      reverseParseNum(p.DomainPointerQuota, false),
		DomainQuota:             reverseParseNum(p.DomainQuota, false),
		EmailAutoresponderQuota: reverseParseNum(p.EmailAutoresponderQuota, false),
		EmailForwarderQuota:     reverseParseNum(p.EmailForwarderQuota, false),
		EmailMailingListQuota:   reverseParseNum(p.EmailMailingListQuota, false),
		EmailQuota:              reverseParseNum(p.EmailQuota, false),
		FTPQuota:                reverseParseNum(p.FTPQuota, false),
		GitEnabled:              reverseParseOnOff(p.GitEnabled),
		IOReadBandwidthMax:      reverseParseNum(p.IOReadBandwidthMax, true),
		IOReadIopsMax:           reverseParseNum(p.IOReadIopsMax, true),
		IOWriteBandwidthMax:     reverseParseNum(p.IOWriteBandwidthMax, true),
		IOWriteIopsMax:          reverseParseNum(p.IOWriteIopsMax, true),
		InodeQuota:              reverseParseNum(p.InodeQuota, false),
		JailEnabled:             reverseParseOnOff(p.JailEnabled),
		Language:                p.Language,
		LoginKeysEnabled:        reverseParseOnOff(p.LoginKeysEnabled),
		MemoryHigh:              reverseParseNum(p.MemoryHigh, true),
		MemoryMax:               reverseParseNum(p.MemoryMax, true),
		MysqlQuota:              reverseParseNum(p.MySQLQuota, false),
		Name:                    p.Name,
		NginxEnabled:            reverseParseOnOff(p.NginxEnabled),
		PhpEnabled:              reverseParseOnOff(p.PHPEnabled),
		Quota:                   reverseParseNum(p.Quota, false),
		RedisEnabled:            reverseParseOnOff(p.RedisEnabled),
		SSHEnabled:              reverseParseOnOff(p.SSHEnabled),
		Skin:                    p.Skin,
		SpamAssassinEnabled:     reverseParseOnOff(p.SpamAssassinEnabled),
		SSLEnabled:              reverseParseOnOff(p.SSLEnabled),
		SubdomainQuota:          reverseParseNum(p.SubdomainQuota, false),
		SuspendAtLimitEnabled:   reverseParseOnOff(p.SuspendAtLimitEnabled),
		SysInfoEnabled:          reverseParseOnOff(p.SysInfoEnabled),
		TasksMax:                reverseParseNum(p.TasksMax, true),
		WordPressEnabled:        reverseParseOnOff(p.WordPressEnabled),
	}
}

func (p *rawPackage) translate() Package {
	return Package{
		AnonymousFTPEnabled:     parseOnOff(p.AnonymousFtpEnabled),
		BandwidthQuota:          parseNum(p.BandwidthQuota),
		CPUQuota:                parseNum(p.CPUQuota),
		CatchallEnabled:         parseOnOff(p.CatchallEnabled),
		CGIEnabled:              parseOnOff(p.CGIEnabled),
		CronEnabled:             parseOnOff(p.CronEnabled),
		DNSControlEnabled:       parseOnOff(p.DNSControlEnabled),
		DomainPointerQuota:      parseNum(p.DomainPointerQuota),
		DomainQuota:             parseNum(p.DomainQuota),
		EmailAutoresponderQuota: parseNum(p.EmailAutoresponderQuota),
		EmailForwarderQuota:     parseNum(p.EmailForwarderQuota),
		EmailMailingListQuota:   parseNum(p.EmailMailingListQuota),
		EmailQuota:              parseNum(p.EmailQuota),
		FTPQuota:                parseNum(p.FTPQuota),
		GitEnabled:              parseOnOff(p.GitEnabled),
		IOReadBandwidthMax:      parseNum(p.IOReadBandwidthMax),
		IOReadIopsMax:           parseNum(p.IOReadIopsMax),
		IOWriteBandwidthMax:     parseNum(p.IOWriteBandwidthMax),
		IOWriteIopsMax:          parseNum(p.IOWriteIopsMax),
		InodeQuota:              parseNum(p.InodeQuota),
		JailEnabled:             parseOnOff(p.JailEnabled),
		Language:                p.Language,
		LoginKeysEnabled:        parseOnOff(p.LoginKeysEnabled),
		MemoryHigh:              parseNum(p.MemoryHigh),
		MemoryMax:               parseNum(p.MemoryMax),
		MySQLQuota:              parseNum(p.MysqlQuota),
		Name:                    p.Name,
		NginxEnabled:            parseOnOff(p.NginxEnabled),
		PHPEnabled:              parseOnOff(p.PhpEnabled),
		Quota:                   parseNum(p.Quota),
		RedisEnabled:            parseOnOff(p.RedisEnabled),
		SSHEnabled:              parseOnOff(p.SSHEnabled),
		Skin:                    p.Skin,
		SpamAssassinEnabled:     parseOnOff(p.SpamAssassinEnabled),
		SSLEnabled:              parseOnOff(p.SSLEnabled),
		SubdomainQuota:          parseNum(p.SubdomainQuota),
		SuspendAtLimitEnabled:   parseOnOff(p.SuspendAtLimitEnabled),
		SysInfoEnabled:          parseOnOff(p.SysInfoEnabled),
		TasksMax:                parseNum(p.TasksMax),
		WordPressEnabled:        parseOnOff(p.WordPressEnabled),
	}
}
