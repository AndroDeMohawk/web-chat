package service

import (
	"crypto/rand"
	"encoding/hex"
	"sync"
	"time"
)

type Session struct {
	UserID    int64
	Username  string
	ExpiresAt time.Time
}

type SessionManager struct {
	mu       sync.RWMutex
	sessions map[string]Session
}

func NewSessionManager() *SessionManager {
	return &SessionManager{
		sessions: make(map[string]Session),
	}
}

func (m *SessionManager) CreateSession(userId int64, username string) (string, error) {
	b := make([]byte, 32)
	if _, err := rand.Read(b); err != nil {
		return "", nil
	}
	token := hex.EncodeToString(b)
	m.mu.Lock()
	defer m.mu.Unlock()
	m.sessions[token] = Session{
		UserID:    userId,
		Username:  username,
		ExpiresAt: time.Now().Add(time.Hour * 24),
	}
	return token, nil

}

func (m *SessionManager) GetSession(token string) (Session, bool) {
	m.mu.RLock()
	defer m.mu.RUnlock()
	sess, exists := m.sessions[token]
	if !exists || time.Now().After(sess.ExpiresAt) {
		return Session{}, false
	}
	return sess, true
}
func (m *SessionManager) DeleteSession(token string) {
	m.mu.Lock()
	defer m.mu.Unlock()
	delete(m.sessions, token)
}
