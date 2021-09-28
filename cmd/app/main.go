package main

import (
	"flag"
	"github.com/BurntSushi/toml"
	"github.com/sirupsen/logrus"
	"log"
	"wbchat/internal/app"
)

var (
	configPath string
	logger     = logrus.New()
)

func init() {
	flag.StringVar(&configPath, "configs-path", "configs/app.toml", "path to app config file")
}

func main() {
	flag.Parse()

	config := app.NewConfig()
	if _, err := toml.DecodeFile(configPath, config); err != nil {
		log.Fatal("Config file not exists")
	}

	application := app.New(config, logger)
	application.Start()
}
