package authn

import (
	"encoding/base64"
	"net/http"
	"strings"
)

type StaticBasicAuthValidator struct {
	AuthorizationHeader string
	Realm               string
	Username            string
	Password            string
}

func (v *StaticBasicAuthValidator) Register(mux *http.ServeMux) {
	// noop
}

func (v *StaticBasicAuthValidator) Validate(w http.ResponseWriter, r *http.Request) (*AuthChallenge, any) {
	username, password, ok := v.basicAuth(r)
	if !ok || username != v.Username || password != v.Password {
		return &AuthChallenge{
			Scheme: AuthScheme_Basic,
			Attributes: map[string]string{
				"realm": v.Realm,
			},
		}, nil
	}
	return nil, nil
}

func (v *StaticBasicAuthValidator) basicAuth(r *http.Request) (username, password string, ok bool) {
	auth := r.Header.Get(v.AuthorizationHeader)
	if auth == "" {
		return "", "", false
	}
	return parseBasicAuth(auth)
}

func equalFold(s, t string) bool {
	if len(s) != len(t) {
		return false
	}
	for i := 0; i < len(s); i++ {
		if toLower(s[i]) != toLower(t[i]) {
			return false
		}
	}
	return true
}

func toLower(c byte) byte {
	if c >= 'A' && c <= 'Z' {
		return c + ('a' - 'A')
	}
	return c
}

func parseBasicAuth(auth string) (username, password string, ok bool) {
	const prefix = AuthScheme_Basic + " "
	if len(auth) < len(prefix) || !equalFold(auth[:len(prefix)], prefix) {
		return "", "", false
	}
	c, err := base64.StdEncoding.DecodeString(auth[len(prefix):])
	if err != nil {
		return "", "", false
	}
	cs := string(c)
	username, password, ok = strings.Cut(cs, ":")
	if !ok {
		return "", "", false
	}
	return username, password, true
}
