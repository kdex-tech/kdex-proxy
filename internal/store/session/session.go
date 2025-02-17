package session

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"time"

	"kdex.dev/proxy/internal/config"
)

var (
	ErrSessionNotFound = errors.New("session not found")
)

type SessionData struct {
	AccessToken  string                 `json:"access_token"`
	CreatedAt    time.Time              `json:"created_at"`
	RefreshToken string                 `json:"refresh_token"`
	Data         map[string]interface{} `json:"data"`
}

type SessionHelper struct {
	Config       *config.Config
	SessionStore *SessionStore
}

type SessionStore interface {
	Delete(ctx context.Context, sessionID string) error
	Get(ctx context.Context, sessionID string) (*SessionData, error)
	IsLoggedIn(ctx context.Context, sessionID string) (bool, error)
	Set(ctx context.Context, sessionID string, data SessionData) error
}

func NewSessionStore(ctx context.Context, config *config.SessionConfig) (SessionStore, error) {
	switch config.Store {
	case "memory":
		return NewMemorySessionStore(config), nil
	}
	return nil, fmt.Errorf("invalid store type: %s", config.Store)
}

func (sh *SessionHelper) GetSessionID(r *http.Request) string {
	sessionCookie, err := r.Cookie(sh.Config.Session.CookieName)
	if err != nil {
		return ""
	}
	return sessionCookie.Value
}

func (sh *SessionHelper) IsLoggedIn(r *http.Response) (bool, error) {
	if sh.SessionStore == nil {
		return false, nil
	}

	sessionId := sh.GetSessionID(r.Request)

	isLoggedIn, err := (*sh.SessionStore).IsLoggedIn(r.Request.Context(), sessionId)
	if err != nil {
		return false, err
	}

	return isLoggedIn, nil
}
