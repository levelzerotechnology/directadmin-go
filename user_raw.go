package directadmin

import (
	"golang.org/x/text/language"
	"strings"
	"time"

	"github.com/spf13/cast"
	"golang.org/x/text/cases"
)

type rawUserConfig struct {
	Aftp                           string `json:"aftp"`
	APIWithPassword                string `json:"api_with_password"`
	Bandwidth                      string `json:"bandwidth"`
	Catchall                       string `json:"catchall"`
	Cgi                            string `json:"cgi"`
	Clamav                         string `json:"clamav"`
	Comments                       string `json:"comments"`
	Creator                        string `json:"creator"`
	Cron                           string `json:"cron"`
	DateCreated                    string `json:"date_created"`
	Dnscontrol                     string `json:"dnscontrol"`
	Domain                         string `json:"domain"`
	Email                          string `json:"email"`
	Ftp                            string `json:"ftp"`
	Git                            string `json:"git"`
	IP                             string `json:"ip"`
	Ips                            string `json:"ips"`
	Jail                           string `json:"jail"`
	Language                       string `json:"language"`
	LoginKeys                      string `json:"login_keys"`
	Name                           string `json:"name"`
	NginxUnit                      string `json:"nginx_unit"`
	NotifyOnAllQuestionFailures    string `json:"notify_on_all_question_failures"`
	NotifyOnAllTwostepAuthFailures string `json:"notify_on_all_twostep_auth_failures"`
	Ns1                            string `json:"ns1"`
	Ns2                            string `json:"ns2"`
	Package                        string `json:"package"`
	Php                            string `json:"php"`
	Redis                          string `json:"redis"`
	SecurityQuestions              string `json:"security_questions"`
	Skin                           string `json:"skin"`
	Spam                           string `json:"spam"`
	SSH                            string `json:"ssh"`
	Ssl                            string `json:"ssl"`
	SuspendAtLimit                 string `json:"suspend_at_limit"`
	Suspended                      string `json:"suspended"`
	Sysinfo                        string `json:"sysinfo"`
	TwostepAuth                    string `json:"twostep_auth"`
	UserEmail                      string `json:"user_email"`
	Username                       string `json:"username"`
	Usertype                       string `json:"usertype"`
	Wordpress                      string `json:"wordpress"`
}

func (r *rawUserConfig) parse() (UserConfig, error) {
	created, err := time.Parse("Mon Jan 02 15:04:05 2006", r.DateCreated)
	if err != nil {
		// TODO: figure out why this happens with some DA installations
		// real error received in testing: parsing time \"Mon Dec  2 03:18:43 2019\" as \"Mon Jan 02 15:04:05 2006\": cannot parse \"2 03:18:43 2019\" as \"02\"
		//return UserConfig{}, fmt.Errorf("failed to parse date created: %v", err)
	}

	return UserConfig{
		AftpEnabled:                    parseOnOff(r.Aftp),
		ApiAccessWithPasswordEnabled:   parseOnOff(r.APIWithPassword),
		CatchAllEnabled:                parseOnOff(r.Catchall),
		CgiEnabled:                     parseOnOff(r.Cgi),
		ClamAvEnabled:                  parseOnOff(r.Clamav),
		Created:                        created,
		Creator:                        r.Creator,
		CronEnabled:                    parseOnOff(r.Cron),
		Domain:                         r.Domain,
		DnsEnabled:                     parseOnOff(r.Dnscontrol),
		Email:                          r.Email,
		GitEnabled:                     parseOnOff(r.Git),
		IpAddresses:                    strings.Split(r.Ips, ", "),
		JailEnabled:                    parseOnOff(r.Jail),
		Language:                       r.Language,
		LoginKeysEnabled:               parseOnOff(r.LoginKeys),
		NginxEnabled:                   parseOnOff(r.NginxUnit),
		NotifyOnQuestionFailures:       parseOnOff(r.NotifyOnAllQuestionFailures),
		NotifyOnTwoFactorFailures:      parseOnOff(r.NotifyOnAllTwostepAuthFailures),
		Ns1:                            r.Ns1,
		Ns2:                            r.Ns2,
		Package:                        r.Package,
		PhpEnabled:                     parseOnOff(r.Php),
		RedisEnabled:                   parseOnOff(r.Redis),
		SecurityQuestionsEnabled:       parseOnOff(r.SecurityQuestions),
		Skin:                           r.Skin,
		SpamEnabled:                    parseOnOff(r.Spam),
		SshEnabled:                     parseOnOff(r.SSH),
		SslEnabled:                     parseOnOff(r.Ssl),
		Suspended:                      parseOnOff(r.Suspended),
		SuspendAtLimitEnabled:          parseOnOff(r.SuspendAtLimit),
		SysInfoEnabled:                 parseOnOff(r.Sysinfo),
		TwoFactorAuthenticationEnabled: parseOnOff(r.TwostepAuth),
		Username:                       r.Username,
		UserType:                       cases.Title(language.Und).String(r.Usertype),
		WordPressToolkitEnabled:        parseOnOff(r.Wordpress),
	}, nil
}

type rawApiStat struct {
	MaxUsage string `json:"max_usage"`
	Setting  string `json:"setting"`
	Usage    any    `json:"usage"` // this must be an any/interface type due to DA returning inconsistent data types
}

type rawDomainUsage struct {
	Domain    string `json:"domain"`
	Bandwidth struct {
		Limit string `json:"limit"`
		Usage string `json:"usage"`
	} `json:"bandwidth"`
	DiskUsage      any    `json:"quota"`
	LogUsage       string `json:"log_usage"`
	SubdomainUsage string `json:"nsubdomains"`
	Suspended      string `json:"suspended"`
	Settings       struct {
		Cgi        string   `json:"cgi"`
		Php        string   `json:"php"`
		Ssl        string   `json:"ssl"`
		Subdomains []string `json:"subdomains"`
	} `json:"settings"`
}

type domainUsage struct {
	BandwidthUsage int      `json:"bandwidthUsage" yaml:"bandwidthUsage"`
	CgiEnabled     bool     `json:"cgiEnabled" yaml:"cgiEnabled"`
	DiskUsage      int      `json:"diskUsage" yaml:"diskUsage"`
	Domain         string   `json:"domain" yaml:"domain"`
	LogUsage       int      `json:"logUsage" yaml:"logUsage"`
	PhpEnabled     bool     `json:"phpEnabled" yaml:"phpEnabled"`
	SslEnabled     bool     `json:"sslEnabled" yaml:"sslEnabled"`
	Subdomains     []string `json:"subdomains" yaml:"subdomains"`
	SubdomainUsage int      `json:"subdomainUsage" yaml:"subdomainUsage"`
	Suspended      bool     `json:"suspended" yaml:"suspended"`
}

// rawUserUsage maps directly to the raw JSON result returned by CMD_API_SHOW_USER_USAGE. We don't map all fields, as
// the stats map contains several fields with additional data
type rawUserUsage struct {
	Domains map[string]rawDomainUsage `json:"domains"`
	Stats   map[string]rawApiStat     `json:"stats"`
}

func (r *rawUserUsage) parse() UserUsage {
	parsedStats := make(map[string]rawApiStat)

	// allow us to access stats based on their setting name
	for _, stat := range r.Stats {
		switch stat.Setting {
		case "language", "send_usage_message", "":
		default:
			parsedStats[stat.Setting] = rawApiStat{
				Usage:    stat.Usage,
				MaxUsage: stat.MaxUsage,
			}
		}
	}

	r.Stats = parsedStats

	userUsage := UserUsage{
		BandwidthQuota:        cast.ToInt(r.Stats["bandwidth"].MaxUsage),
		BandwidthUsage:        cast.ToInt(r.Stats["bandwidth"].Usage),
		DbQuota:               cast.ToInt(r.Stats["mysql"].MaxUsage),
		DbUsage:               cast.ToInt(r.Stats["mysql"].Usage),
		DiskQuota:             cast.ToInt(r.Stats["quota"].MaxUsage),
		DiskUsage:             cast.ToInt(r.Stats["quota"].Usage),
		DomainPointersQuota:   cast.ToInt(r.Stats["domainptr"].MaxUsage),
		DomainPointersUsage:   cast.ToInt(r.Stats["domainptr"].Usage),
		DomainQuota:           cast.ToInt(r.Stats["vdomains"].MaxUsage),
		DomainUsage:           cast.ToInt(r.Stats["vdomains"].Usage),
		EmailQuota:            cast.ToInt(r.Stats["nemails"].MaxUsage),
		EmailUsage:            cast.ToInt(r.Stats["nemails"].Usage),
		EmailForwardersQuota:  cast.ToInt(r.Stats["nemailf"].MaxUsage),
		EmailForwardersUsage:  cast.ToInt(r.Stats["nemailf"].Usage),
		EmailMailingListQuota: cast.ToInt(r.Stats["nemailml"].MaxUsage),
		EmailMailingListUsage: cast.ToInt(r.Stats["nemailml"].Usage),
		FtpQuota:              cast.ToInt(r.Stats["ftp"].MaxUsage),
		FtpUsage:              cast.ToInt(r.Stats["ftp"].Usage),
		InodeQuota:            cast.ToInt(r.Stats["inode"].MaxUsage),
		InodeUsage:            cast.ToInt(r.Stats["inode"].Usage),
		SubdomainQuota:        cast.ToInt(r.Stats["nsubdomains"].MaxUsage),
		SubdomainUsage:        cast.ToInt(r.Stats["nsubdomains"].Usage),
	}

	for id, domain := range r.Domains {
		if id != "info" {
			var diskUsage int

			// this is necessary because DA either returns a string, or a map with the disk usage in a "usage" field depending on which usage endpoint we hit
			switch domain.DiskUsage.(type) {
			case string:
				diskUsage = cast.ToInt(domain.DiskUsage)
			case map[string]any:
				diskUsageRaw := domain.DiskUsage.(map[string]any)
				diskUsage = cast.ToInt(diskUsageRaw["usage"])
			}

			userUsage.Domains = append(userUsage.Domains, domainUsage{
				BandwidthUsage: cast.ToInt(domain.Bandwidth.Usage),
				CgiEnabled:     parseOnOff(domain.Settings.Cgi),
				DiskUsage:      diskUsage,
				Domain:         domain.Domain,
				LogUsage:       cast.ToInt(domain.LogUsage),
				PhpEnabled:     parseOnOff(domain.Settings.Php),
				SslEnabled:     parseOnOff(domain.Settings.Ssl),
				Subdomains:     domain.Settings.Subdomains,
				SubdomainUsage: cast.ToInt(domain.SubdomainUsage),
				Suspended:      parseOnOff(domain.Suspended),
			})
		}
	}

	return userUsage
}
