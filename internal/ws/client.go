package ws

import (
	"log"
	"net/http"
	"time"

	"github.com/gorilla/websocket"
)

const (
	// Время на запись сообщения клиенту
	writeWait = 10 * time.Second

	// Время ожидания следующего pong-сообщения от клиента
	pongWait = 60 * time.Second

	// Интервал отправки ping-сообщений (должен быть меньше pongWait)
	pingPeriod = (pongWait * 9) / 10

	// Максимальный размер сообщения от клиента (10 KB)
	maxMessageSize = 10240
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	// Проверка CORS (в продакшене лучше настроить строго, сейчас разрешаем всё)
	CheckOrigin: func(r *http.Request) bool {
		return true
	},
}

// Client — промежуточное звено между WS соединением и Hub
type Client struct {
	Hub *Hub

	// WebSocket соединение
	Conn *websocket.Conn

	// Буферизованный канал отправляемых сообщений
	send chan []byte

	// Данные авторизованного пользователя
	UserID   int64
	Username string
}

func NewClient(hub *Hub, conn *websocket.Conn, userID int64, username string) *Client {
	return &Client{
		Hub:      hub,
		Conn:     conn,
		send:     make(chan []byte, 256),
		UserID:   userID,
		Username: username,
	}
}

// readPump читает входящие сообщения из WebSocket
func (c *Client) ReadPump(onMessage func(client *Client, msg []byte)) {
	defer func() {
		c.Hub.unregister <- c
		c.Conn.Close()
	}()

	c.Conn.SetReadLimit(maxMessageSize)
	c.Conn.SetReadDeadline(time.Now().Add(pongWait))
	c.Conn.SetPongHandler(func(string) error {
		c.Conn.SetReadDeadline(time.Now().Add(pongWait))
		return nil
	})

	for {
		_, message, err := c.Conn.ReadMessage()
		if err != nil {
			if websocket.IsUnexpectedCloseError(err, websocket.CloseGoingAway, websocket.CloseAbnormalClosure) {
				log.Printf("error WS read: %v", err)
			}
			break
		}

		// Вызываем внешнюю логику обработки входящего сообщения
		if onMessage != nil {
			onMessage(c, message)
		}
	}
}

// writePump отправляет сообщения из канала send в WebSocket
func (c *Client) WritePump() {
	ticker := time.NewTicker(pingPeriod)
	defer func() {
		ticker.Stop()
		c.Conn.Close()
	}()

	for {
		select {
		case message, ok := <-c.send:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if !ok {
				// Hub закрыл канал
				c.Conn.WriteMessage(websocket.CloseMessage, []byte{})
				return
			}

			w, err := c.Conn.NextWriter(websocket.TextMessage)
			if err != nil {
				return
			}
			w.Write(message)

			// Отправляем скомпонованные сообщения из буфера (если есть)
			n := len(c.send)
			for i := 0; i < n; i++ {
				w.Write(<-c.send)
			}

			if err := w.Close(); err != nil {
				return
			}

		case <-ticker.C:
			c.Conn.SetWriteDeadline(time.Now().Add(writeWait))
			if err := c.Conn.WriteMessage(websocket.PingMessage, nil); err != nil {
				return
			}
		}
	}
}
