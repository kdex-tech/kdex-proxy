package authn

import (
	"fmt"
	"net/http"
	"strings"

	iauthn "kdex.dev/proxy/internal/authn"
)

type AuthnMiddleware struct {
	AuthenticateHeader string
	ProtectedPaths     []string
	AuthValidators     []iauthn.AuthValidator
}

func NewAuthnMiddleware(config *iauthn.AuthnConfig) *AuthnMiddleware {
	return &AuthnMiddleware{
		AuthenticateHeader: config.AuthenticateHeader,
		ProtectedPaths:     config.ProtectedPaths,
		AuthValidators:     config.AuthValidators,
	}
}

func (a *AuthnMiddleware) Authn(h http.Handler) http.HandlerFunc {
	if len(a.ProtectedPaths) == 0 {
		return h.ServeHTTP
	}

	return func(w http.ResponseWriter, r *http.Request) {
		challenges := a.IsProtected(r)
		if len(challenges) > 0 {
			for _, challenge := range challenges {
				attributesString := ""
				delimiter := " "
				for k, v := range challenge.Attributes {
					attributesString += fmt.Sprintf(`%s%s="%s"`, delimiter, k, v)
					delimiter = ", "
				}
				w.Header().Add(a.AuthenticateHeader, fmt.Sprintf("%s%s", challenge.Scheme, attributesString))
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
