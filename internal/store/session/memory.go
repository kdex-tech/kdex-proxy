// Copyright 2025 KDex Tech
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package session

import (
	"context"
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
		return nil, ErrSessionNotFound
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
