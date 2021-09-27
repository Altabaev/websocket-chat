package main

import (
	"flag"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"log"
	"net/http"
	"wbchat/internal/app"
)

var addr = flag.String("addr", ":9090", "http service address")
var logger = logrus.New()

func main() {
	logger.Info("Server starting...")
	flag.Parse()

	hub := app.NewHub(logger) // инициализация центра управления
	go hub.Run()              // запускаем в отдельной горутине

	// привязываем обработчик и запускаем сервер
	http.HandleFunc("/ws", wsHandler(hub))
	logger.Info("Start to listen new connections")
	err := http.ListenAndServe(*addr, nil)
	if err != nil {
		log.Fatal("ListerAndServe: ", err)
	}
}

// возвращает обработчик веб сокет соединений
func wsHandler(hub *app.Hub) http.HandlerFunc {
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

		logger.Info("New connection")
		client := app.NewClient(hub, conn, logger) // инициализируем нового клиента на подключение
		hub.Register <- client                     // регисрируем клиента в центре управления

		// запускаем горутины обслуживающие ввод/вывод конкретного клиента
		go client.Read()
		go client.Write()
	}
}
