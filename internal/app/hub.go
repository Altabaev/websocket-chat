package app

import "github.com/sirupsen/logrus"

type Hub struct {
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	Clients    map[*Client]bool
	logger     *logrus.Logger
}

func NewHub(logger *logrus.Logger) *Hub {
	return &Hub{
		Clients:    make(map[*Client]bool),
		Broadcast:  make(chan []byte),
		Register:   make(chan *Client),
		Unregister: make(chan *Client),
		logger:     logger,
	}
}

func (h *Hub) Run() {
	h.logger.Info("Hub runned")
	for {
		select {
		case client := <-h.Register:
			h.Clients[client] = true // регистрируем нового клиента в хеш-таблице
		case client := <-h.Unregister:
			// если такой клиент зарегистрирован, удаляем из хеш-таблицы и закрываем его канал
			if _, ok := h.Clients[client]; ok {
				delete(h.Clients, client)
				close(client.Send)
			}
		case message := <-h.Broadcast:
			// каждое новое сообщение рассылается каждому зарегистрированному клиенту
			for client := range h.Clients {
				select {
				case client.Send <- message:
					h.logger.WithFields(logrus.Fields{
						"client": client,
						"message": message,
					}).Info("New message")
				default:
					// если канал клиента недоступен, закрываем его и удаляем клиента из хеш-таблицы
					close(client.Send)
					delete(h.Clients, client)
				}
			}
		}
	}
}
