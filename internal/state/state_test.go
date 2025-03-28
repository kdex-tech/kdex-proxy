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

package state

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"kdex.dev/proxy/internal/config"
	kctx "kdex.dev/proxy/internal/context"
	"kdex.dev/proxy/internal/expression"
	"kdex.dev/proxy/internal/store/session"
	"kdex.dev/proxy/internal/util"
)

func TestStateHandler_StateHandler(t *testing.T) {
	defaultConfig := config.DefaultConfig()
	defaultConfig.Expressions.Principal = "data.sub"
	tests := []struct {
		name           string
		FieldEvaluator *expression.FieldEvaluator
		session        *session.SessionData
		want           string
	}{
		{
			name:           "not logged in",
			FieldEvaluator: expression.NewFieldEvaluator(defaultConfig),
			session:        nil,
			want:           `{"principal":"","roles":[],"data":{}}`,
		},
		{
			name:           "logged in",
			FieldEvaluator: expression.NewFieldEvaluator(defaultConfig),
			session: &session.SessionData{
				Data: map[string]interface{}{
					"roles": []string{"admin"},
					"sub":   "test",
				},
			},
			want: `{"principal":"test","roles":["admin"],"data":{"roles":["admin"],"sub":"test"}}`,
		},
		{
			name:           "logged in with multiple roles",
			FieldEvaluator: expression.NewFieldEvaluator(defaultConfig),
			session: &session.SessionData{
				Data: map[string]interface{}{
					"roles": []string{"admin", "user"},
					"sub":   "test",
				},
			},
			want: `{"principal":"test","roles":["admin","user"],"data":{"roles":["admin","user"],"sub":"test"}}`,
		},
		{
			name:           "no identity",
			FieldEvaluator: expression.NewFieldEvaluator(defaultConfig),
			session: &session.SessionData{
				Data: map[string]interface{}{
					"roles": []string{"admin", "user"},
				},
			},
			want: `{"principal":"","roles":["admin","user"],"data":{"roles":["admin","user"]}}`,
		},
		{
			name:           "no roles",
			FieldEvaluator: expression.NewFieldEvaluator(defaultConfig),
			session: &session.SessionData{
				Data: map[string]interface{}{
					"sub": "test",
				},
			},
			want: `{"principal":"test","roles":[],"data":{"sub":"test"}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &StateHandler{
				FieldEvaluator: tt.FieldEvaluator,
			}
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest("GET", "/", nil)
			request = request.WithContext(context.WithValue(request.Context(), kctx.SessionDataKey, tt.session))
			handler := h.StateHandler()
			handler(recorder, request)
			assert.Equal(t, tt.want, util.NormalizeString(recorder.Body.String()))
		})
	}
}
