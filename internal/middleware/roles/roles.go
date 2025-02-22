package roles

import (
	"context"
	"log"
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
		var roles []string

		sessionData, ok := r.Context().Value(authn.ContextUserKey).(*session.SessionData)
		if ok && sessionData != nil {
			var err error
			roles, err = m.FieldEvaluator.EvaluateRoles(sessionData.Data)
			if err != nil {
				log.Printf("failed to evaluate roles: %v", err)
				roles = []string{"anonymous"}
			}
		} else {
			roles = []string{"anonymous"}
		}

		// Add roles to context
		ctx := context.WithValue(r.Context(), authz.ContextUserRolesKey, roles)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
