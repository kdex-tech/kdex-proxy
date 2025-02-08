package session

import (
	"context"
	"fmt"
	"time"
)

type SessionData struct {
	AccessToken  string                 `json:"access_token"`
	CreatedAt    time.Time              `json:"created_at"`
	RefreshToken string                 `json:"refresh_token"`
	Data         map[string]interface{} `json:"data"`
}

type SessionStore interface {
	Delete(ctx context.Context, sessionID string) error
	Get(ctx context.Context, sessionID string) (*SessionData, error)
	IsLoggedIn(ctx context.Context, sessionID string) (bool, error)
	Set(ctx context.Context, sessionID string, data SessionData) error
}

func NewSessionStore(ctx context.Context, storeType string) (SessionStore, error) {
	switch storeType {
	case "memory":
		return NewMemorySessionStore(), nil
	}
	return nil, fmt.Errorf("invalid store type: %s", storeType)
}
