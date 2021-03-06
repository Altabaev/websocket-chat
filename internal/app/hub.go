package app

import (
	"encoding/json"
	"github.com/sirupsen/logrus"
	"wbchat/internal/models"
)

const (
	firstContactMessageType = "client_id"
)

type hub struct {
	Broadcast  chan []byte
	Register   chan *client
	Unregister chan *client
	Clients    map[string]*client
	logger     *logrus.Logger
}

func NewHub(logger *logrus.Logger) *hub {
	return &hub{
		Clients:    make(map[string]*client),
		Broadcast:  make(chan []byte),
		Register:   make(chan *client),
		Unregister: make(chan *client),
		logger:     logger,
	}
}

func (h *hub) Run() {
	for {
		select {
		case client := <-h.Register:
			h.Clients[client.Id] = client // регистрируем нового клиента в хеш-таблице
			// отправляем клиенту его идентификатор
			message := &models.Message{
				Type:     firstContactMessageType,
				ClientId: client.Id,
				Content:  client.Id,
			}
			data, err := json.Marshal(message)
			if err != nil {
				return
			}
			client.Send <- data
		case client := <-h.Unregister:
			// если такой клиент зарегистрирован, удаляем из хеш-таблицы и закрываем его канал
			if _, ok := h.Clients[client.Id]; ok {
				delete(h.Clients, client.Id)
				close(client.Send)
			}
		case data := <-h.Broadcast:
			message := &models.Message{}
			err := json.Unmarshal(data, message)
			if err != nil {
				logrus.Info(string(data))
				logrus.Error(err)
				return
			}

			// каждое новое сообщение рассылается каждому зарегистрированному клиенту
			for _, client := range h.Clients {
				if message.ClientId == client.Id {
					continue
				}
				select {
				case client.Send <- data:
					h.logger.Info("New message")
				default:
					// если канал клиента недоступен, закрываем его и удаляем клиента из хеш-таблицы
					close(client.Send)
					delete(h.Clients, client.Id)
				}
			}
		}
	}
}
