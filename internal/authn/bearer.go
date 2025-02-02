package authn

import (
	"fmt"
	"log"
	"net/http"
	"strings"
)

type BearerAuthValidator struct {
	AuthorizationHeader string
	Realm               string
	Scopes              []string
}

func (v *BearerAuthValidator) Validate(r *http.Request) *AuthChallenge {
	token, err := v.bearerAuth(r)

	if err != nil {
		return &AuthChallenge{
			Scheme: AuthScheme_Bearer,
			Attributes: map[string]string{
				"realm":  v.Realm,
				"scopes": strings.Join(v.Scopes, " "),
				"error":  err.Error(),
			},
		}
	}

	if token == "" {
		return &AuthChallenge{
			Scheme: AuthScheme_Bearer,
			Attributes: map[string]string{
				"realm": v.Realm,
			},
		}
	}

	// FIX ME: process the token
	log.Printf("Bearer token: %s", token)

	return nil
}
func (v *BearerAuthValidator) bearerAuth(r *http.Request) (token string, err error) {
	auth := r.Header.Get(v.AuthorizationHeader)
	if auth == "" {
		return "", nil
	}
	return parseBearerAuth(auth)
}

func parseBearerAuth(auth string) (token string, err error) {
	const prefix = AuthScheme_Bearer + " "
	if len(auth) < len(prefix) || !equalFold(auth[:len(prefix)], prefix) {
		return "", fmt.Errorf("invalid_token")
	}
	return auth[len(prefix):], nil
}
