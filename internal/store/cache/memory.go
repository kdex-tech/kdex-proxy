package cache

import (
	"context"
	"sync"
	"time"
)

type memoryCacheStore struct {
	cache map[string]CacheEntry
	mu    sync.RWMutex
	ttl   time.Duration
}

func NewMemoryCacheStore(ttl time.Duration) CacheStore {
	store := &memoryCacheStore{
		cache: make(map[string]CacheEntry),
		ttl:   ttl,
	}

	// Start cleanup goroutine
	go func() {
		for now := range time.Tick(time.Second * 30) {
			store.mu.Lock()
			for key, entry := range store.cache {
				if now.Sub(entry.CreatedAt) > ttl {
					delete(store.cache, key)
				}
			}
			store.mu.Unlock()
		}
	}()

	return store
}

func (s *memoryCacheStore) Set(ctx context.Context, key string, entry CacheEntry) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.cache[key] = entry
	return nil
}

func (s *memoryCacheStore) Get(ctx context.Context, key string) (*CacheEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	entry, exists := s.cache[key]
	if !exists {
		return nil, nil
	}

	if time.Since(entry.CreatedAt) > s.ttl {
		delete(s.cache, key)
		return nil, nil
	}

	return &entry, nil
}

func (s *memoryCacheStore) Delete(ctx context.Context, key string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	delete(s.cache, key)
	return nil
}
