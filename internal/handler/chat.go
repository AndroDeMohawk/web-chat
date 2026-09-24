package handler

import (
	"bytes"
	"html/template"
	"net/http"

	"github.com/AndroDeMohawk/web-chat/internal/kafka"
	"github.com/AndroDeMohawk/web-chat/internal/repository"
	"github.com/AndroDeMohawk/web-chat/internal/ws"
)

type ChatHandler struct {
	templates *template.Template
	repo      *repository.Repository
	hub       *ws.Hub
}

func NewChatHandler(repo *repository.Repository, hub *ws.Hub) (*ChatHandler, error) {
	// Загружаем все HTML шаблоны
	tmpl, err := template.ParseGlob("templates/*.html")
	if err != nil {
		return nil, err
	}

	return &ChatHandler{
		templates: tmpl,
		repo:      repo,
		hub:       hub,
	}, nil
}

// RenderIndex выдает главную страницу (с чатом или формой входа)
func (h *ChatHandler) RenderIndex(w http.ResponseWriter, r *http.Request) {
	userID, ok := GetUserIDFromContext(r.Context())
	username, _ := GetUsernameFromContext(r.Context())

	data := map[string]interface{}{
		"IsAuthenticated": ok,
		"Username":        username,
		"UserID":          userID,
	}

	// Если пользователь авторизован — подгрузим последние сообщения из БД
	if ok {
		messages, err := h.repo.GetRecentMessages(r.Context(), 50)
		if err == nil {
			data["Messages"] = messages
		}
	}

	_ = h.templates.ExecuteTemplate(w, "index.html", data)
}

// RenderMessageHTML превращает объект сообщения из Kafka в HTML-строку для HTMX
func (h *ChatHandler) RenderMessageHTML(payload kafka.MessagePayload) ([]byte, error) {
	var buf bytes.Buffer
	err := h.templates.ExecuteTemplate(&buf, "message.html", payload)
	if err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}

func (h *ChatHandler) ClearChat(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}

	if err := h.repo.ClearAllMessages(r.Context()); err != nil {
		http.Error(w, "Failed to clear chat", http.StatusInternalServerError)
		return
	}

	// Рассылаем по сокету специальный HTML-сигнал для очистки блоков у всех
	clearHTML := `<div id="chat-messages" hx-swap-oob="innerHTML"></div>`
	h.hub.Broadcast([]byte(clearHTML))

	w.WriteHeader(http.StatusOK)
}
