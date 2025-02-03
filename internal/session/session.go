package session

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type SessionData struct {
	AccessToken string    `json:"access_token"`
	UserInfo    UserInfo  `json:"user_info"`
	CreatedAt   time.Time `json:"created_at"`
}

type UserInfo struct {
	Username string `json:"username"`
	Email    string `json:"email"`
	// Add other user fields you need
}

type SessionStore interface {
	Set(ctx context.Context, sessionID string, data SessionData) error
	Get(ctx context.Context, sessionID string) (*SessionData, error)
	Delete(ctx context.Context, sessionID string) error
}

func NewSessionStore(ctx context.Context, storeType string) (SessionStore, error) {
	switch storeType {
	case "memory":
		return NewMemorySessionStore(), nil
	}
	return nil, fmt.Errorf("invalid store type: %s", storeType)
}

func NewMemorySessionStore() SessionStore {
	return &memorySessionStore{
		sessions: make(map[string]SessionData),
	}
}

type memorySessionStore struct {
	sessions map[string]SessionData
	mu       sync.RWMutex
}

func (s *memorySessionStore) Set(ctx context.Context, sessionID string, data SessionData) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.sessions[sessionID] = data
	return nil
}

func (s *memorySessionStore) Get(ctx context.Context, sessionID string) (*SessionData, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, ok := s.sessions[sessionID]
	if !ok {
		return nil, fmt.Errorf("session not found")
	}
	return &data, nil
}

func (s *memorySessionStore) Delete(ctx context.Context, sessionID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.sessions, sessionID)
	return nil
}

func (s *memorySessionStore) Cleanup(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	// Implement cleanup logic here
}
