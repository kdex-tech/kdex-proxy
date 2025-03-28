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

package expression

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluator_Evaluate(t *testing.T) {
	e := NewEvaluator()
	tests := []struct {
		name       string
		expression string
		data       map[string]interface{}
		want       any
		wantErr    bool
	}{
		{
			name:       "boolean result",
			expression: "data.name == 'John'",
			data:       map[string]interface{}{"name": "John"},
			want:       true,
			wantErr:    false,
		},
		{
			name:       "string result",
			expression: "data.name",
			data:       map[string]interface{}{"name": "John"},
			want:       "John",
			wantErr:    false,
		},
		{
			name:       "integer result",
			expression: "data.age",
			data:       map[string]interface{}{"age": 25},
			want:       int64(25),
			wantErr:    false,
		},
		{
			name:       "float result",
			expression: "data.height",
			data:       map[string]interface{}{"height": 170.5},
			want:       170.5,
			wantErr:    false,
		},
		{
			name:       "array result",
			expression: "data.roles",
			data:       map[string]interface{}{"roles": []string{"admin", "user"}},
			want:       []string{"admin", "user"},
			wantErr:    false,
		},
		{
			name:       "invalid expression",
			expression: "data.roles[0",
			data:       map[string]interface{}{"roles": []string{"admin", "user"}},
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "invalid data",
			expression: "data.name == 'John'",
			data:       map[string]interface{}{"roles": []string{"admin", "user"}},
			want:       nil,
			wantErr:    true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := e.Evaluate(tt.expression, tt.data)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, got)
		})
	}
}
