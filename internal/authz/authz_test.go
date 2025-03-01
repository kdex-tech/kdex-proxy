package authz

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"kdex.dev/proxy/internal/config"
	kctx "kdex.dev/proxy/internal/context"
)

type mockPermissionProvider struct {
	GetPermissionsFunc func(path string) ([]config.Permission, error)
}

func (m *mockPermissionProvider) GetPermissions(path string) ([]config.Permission, error) {
	return m.GetPermissionsFunc(path)
}

func Test_defaultAuthorizer_CheckAccess(t *testing.T) {
	tests := []struct {
		name    string
		checker Checker
		r       *http.Request
		err     error
	}{
		{
			name: "permission required but don't have roles",
			checker: Checker{
				PermissionProvider: &mockPermissionProvider{
					GetPermissionsFunc: func(path string) ([]config.Permission, error) {
						return []config.Permission{{Principal: "admin"}}, nil
					},
				},
			},
			r:   httptest.NewRequest("GET", "/", nil),
			err: ErrUnauthorized,
		},
		{
			name: "no permission are defined",
			checker: Checker{
				PermissionProvider: &mockPermissionProvider{
					GetPermissionsFunc: func(path string) ([]config.Permission, error) {
						return nil, ErrNoPermissions
					},
				},
			},
			r: func() *http.Request {
				r := httptest.NewRequest("GET", "/", nil)
				r = r.WithContext(context.WithValue(r.Context(), kctx.UserRolesKey, []string{"admin"}))
				return r
			}(),
			err: ErrUnauthorized,
		},
		{
			name: "permission required but user doesn't have correct role",
			checker: Checker{
				PermissionProvider: &mockPermissionProvider{
					GetPermissionsFunc: func(path string) ([]config.Permission, error) {
						return []config.Permission{{Principal: "admin"}}, nil
					},
				},
			},
			r: func() *http.Request {
				r := httptest.NewRequest("GET", "/", nil)
				r = r.WithContext(context.WithValue(r.Context(), kctx.UserRolesKey, []string{"user"}))
				return r
			}(),
			err: ErrUnauthorized,
		},
		{
			name: "permission required and user has correct role",
			checker: Checker{
				PermissionProvider: &mockPermissionProvider{
					GetPermissionsFunc: func(path string) ([]config.Permission, error) {
						return []config.Permission{{Principal: "admin", Action: "read"}}, nil
					},
				},
			},
			r: func() *http.Request {
				r := httptest.NewRequest("GET", "/", nil)
				r = r.WithContext(context.WithValue(r.Context(), kctx.UserRolesKey, []string{"admin"}))
				return r
			}(),
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &defaultAuthorizer{
				checker: tt.checker,
			}
			err := a.CheckAccess(tt.r)
			assert.Equal(t, tt.err, err)
		})
	}
}
