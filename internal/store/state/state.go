package state

import (
	"context"
	"fmt"

	"kdex.dev/proxy/internal/config"
)

type StateStore interface {
	Set(ctx context.Context, state string) error
	Get(ctx context.Context, state string) (string, error)
	Delete(ctx context.Context, state string) error
}

func NewStateStore(config *config.Config) (StateStore, error) {
	switch config.State.Type {
	case "memory":
		return NewMemoryStateStore(config.State.TTL), nil
	}
	return nil, fmt.Errorf("invalid store type: %s", config.State.Type)
}
