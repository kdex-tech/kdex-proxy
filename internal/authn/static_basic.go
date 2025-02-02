package authn

import (
	"net/http"
)

type StaticBasicAuthValidator struct {
	AuthorizationHeader string
	Realm               string
	Username            string
	Password            string
}

func (v *StaticBasicAuthValidator) Validate(r *http.Request) *AuthChallenge {
	username, password, ok := v.basicAuth(r)
	if !ok || username != v.Username || password != v.Password {
		return &AuthChallenge{
			Type:  AuthType_Basic,
			Realm: v.Realm,
		}
	}
	return nil
}

func (v *StaticBasicAuthValidator) basicAuth(r *http.Request) (username, password string, ok bool) {
	auth := r.Header.Get(v.AuthorizationHeader)
	if auth == "" {
		return "", "", false
	}
	return parseBasicAuth(auth)
}
