package app

import (
	"encoding/json"
	"github.com/gorilla/mux"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
)

type server struct {
	hub    *hub
	router *mux.Router
	logger *logrus.Logger
}

func NewServer(hub *hub, logger *logrus.Logger) *server {
	s := &server{
		hub:    hub,
		router: mux.NewRouter(),
		logger: logger,
	}

	s.configureRouter()

	return s
}

func (s *server) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	s.router.ServeHTTP(w, r)
}

func (s *server) configureRouter() {
	s.router.HandleFunc("/ws", s.webSocketHandler()) // привязываем обработчик websocket соединения
	s.router.HandleFunc("/clients", s.handleClients()).Methods("GET")
}

func (s *server) webSocketHandler() http.HandlerFunc {
	// апгрейдер "апгрейдит" HTTP соединение до WS соединения
	var upgrader = websocket.Upgrader{
		ReadBufferSize:  1024,
		WriteBufferSize: 1024,
		CheckOrigin: func(r *http.Request) bool {
			return true
		},
	}

	return func(w http.ResponseWriter, r *http.Request) {
		conn, err := upgrader.Upgrade(w, r, nil)
		if err != nil {
			log.Fatal(err)
		}

		s.logger.Info("New connection")
		client := NewClient(s.hub, conn, s.logger) // инициализируем нового клиента на подключение
		s.hub.Register <- client                   // регисрируем клиента в центре управления

		// запускаем горутины обслуживающие ввод/вывод конкретного клиента
		go client.Read()
		go client.Write()
	}
}

func (s *server) handleClients() http.HandlerFunc {
	type localClient struct {
		Id string
	}

	return func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		var clients []localClient
		for _, client := range s.hub.Clients {
			lc := localClient{Id: client.Id}
			clients = append(clients, lc)
		}
		err := json.NewEncoder(w).Encode(clients)
		if err != nil {
			return
		}
	}
}
