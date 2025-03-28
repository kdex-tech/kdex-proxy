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

func NewSessionStore(config *config.Config) (SessionStore, error) {
	switch config.Session.Store {
	case "memory":
		return NewMemorySessionStore(&config.Session), nil
	}
	return nil, fmt.Errorf("invalid store type: %s", config.Session.Store)
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
