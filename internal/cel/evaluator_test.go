package cel

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestEvaluator_Evaluate(t *testing.T) {
	e, err := NewEvaluator()
	if err != nil {
		t.Fatalf("failed to create evaluator: %v", err)
	}
	tests := []struct {
		name       string
		expression string
		data       map[string]interface{}
		want       any
		wantErr    bool
	}{
		{
			name:       "boolean result",
			expression: "this.name == 'John'",
			data:       map[string]interface{}{"name": "John"},
			want:       true,
			wantErr:    false,
		},
		{
			name:       "string result",
			expression: "this.name",
			data:       map[string]interface{}{"name": "John"},
			want:       "John",
			wantErr:    false,
		},
		{
			name:       "integer result",
			expression: "this.age",
			data:       map[string]interface{}{"age": 25},
			want:       int64(25),
			wantErr:    false,
		},
		{
			name:       "float result",
			expression: "this.height",
			data:       map[string]interface{}{"height": 170.5},
			want:       170.5,
			wantErr:    false,
		},
		{
			name:       "array result",
			expression: "this.roles",
			data:       map[string]interface{}{"roles": []string{"admin", "user"}},
			want:       []string{"admin", "user"},
			wantErr:    false,
		},
		{
			name:       "invalid expression",
			expression: "this.roles[0",
			data:       map[string]interface{}{"roles": []string{"admin", "user"}},
			want:       nil,
			wantErr:    true,
		},
		{
			name:       "invalid data",
			expression: "this.name == 'John'",
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
