package handler

import (
	"net/http"
	"time"

	"github.com/AndroDeMohawk/web-chat/internal/service"
)

type AuthHandler struct {
	authService    *service.AuthService
	sessionManager *service.SessionManager
}

func NewAuthHandler(authService *service.AuthService, sessionManager *service.SessionManager) *AuthHandler {
	return &AuthHandler{authService: authService, sessionManager: sessionManager}
}

func (h *AuthHandler) Register(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	user, err := h.authService.Register(r.Context(), username, password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.setSessionAndRedirect(w, r, user.ID, user.Username)
}
func (h *AuthHandler) Login(w http.ResponseWriter, r *http.Request) {
	if r.Method != "POST" {
		http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
		return
	}
	username := r.FormValue("username")
	password := r.FormValue("password")
	user, err := h.authService.Login(r.Context(), username, password)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	h.setSessionAndRedirect(w, r, user.ID, user.Username)
}
func (h *AuthHandler) Logout(w http.ResponseWriter, r *http.Request) {
	cookie, err := r.Cookie(CookieName)
	if err != nil {
		h.sessionManager.DeleteSession(cookie.Value)
	}
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    "",
		Path:     "/",
		Expires:  time.Now().Add(-1 * time.Hour),
		HttpOnly: true,
	})
	w.Header().Set("HX-Redirect", "/")
	http.Redirect(w, r, "/", http.StatusFound)
}

func (h *AuthHandler) setSessionAndRedirect(w http.ResponseWriter, r *http.Request, userId int64, username string) {
	token, err := h.sessionManager.CreateSession(userId, username)
	if err != nil {
		http.Error(w, err.Error(), http.StatusInternalServerError)
		return
	}
	http.SetCookie(w, &http.Cookie{
		Name:     CookieName,
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(time.Hour * 24),
		HttpOnly: true,
		SameSite: http.SameSiteLaxMode,
	})
	w.Header().Set("HX-Redirect", "/")
	http.Redirect(w, r, "/", http.StatusSeeOther)
}
