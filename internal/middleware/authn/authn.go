package authn

import (
	"context"
	"fmt"
	"log"
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
		ah := a.AuthenticateHeader
		challenge, data := a.IsProtected(w, r)

		if challenge != nil {
			attributesString := ""
			delimiter := " "
			for k, v := range challenge.Attributes {
				attributesString += fmt.Sprintf(`%s%s="%s"`, delimiter, k, v)
				delimiter = ", "
			}
			w.Header().Add(ah, fmt.Sprintf("%s%s", challenge.Scheme, attributesString))

			log.Printf("Sending %d Unauthorized, %s: %v", a.AuthenticateStatusCode, ah, w.Header().Get(ah))
			http.Error(w, "Unauthorized", a.AuthenticateStatusCode)
			return
		}

		if data != nil {
			r = r.WithContext(context.WithValue(r.Context(), ContextUserDataKey, data))
		}

		h.ServeHTTP(w, r)
	}
}

func (a *AuthnMiddleware) IsProtected(w http.ResponseWriter, r *http.Request) (*iauthn.AuthChallenge, any) {
	path := r.URL.Path
	for _, p := range a.ProtectedPaths {
		if strings.HasPrefix(path, p) {
			challenge, datum := a.AuthValidator.Validate(w, r)
			if challenge != nil {
				return challenge, datum
			}
		}
	}
	return nil, nil
}

type ContextKey string

const ContextUserDataKey ContextKey = "user_data"
