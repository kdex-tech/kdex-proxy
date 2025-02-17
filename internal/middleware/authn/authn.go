package authn

import (
	"net/http"

	iauthn "kdex.dev/proxy/internal/authn"
)

type AuthnMiddleware struct {
	AuthenticateHeader     string
	AuthenticateStatusCode int
	AuthValidator          iauthn.AuthValidator
}

func (a *AuthnMiddleware) Authn(h http.Handler) http.HandlerFunc {
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
	challenge := a.AuthValidator.Validate(w, r)
	if challenge != nil {
		return challenge
	}
	return nil
}
