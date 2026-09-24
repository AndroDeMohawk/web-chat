package handler

import (
	"context"
	"encoding/json"
	"log"
	"net/http"

	"github.com/AndroDeMohawk/web-chat/internal/kafka"
	"github.com/AndroDeMohawk/web-chat/internal/repository"
	"github.com/AndroDeMohawk/web-chat/internal/repository/db"
	"github.com/AndroDeMohawk/web-chat/internal/ws"
	"github.com/gorilla/websocket"
)

var upgrader = websocket.Upgrader{
	ReadBufferSize:  1024,
	WriteBufferSize: 1024,
	CheckOrigin: func(r *http.Request) bool {
		return true // Для локальной разработки
	},
}

type htmxWSIncoming struct {
	Content string `json:"content"`
}
type WSHandler struct {
	hub      *ws.Hub
	producer *kafka.Producer
	repo     *repository.Repository
}

func NewWSHandler(hub *ws.Hub, producer *kafka.Producer, repo *repository.Repository) *WSHandler {
	return &WSHandler{
		hub:      hub,
		producer: producer,
		repo:     repo,
	}
}

func (h *WSHandler) ServeWS(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	if !ok {
		http.Error(w, "Unauthorized", http.StatusUnauthorized)
		return
	}
	username, _ := GetUsernameFromContext(r.Context())

	conn, err := upgrader.Upgrade(w, r, nil)
	if err != nil {
		log.Printf("failed to upgrade connection: %v", err)
		return
	}

	client := ws.NewClient(h.hub, conn, userID, username)
	h.hub.RegisterClient(client)

	go client.WritePump()

	go client.ReadPump(func(c *ws.Client, msg []byte) {
		ctx := context.Background()

		// Парсим payload от HTMX ws-send
		var incoming htmxWSIncoming
		if err := json.Unmarshal(msg, &incoming); err != nil {
			log.Printf("failed to unmarshal htmx ws msg: %v", err)
			return
		}

		if incoming.Content == "" {
			return
		}

		// 1. Сохраняем сообщение в Postgres через sqlc
		savedMsg, err := h.repo.CreateMessage(ctx, db.CreateMessageParams{
			UserID:  c.UserID,
			Content: incoming.Content,
		})
		if err != nil {
			log.Printf("failed to save message to db: %v", err)
			return
		}

		// 2. Отправляем событие в Kafka
		payload := kafka.MessagePayload{
			ID:        savedMsg.ID,
			UserID:    c.UserID,
			Username:  c.Username,
			Content:   savedMsg.Content,
			CreatedAt: savedMsg.CreatedAt.Time,
		}

		if err := h.producer.SendMessage(ctx, payload); err != nil {
			log.Printf("failed to send message to kafka: %v", err)
		}
	})
}
