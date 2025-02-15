package authz

import (
	"context"
	"fmt"
	"net/http"

	"kdex.dev/proxy/internal/authn"
	"kdex.dev/proxy/internal/authz"
	"kdex.dev/proxy/internal/cel"
	"kdex.dev/proxy/internal/config"
	"kdex.dev/proxy/internal/store/session"
)

type RolesMiddleware struct {
	Config    *config.Config
	Evaluator *cel.Evaluator
}

func NewRolesMiddleware(config *config.Config) (*RolesMiddleware, error) {
	evaluator, err := cel.NewEvaluator()
	if err != nil {
		return nil, err
	}

	return &RolesMiddleware{
		Config:    config,
		Evaluator: evaluator,
	}, nil
}

func (m *RolesMiddleware) EvaluateRoles(data map[string]interface{}) ([]string, error) {
	result, err := m.Evaluator.Evaluate(m.Config.Authz.Roles.Expression, data)
	if err != nil {
		return nil, fmt.Errorf("failed to evaluate expression: %v", err)
	}

	typedResult, ok := result.([]interface{})
	if !ok {
		return nil, fmt.Errorf("expression must evaluate to a list of strings")
	}

	roles := make([]string, len(typedResult))
	for i, v := range typedResult {
		roles[i] = fmt.Sprint(v)
	}

	return roles, nil
}

func (m *RolesMiddleware) InjectRoles(next http.Handler) http.HandlerFunc {
	return func(w http.ResponseWriter, r *http.Request) {
		// Get session data from context
		sessionData, ok := r.Context().Value(authn.ContextUserKey).(*session.SessionData)
		if !ok {
			next.ServeHTTP(w, r)
			return
		}

		roles, err := m.EvaluateRoles(sessionData.Data)
		if err != nil {
			http.Error(w, err.Error(), http.StatusInternalServerError)
			return
		}

		// Add roles to context
		ctx := context.WithValue(r.Context(), authz.ContextUserRolesKey, roles)
		next.ServeHTTP(w, r.WithContext(ctx))
	}
}
