package roles

import (
	"context"
	"net/http"

	"kdex.dev/proxy/internal/authn"
	"kdex.dev/proxy/internal/authz"
	"kdex.dev/proxy/internal/expression"
	"kdex.dev/proxy/internal/store/session"
)

type RolesMiddleware struct {
	FieldEvaluator *expression.FieldEvaluator
}

func (m *RolesMiddleware) InjectRoles(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session data from context
		sessionData, ok := r.Context().Value(authn.ContextUserKey).(*session.SessionData)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		roles, err := m.FieldEvaluator.EvaluateRoles(sessionData.Data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Add roles to context
		ctx := context.WithValue(r.Context(), authz.ContextUserRolesKey, roles)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
