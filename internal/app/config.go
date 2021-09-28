package app

type Config struct {
	LogLevel    string `toml:"log_level"`
	Port        string `toml:"port"`
	CookieKey   string `toml:"cookie_key"`
	DatabaseUrl string `toml:"database_url"`
}

func NewConfig() *Config {
	return &Config{}
}
