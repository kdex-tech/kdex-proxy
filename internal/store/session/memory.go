package session

import (
	"context"
	"fmt"
	"sync"
	"time"

	"kdex.dev/proxy/internal/config"
	"kdex.dev/proxy/internal/util"
)

func NewMemorySessionStore(config *config.SessionConfig) SessionStore {
	return &memorySessionStore{
		config:   config,
		sessions: make(map[string]SessionData),
	}
}

type memorySessionStore struct {
	config   *config.SessionConfig
	sessions map[string]SessionData
	mu       sync.RWMutex
}

func (s *memorySessionStore) IsLoggedIn(ctx context.Context, sessionID string) (bool, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	sessionData, ok := s.sessions[sessionID]
	if ok {
		exp := util.TimeFromFloat64Seconds(sessionData.Data["exp"].(float64))
		if exp.Before(time.Now()) {
			delete(s.sessions, sessionID)
			return false, nil
		}
		return true, nil
	}
	return false, nil
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
	s.sessions = make(map[string]SessionData)
}
