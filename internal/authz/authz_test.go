// Copyright 2025 KDex Tech
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package authz

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"kdex.dev/proxy/internal/check"
	"kdex.dev/proxy/internal/config"
	kctx "kdex.dev/proxy/internal/context"
	"kdex.dev/proxy/internal/permission"
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
		checker *check.Checker
		r       *http.Request
		err     error
	}{
		{
			name: "permission required but don't have roles",
			checker: &check.Checker{
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
			checker: &check.Checker{
				PermissionProvider: &mockPermissionProvider{
					GetPermissionsFunc: func(path string) ([]config.Permission, error) {
						return nil, permission.ErrNoPermissions
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
			checker: &check.Checker{
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
			checker: &check.Checker{
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
