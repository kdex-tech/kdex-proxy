package state

import (
	"context"
	"fmt"
	"time"
)

type StateStore interface {
	Set(ctx context.Context, state string) error
	Get(ctx context.Context, state string) (string, error)
	Delete(ctx context.Context, state string) error
}

func NewStateStore(ctx context.Context, storeType string) (StateStore, error) {
	switch storeType {
	case "memory":
		return NewMemoryStateStore(time.Minute * 2), nil
	}
	return nil, fmt.Errorf("invalid store type: %s", storeType)
}
