package authn

import (
	"net/http"
	"strings"

	iauthn "kdex.dev/proxy/internal/authn"
)

type AuthnMiddleware struct {
	AuthenticateHeader     string
	AuthenticateStatusCode int
	ProtectedPaths         []string
	AuthValidator          iauthn.AuthValidator
}

func NewAuthnMiddleware(config *iauthn.AuthnConfig) *AuthnMiddleware {
	return &AuthnMiddleware{
		AuthenticateHeader:     config.AuthenticateHeader,
		AuthenticateStatusCode: config.AuthenticateStatusCode,
		ProtectedPaths:         config.ProtectedPaths,
		AuthValidator:          config.AuthValidators[0],
	}
}

func (a *AuthnMiddleware) Authn(h http.Handler) http.HandlerFunc {
	if len(a.ProtectedPaths) == 0 {
		return h.ServeHTTP
	}

	return func(w http.ResponseWriter, r *http.Request) {
		challenge := a.IsProtected(w, r)

		if challenge != nil {
			challenge(h)
			return
		}

		h.ServeHTTP(w, r)
	}
}

func (a *AuthnMiddleware) IsProtected(w http.ResponseWriter, r *http.Request) func(h http.Handler) {
	path := r.URL.Path
	for _, p := range a.ProtectedPaths {
		if strings.HasPrefix(path, p) {
			challenge := a.AuthValidator.Validate(w, r)
			if challenge != nil {
				return challenge
			}
		}
	}
	return nil
}
