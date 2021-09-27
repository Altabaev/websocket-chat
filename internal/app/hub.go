package app

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
)

type Hub struct {
	Broadcast  chan []byte
	Register   chan *Client
	Unregister chan *Client
	Clients    map[string]*Client
	logger     *logrus.Logger
}

func NewHub(logger *logrus.Logger) *Hub {
	return &Hub{
		Clients:    make(map[string]*Client),
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
			h.Clients[client.Id] = client // регистрируем нового клиента в хеш-таблице
		case client := <-h.Unregister:
			// если такой клиент зарегистрирован, удаляем из хеш-таблицы и закрываем его канал
			if _, ok := h.Clients[client.Id]; ok {
				delete(h.Clients, client.Id)
				close(client.Send)
			}
		case data := <-h.Broadcast:
			message := &Message{}
			err := json.Unmarshal(data, message)
			if err != nil {
				logrus.Info(string(data))
				logrus.Error(err)
				return
			}

			// каждое новое сообщение рассылается каждому зарегистрированному клиенту
			for _, client := range h.Clients {
				select {
				case client.Send <- data:
				default:
					// если канал клиента недоступен, закрываем его и удаляем клиента из хеш-таблицы
					close(client.Send)
					delete(h.Clients, client.Id)
				}
			}
		}
	}
}
