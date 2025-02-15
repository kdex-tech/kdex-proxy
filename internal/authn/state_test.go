package authn

import (
	"context"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"kdex.dev/proxy/internal/config"
	"kdex.dev/proxy/internal/expression"
	"kdex.dev/proxy/internal/store/session"
	"kdex.dev/proxy/internal/util"
)

func TestStateHandler_StateHandler(t *testing.T) {
	defaultConfig := config.DefaultConfig()
	evaluator, err := expression.NewEvaluator()
	if err != nil {
		t.Fatalf("failed to create evaluator: %v", err)
	}
	type fields struct {
		FieldEvaluator *expression.FieldEvaluator
	}
	tests := []struct {
		name    string
		fields  fields
		session *session.SessionData
		want    string
	}{
		{
			name: "not logged in",
			fields: fields{
				FieldEvaluator: expression.NewFieldEvaluator(evaluator, &defaultConfig),
			},
			session: nil,
			want:    `{"identity":"","isLoggedIn":false,"roles":[],"data":{}}`,
		},
		{
			name: "logged in",
			fields: fields{
				FieldEvaluator: expression.NewFieldEvaluator(evaluator, &defaultConfig),
			},
			session: &session.SessionData{
				Data: map[string]interface{}{
					"roles": []string{"admin"},
					"sub":   "test",
				},
			},
			want: `{"identity":"test","isLoggedIn":true,"roles":["admin"],"data":{"roles":["admin"],"sub":"test"}}`,
		},
		{
			name: "logged in with multiple roles",
			fields: fields{
				FieldEvaluator: expression.NewFieldEvaluator(evaluator, &defaultConfig),
			},
			session: &session.SessionData{
				Data: map[string]interface{}{
					"roles": []string{"admin", "user"},
					"sub":   "test",
				},
			},
			want: `{"identity":"test","isLoggedIn":true,"roles":["admin","user"],"data":{"roles":["admin","user"],"sub":"test"}}`,
		},
		{
			name: "no identity",
			fields: fields{
				FieldEvaluator: expression.NewFieldEvaluator(evaluator, &defaultConfig),
			},
			session: &session.SessionData{
				Data: map[string]interface{}{
					"roles": []string{"admin", "user"},
				},
			},
			want: `{"identity":"","isLoggedIn":true,"roles":["admin","user"],"data":{"roles":["admin","user"]}}`,
		},
		{
			name: "no roles",
			fields: fields{
				FieldEvaluator: expression.NewFieldEvaluator(evaluator, &defaultConfig),
			},
			session: &session.SessionData{
				Data: map[string]interface{}{
					"sub": "test",
				},
			},
			want: `{"identity":"test","isLoggedIn":true,"roles":[],"data":{"sub":"test"}}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &StateHandler{
				FieldEvaluator: tt.fields.FieldEvaluator,
			}
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest("GET", "/", nil)
			request = request.WithContext(context.WithValue(request.Context(), ContextUserKey, tt.session))
			handler := h.StateHandler()
			handler(recorder, request)
			assert.Equal(t, tt.want, util.NormalizeString(recorder.Body.String()))
		})
	}
}
