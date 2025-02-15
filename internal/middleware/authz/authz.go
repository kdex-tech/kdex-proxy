package authz

import (
	"net/http"

	"kdex.dev/proxy/internal/authz"
)

type AuthzMiddleware struct {
	Authorizer authz.Authorizer
}

func (a *AuthzMiddleware) Authz(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		if err := a.Authorizer.CheckAccess(r); err != nil {
			http.Error(w, "Forbidden", http.StatusForbidden)
			return
		}
		next.ServeHTTP(w, r)
	}
}
