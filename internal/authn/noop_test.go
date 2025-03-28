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

package authn

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestNoOpAuthValidator_Register(t *testing.T) {
	tests := []struct {
		name string
		v    *NoOpAuthValidator
	}{
		{
			name: "register",
			v:    &NoOpAuthValidator{},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.v
			v.Register(nil)
		})
	}
}

func TestNoOpAuthValidator_Validate(t *testing.T) {
	type args struct {
		w http.ResponseWriter
		r *http.Request
	}
	tests := []struct {
		name string
		v    *NoOpAuthValidator
		args args
		want func(h http.Handler)
	}{
		{
			name: "validate",
			v:    &NoOpAuthValidator{},
			args: args{
				w: httptest.NewRecorder(),
				r: httptest.NewRequest("GET", "/", nil),
			},
			want: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			v := tt.v
			got := v.Validate(tt.args.w, tt.args.r)
			assert.True(t, tt.want == nil, got == nil)
		})
	}
}
