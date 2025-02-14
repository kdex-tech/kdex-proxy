package authz

import (
	"context"
	"net/http"

	"kdex.dev/proxy/internal/authn"
	"kdex.dev/proxy/internal/authz"
	"kdex.dev/proxy/internal/store/session"
)

type RolesMiddleware struct{}

func NewRolesMiddleware() *RolesMiddleware {
	return &RolesMiddleware{}
}

func (m *RolesMiddleware) InjectRoles(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session data from context
		sessionData, ok := r.Context().Value(authn.ContextUserKey).(*session.SessionData)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		// Extract roles from session data
		var roles []string
		if rolesIface, ok := sessionData.Data["roles"].([]interface{}); ok {
			roles = make([]string, len(rolesIface))
			for i, role := range rolesIface {
				roles[i] = role.(string)
			}
		}

		// Add roles to context
		ctx := context.WithValue(r.Context(), authz.ContextUserRolesKey, roles)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
