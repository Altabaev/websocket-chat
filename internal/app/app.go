package app

import (
	"github.com/gorilla/sessions"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
)

type app struct {
	config *Config
	logger *logrus.Logger
}

func New(config *Config, logger *logrus.Logger) *app {
	return &app{config, logger}
}

func (a *app) Start() {
	a.logger.Info("Application starting...")

	logLevel, err := logrus.ParseLevel(a.config.LogLevel)
	if err != nil {
		log.Fatal("Log level is wrong")
	}
	a.logger.SetLevel(logLevel)

	hub := NewHub(a.logger) // инициализация центра управления
	go hub.Run()            // запускаем в отдельной горутине

	server := NewServer(hub, a.logger, sessions.NewCookieStore([]byte(a.config.CookieKey)))

	a.logger.Infof("Application ready on http://localhost%s", a.config.Port)

	err = http.ListenAndServe(a.config.Port, server)
	if err != nil {
		log.Fatal("ListerAndServe: ", err)
	}
}
