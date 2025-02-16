package roles

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"kdex.dev/proxy/internal/authn"
	"kdex.dev/proxy/internal/authz"
	"kdex.dev/proxy/internal/config"
	"kdex.dev/proxy/internal/expression"
	"kdex.dev/proxy/internal/store/session"
)

func TestRolesMiddleware_InjectRoles(t *testing.T) {
	defaultConfig := config.DefaultConfig()
	fieldEvaluator := expression.NewFieldEvaluator(&defaultConfig)

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		rolesObject := r.Context().Value(authz.ContextUserRolesKey)
		if rolesObject == nil {
			w.WriteHeader(http.StatusOK)
			return
		}
		roles, ok := rolesObject.([]string)
		if !ok {
			roles = []string{}
		}
		if len(roles) > 0 {
			for _, role := range roles {
				w.Header().Add("X-User-Roles", role)
			}
		}
		w.WriteHeader(http.StatusOK)
	})

	type args struct {
		req *http.Request
	}
	tests := []struct {
		name      string
		session   *session.SessionData
		args      args
		want      int
		wantRoles []string
	}{
		{
			name: "no roles",
			session: &session.SessionData{
				Data: map[string]interface{}{},
			},
			args: args{
				req: httptest.NewRequest("GET", "/", nil),
			},
			want:      http.StatusOK,
			wantRoles: nil,
		},
		{
			name: "with roles",
			session: &session.SessionData{
				Data: map[string]interface{}{"roles": []string{"admin"}},
			},
			args: args{
				req: httptest.NewRequest("GET", "/", nil),
			},
			want:      http.StatusOK,
			wantRoles: []string{"admin"},
		},
		{
			name:    "not logged in",
			session: nil,
			args: args{
				req: httptest.NewRequest("GET", "/", nil),
			},
			want:      http.StatusOK,
			wantRoles: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &RolesMiddleware{
				FieldEvaluator: fieldEvaluator,
			}
			recorder := httptest.NewRecorder()
			request := tt.args.req
			request = request.WithContext(context.WithValue(request.Context(), authn.ContextUserKey, tt.session))
			handler := m.InjectRoles(next)
			handler.ServeHTTP(recorder, request)
			assert.Equal(t, tt.want, recorder.Code)
			assert.Equal(t, tt.wantRoles, recorder.Header().Values("X-User-Roles"))
		})
	}
}
