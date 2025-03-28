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
