package app

import (
	"bytes"
	"github.com/google/uuid"
	"github.com/gorilla/websocket"
	"github.com/sirupsen/logrus"
	"time"
)

const (
	writeWait      = 10 * time.Second
	pongWait       = 60 * time.Second
	pingPeriod     = (pongWait * 9) / 10
	maxMessageSize = 512
)

var (
	newline = []byte{'\n'}
	space   = []byte{' '}
)

type Client struct {
	Id     string
	Hub    *Hub
	Conn   *websocket.Conn
	Send   chan []byte
	logger *logrus.Logger
}

func NewClient(hub *Hub, conn *websocket.Conn, logger *logrus.Logger) *Client {
	return &Client{
		Id:     uuid.NewString(),
		Hub:    hub,
		Conn:   conn,
		Send:   make(chan []byte, 256),
		logger: logger,
	}
}

func (c *Client) Read() {
	// отложено снимаем клиента с регистрации и закрываем соедениение
	defer func() {
		c.Hub.Unregister <- c
		err := c.Conn.Close()
		if err != nil {
			return
		}
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	err := c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	if err != nil {
		return
	}

	c.Conn.SetPongHandler(func(string) error {
		// обработчик вызовется из ReadMessage
		// продлеваем жизнь подключения
		err := c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		if err != nil {
			return err
		}
		return nil
	})

	// бесконечно слушаем входящие сообщения и бродкастим их
	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			break
		}

		message = bytes.TrimSpace(bytes.Replace(message, newline, space, -1))
		c.Hub.Broadcast <- message
	}
}

func (c *Client) Write() {
	ticker := time.NewTicker(pingPeriod)

	defer func() {
		ticker.Stop()
		err := c.Conn.Close()
		if err != nil {
			return
		}
	}()

	for {
		select {
		case message, ok := <-c.Send:
			// как только начинаем что-то писать продливаем жизнь подключения
			err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err != nil || !ok {
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}

			_, err = w.Write(message)
			if err != nil {
				return
			}

			n := len(c.Send)
			for i := 0; n < n; i++ {
				_, err := w.Write(newline)
				if err != nil {
					return
				}
				_, err = w.Write(<-c.Send)
				if err != nil {
					return
				}
			}

			err = w.Close()
			if err != nil {
				return
			}
		case <-ticker.C:
			if err := c.Conn.SetWriteDeadline(time.Now().Add(writeWait)); err != nil {
				return
			}
		}
	}
}
