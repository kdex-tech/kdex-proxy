package authz

import (
	"errors"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"kdex.dev/proxy/internal/authz"
)

type AuthorizerMock struct {
	authz.Authorizer
	CheckAccessFunc func(r *http.Request) error
}

func (a *AuthorizerMock) CheckAccess(r *http.Request) error {
	return a.CheckAccessFunc(r)
}

func TestAuthzMiddleware_Authz(t *testing.T) {
	type fields struct {
		Authorizer authz.Authorizer
	}
	type args struct {
		next http.Handler
		req  *http.Request
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   int
	}{
		{
			name: "authorization failed",
			fields: fields{
				Authorizer: &AuthorizerMock{
					CheckAccessFunc: func(r *http.Request) error { return errors.New("no permission") },
				},
			},
			args: args{
				next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
				req: httptest.NewRequest("GET", "/", nil),
			},
			want: http.StatusForbidden,
		},
		{
			name: "authorization passed",
			fields: fields{
				Authorizer: &AuthorizerMock{
					CheckAccessFunc: func(r *http.Request) error { return nil },
				},
			},
			args: args{
				next: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
					w.WriteHeader(http.StatusOK)
				}),
				req: httptest.NewRequest("GET", "/", nil),
			},
			want: http.StatusOK,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &AuthzMiddleware{
				Authorizer: tt.fields.Authorizer,
			}
			handler := a.Authz(tt.args.next)
			recorder := httptest.NewRecorder()
			handler.ServeHTTP(recorder, tt.args.req)
			assert.Equal(t, tt.want, recorder.Code)
		})
	}
}
