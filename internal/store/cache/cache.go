package cache

import (
	"context"
	"time"

	"kdex.dev/proxy/internal/config"
)

type CacheEntry struct {
	Content     []byte
	ContentType string
	CreatedAt   time.Time
	ETag        string
	StatusCode  int
}

type CacheStore interface {
	Set(ctx context.Context, key string, entry CacheEntry) error
	Get(ctx context.Context, key string) (*CacheEntry, error)
	Delete(ctx context.Context, key string) error
}

// NewCacheStore creates a new cache store implementation
func NewCacheStore(config *config.Config) *CacheStore {
	if config.Proxy.Cache.Type == "memory" {
		store := NewMemoryCacheStore(config.Proxy.Cache.TTL)
		return &store
	}

	return nil
}
