package authn

import (
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

	iauthn "kdex.dev/proxy/internal/authn"
)

const (
	ProtectedPathsEnvVar = "PROTECTED_PATHS"
)

type AuthnMiddleware struct {
	AuthenticateHeader string
	ProtectedPaths     []string
	AuthValidators     []iauthn.AuthValidator
}

func NewAuthnMiddlewareFromEnv() *AuthnMiddleware {
	var protected_paths []string

	protected_paths_env := os.Getenv(ProtectedPathsEnvVar)
	if protected_paths_env == "" {
		protected_paths = []string{}
	} else {
		protected_paths = strings.Split(protected_paths_env, ",")
	}
	log.Printf("Protected paths: %v", protected_paths)

	return &AuthnMiddleware{
		ProtectedPaths: protected_paths,
	}
}

func (a *AuthnMiddleware) WithAuthenticateHeader(header string) *AuthnMiddleware {
	a.AuthenticateHeader = header
	return a
}

func (a *AuthnMiddleware) WithValidator(validator iauthn.AuthValidator) *AuthnMiddleware {
	a.AuthValidators = append(a.AuthValidators, validator)
	return a
}

func (a *AuthnMiddleware) Authn(h http.Handler) http.HandlerFunc {
	if len(a.ProtectedPaths) == 0 {
		return h.ServeHTTP
	}

	return func(w http.ResponseWriter, r *http.Request) {
		challenges := a.IsProtected(r)
		if len(challenges) > 0 {
			for _, challenge := range challenges {
				w.Header().Add(a.AuthenticateHeader, fmt.Sprintf("%s realm=\"%s\"", challenge.Type, challenge.Realm))
			}
			http.Error(w, "Unauthorized", http.StatusUnauthorized)
			return
		}

		h.ServeHTTP(w, r)
	}
}

func (a *AuthnMiddleware) IsProtected(r *http.Request) []iauthn.AuthChallenge {
	path := r.URL.Path
	var challenges []iauthn.AuthChallenge
	for _, p := range a.ProtectedPaths {
		if strings.HasPrefix(path, p) {
			for _, validator := range a.AuthValidators {
				challenge := validator.Validate(r)
				if challenge != nil {
					challenges = append(challenges, *challenge)
				}
			}
			return challenges
		}
	}
	return nil
}
