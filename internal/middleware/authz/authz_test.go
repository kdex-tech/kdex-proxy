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
