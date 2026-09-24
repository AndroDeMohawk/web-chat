package handler

import (
	"context"
	"net/http"

	"github.com/AndroDeMohawk/web-chat/internal/service"
)

type contextKey string

const (
	UserIDCtxKey   contextKey = "userID"
	UsernameCtxKey contextKey = "username"
)

const CookieName = "session_token"

type AuthMiddleware struct {
	sessionManager *service.SessionManager
}

func NewAuthMiddleware(sm *service.SessionManager) *AuthMiddleware {
	return &AuthMiddleware{sessionManager: sm}
}

func (m *AuthMiddleware) Middleware(next http.Handler) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		cookie, err := r.Cookie(CookieName)
		if err != nil {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		sess, ok := m.sessionManager.GetSession(cookie.Value)
		if !ok {
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}
		ctx := context.WithValue(r.Context(), UserIDCtxKey, sess.UserID)
		ctx = context.WithValue(ctx, UsernameCtxKey, sess.Username)

		next.ServeHTTP(w, r.WithContext(ctx))
	})
}

func GetUserIDFromContext(ctx context.Context) (int64, bool) {
	id, ok := ctx.Value(UserIDCtxKey).(int64)
	return id, ok
}

func GetUsernameFromContext(ctx context.Context) (string, bool) {
	username, ok := ctx.Value(UsernameCtxKey).(string)
	return username, ok
}
