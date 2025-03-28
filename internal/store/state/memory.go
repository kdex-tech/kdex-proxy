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

package state

import (
	"context"
	"fmt"
	"sync"
	"time"
)

type State struct {
	state     string
	createdAt time.Time
}

func NewMemoryStateStore(ttl time.Duration) StateStore {
	store := &memoryStateStore{
		states: make(map[string]State),
		ttl:    ttl,
	}

	go func() {
		for now := range time.Tick(time.Second * 10) {
			store.mu.Lock()
			for key, state := range store.states {
				if now.Sub(state.createdAt) > ttl {
					delete(store.states, key)
				}
			}
			store.mu.Unlock()
		}
	}()

	return store
}

type memoryStateStore struct {
	states map[string]State
	mu     sync.RWMutex
	ttl    time.Duration
}

func (s *memoryStateStore) Set(ctx context.Context, state string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states[state] = State{
		state:     state,
		createdAt: time.Now(),
	}
	return nil
}

func (s *memoryStateStore) Get(ctx context.Context, state string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	data, ok := s.states[state]
	if !ok {
		return "", fmt.Errorf("state not found")
	}
	if time.Since(data.createdAt) > s.ttl {
		delete(s.states, state)
		return "", fmt.Errorf("state expired")
	}
	return data.state, nil
}

func (s *memoryStateStore) Delete(ctx context.Context, state string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.states, state)
	return nil
}

func (s *memoryStateStore) Cleanup(ctx context.Context) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.states = make(map[string]State)
}
