package config

type Config struct {
	Title      string
	Notion     Notion
	Proxy struct {
		On bool
		Mode string
		ZhimaProxy struct {
			Url string
		} `toml:"zhimaproxy"`
		FreeProxy struct {
			ProxyList  []FreeProxy `toml:"proxylist"`
		}`toml:"freeproxy"`

	}
	Mails      []Mail
	UserAgents []string `toml:"useragents"`
	FirstNames []string `toml:"first_names"`
	LastNames  []string `toml:"last_names"`
	Timezones []string `toml:"timezones"`
}

type Mail struct {
	Domain        string
	ReceiveServer string `toml:"receive_server"`
	ReceivePort   int    `toml:"receive_port"`
	Type          string
	Delimiter     string
	Data          []string
}

type FreeProxy struct {
	Url        string
	IpColumn   int `toml:"ip_column"`
	PortColumn int `toml:"port_column"`
}

type Notion struct {
	Host      string
	Endpoints struct {
		MailCheck         string `toml:"mail_check"`
		SendTmpPass       string `toml:"send_tmp_pass"`
		TmpPassLogin      string `toml:"tmp_pass_login"`
		ActivateRefer     string `toml:"activate_ref"`
		MailInvitation    string `toml:"mail_invite"`
		SubmitTransaction string `toml:"submit_transactions"`
	}
	Src []string
	Job []string
	Scope []string
}
