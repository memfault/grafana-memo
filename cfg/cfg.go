package cfg

type Config struct {
	LogLevel string `toml:"log_level"`
	Slack    Slack
	Discord  Discord
	Grafana  Grafana
}

type Slack struct {
	Enabled  bool   `toml:"enabled"`
	BotToken string `toml:"bot_token"`
	AppToken string `toml:"app_token"`
}

type Discord struct {
	Enabled  bool   `toml:"enabled"`
	BotToken string `toml:"bot_token"`
}

type Grafana struct {
	ApiKey  string `toml:"api_key"`
	ApiUrl  string `toml:"api_url"`
	TLSKey  string `toml:"tls_key"`
	TLSCert string `toml:"tls_cert"`
}
