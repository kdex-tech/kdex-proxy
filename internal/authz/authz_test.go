package authz

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"kdex.dev/proxy/internal/config"
)

type mockPermissionProvider struct {
	GetPermissionsFunc func(path string) ([]config.Permission, error)
}

func (m *mockPermissionProvider) GetPermissions(path string) ([]config.Permission, error) {
	return m.GetPermissionsFunc(path)
}

func Test_defaultAuthorizer_CheckAccess(t *testing.T) {
	tests := []struct {
		name               string
		permissionProvider PermissionProvider
		r                  *http.Request
		err                error
	}{
		{
			name: "no roles",
			permissionProvider: &mockPermissionProvider{
				GetPermissionsFunc: func(path string) ([]config.Permission, error) {
					return nil, nil
				},
			},
			r:   httptest.NewRequest("GET", "/", nil),
			err: ErrNoRoles,
		},
		{
			name: "no permissions are required",
			permissionProvider: &mockPermissionProvider{
				GetPermissionsFunc: func(path string) ([]config.Permission, error) {
					return []config.Permission{}, nil
				},
			},
			r: func() *http.Request {
				r := httptest.NewRequest("GET", "/", nil)
				r = r.WithContext(context.WithValue(r.Context(), ContextUserRolesKey, []string{"admin"}))
				return r
			}(),
			err: nil,
		},
		{
			name: "permissions are required",
			permissionProvider: &mockPermissionProvider{
				GetPermissionsFunc: func(path string) ([]config.Permission, error) {
					return []config.Permission{{Roles: []string{"admin"}}}, nil
				},
			},
			r: func() *http.Request {
				r := httptest.NewRequest("GET", "/", nil)
				r = r.WithContext(context.WithValue(r.Context(), ContextUserRolesKey, []string{"user"}))
				return r
			}(),
			err: ErrUnauthorized,
		},
		{
			name: "permissions are required and user has one",
			permissionProvider: &mockPermissionProvider{
				GetPermissionsFunc: func(path string) ([]config.Permission, error) {
					return []config.Permission{{Roles: []string{"admin"}}}, nil
				},
			},
			r: func() *http.Request {
				r := httptest.NewRequest("GET", "/", nil)
				r = r.WithContext(context.WithValue(r.Context(), ContextUserRolesKey, []string{"admin"}))
				return r
			}(),
			err: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := &defaultAuthorizer{
				permissionProvider: tt.permissionProvider,
			}
			err := a.CheckAccess(tt.r)
			assert.Equal(t, tt.err, err)
		})
	}
}
