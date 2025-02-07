package session

import (
	"context"
	"fmt"
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
