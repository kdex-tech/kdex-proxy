package authn

import (
	"encoding/base64"
	"fmt"
	"log"
	"net/http"
	"strings"
)

type StaticBasicAuthValidator struct {
	AuthorizationHeader    string
	AuthenticateHeader     string
	AuthenticateStatusCode int
	Realm                  string
	Username               string
	Password               string
}

func (v *StaticBasicAuthValidator) Register(mux *http.ServeMux) {
	// noop
}

func (v *StaticBasicAuthValidator) Validate(w http.ResponseWriter, r *http.Request) func(h http.Handler) {
	username, password, ok := v.basicAuth(r)

	if !ok || username != v.Username || password != v.Password {
		return func(h http.Handler) {
			w.Header().Add(
				v.AuthenticateHeader,
				fmt.Sprintf(`%s realm="%s"`, AuthScheme_Basic, v.Realm),
			)
			log.Printf(
				"Sending %d Unauthorized, %s: %v",
				v.AuthenticateStatusCode,
				v.AuthenticateHeader,
				w.Header().Get(v.AuthenticateHeader),
			)
			http.Error(w, "Unauthorized", v.AuthenticateStatusCode)
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
