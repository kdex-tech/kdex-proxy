package authn

import (
	"encoding/base64"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	AuthType_Basic = "Basic"

	DefaultAuthenticateHeader  = "Proxy-Authenticate"
	DefaultAuthorizationHeader = "Proxy-Authorization"

	DefaultRealm = "KDEX Proxy"

	Validator_NoOp            = "noop"
	Validator_StaticBasicAuth = "static_basic_auth"
)

type AuthChallenge struct {
	Type  string
	Realm string
}

type AuthValidator interface {
	Validate(r *http.Request) *AuthChallenge
}

type AuthnConfig struct {
	AuthenticateHeader  string
	AuthorizationHeader string
	AuthValidator       AuthValidator
}

func NewAuthnConfigFromEnv() *AuthnConfig {
	authenticate_header := os.Getenv("AUTHENTICATE_HEADER")
	if authenticate_header == "" {
		authenticate_header = DefaultAuthenticateHeader
	}
	authorization_header := os.Getenv("AUTHORIZATION_HEADER")
	if authorization_header == "" {
		authorization_header = DefaultAuthorizationHeader
	}
	authn_validator_env := os.Getenv("AUTHN_VALIDATOR")
	if authn_validator_env == "" {
		authn_validator_env = Validator_NoOp
	}

	var authn_validator AuthValidator
	switch authn_validator_env {
	case Validator_StaticBasicAuth:
		realm := os.Getenv("STATIC_BASIC_AUTH_REALM")
		if realm == "" {
			realm = DefaultRealm
		}
		username := os.Getenv("STATIC_BASIC_AUTH_USERNAME")
		password := os.Getenv("STATIC_BASIC_AUTH_PASSWORD")
		if username == "" || password == "" {
			log.Fatalf("STATIC_BASIC_AUTH_USERNAME and STATIC_BASIC_AUTH_PASSWORD must be set when using %s validator", Validator_StaticBasicAuth)
		}
		authn_validator = &StaticBasicAuthValidator{
			AuthorizationHeader: authorization_header,
			Realm:               realm,
			Username:            username,
			Password:            password,
		}
	default: // Validator_NoOp
		authn_validator = &NoOpAuthValidator{}
	}

	return &AuthnConfig{
		AuthenticateHeader:  authenticate_header,
		AuthorizationHeader: authorization_header,
		AuthValidator:       authn_validator,
	}
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
	const prefix = "Basic "
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
