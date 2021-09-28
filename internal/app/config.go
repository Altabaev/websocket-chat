package app

type Config struct {
	LogLevel    string `toml:"log_level"`
	Port        string `toml:"port"`
	DatabaseUrl string `toml:"database_url"`
}

func NewConfig() *Config {
	return &Config{}
}
