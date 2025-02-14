package authz

import (
	"net/http"

	"kdex.dev/proxy/internal/authz"
	"kdex.dev/proxy/internal/config"
)

type AuthzMiddleware struct {
	Authorizer      authz.Authorizer
	PathPermissions map[string]config.Permission
}

func NewAuthzMiddleware(authorizer authz.Authorizer) *AuthzMiddleware {
	return &AuthzMiddleware{
		Authorizer: authorizer,
	}
}

func (a *AuthzMiddleware) WithPathPermission(path string, permission config.Permission) *AuthzMiddleware {
	a.PathPermissions[path] = permission
	return a
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
