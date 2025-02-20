package check

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"kdex.dev/proxy/internal/authz"
	"kdex.dev/proxy/internal/config"
	"kdex.dev/proxy/internal/util"
)

func TestCheckHandler_CheckHandler(t *testing.T) {
	type fields struct {
		permissions []config.Permission
		roles       []string
		resource    string
		action      string
	}
	tests := []struct {
		name   string
		fields fields
		status int
		body   string
	}{
		{
			name: "check handler",
			fields: fields{
				permissions: []config.Permission{
					{
						Resource:  "page:/",
						Action:    "read",
						Principal: "admin",
					},
				},
				roles:    []string{"admin"},
				resource: "page:/",
				action:   "read",
			},
			status: http.StatusOK,
			body:   `{"allowed": true}`,
		},
		{
			name: "no roles",
			fields: fields{
				permissions: []config.Permission{},
				roles:       []string{},
				resource:    "page:/",
				action:      "read",
			},
			status: http.StatusInternalServerError,
			body:   `{"error": "no roles found in request context"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &CheckHandler{
				Checker: &authz.Checker{
					PermissionProvider: &authz.StaticPermissionProvider{
						Permissions: tt.fields.permissions,
					},
				},
			}
			handler := h.CheckHandler()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(
				"GET",
				"/check?resource="+tt.fields.resource+"&action="+tt.fields.action,
				nil,
			)
			request = request.WithContext(context.WithValue(request.Context(), authz.ContextUserRolesKey, tt.fields.roles))
			handler.ServeHTTP(recorder, request)
			assert.Equal(t, recorder.Code, tt.status)
			assert.Equal(t, recorder.Body.String(), tt.body)
		})
	}
}

func TestCheckHandler_CheckBatchHandler(t *testing.T) {
	type fields struct {
		permissions []config.Permission
		roles       []string
		jsonBody    string
	}
	tests := []struct {
		name   string
		fields fields
		status int
		body   string
	}{
		{
			name: "no roles",
			fields: fields{
				permissions: []config.Permission{},
				roles:       []string{},
				jsonBody:    `{"tuples": [{"resource": "page:/", "action": "read"}]}`,
			},
			status: http.StatusInternalServerError,
			body:   `{"error": "no roles found in request context"}`,
		},
		{
			name: "check batch handler",
			fields: fields{
				permissions: []config.Permission{
					{Resource: "page:/", Action: "read", Principal: "admin"},
				},
				roles:    []string{"admin"},
				jsonBody: `{"tuples": [{"resource": "page:/", "action": "read"}]}`,
			},
			status: http.StatusOK,
			body:   `[{"resource":"page:/","allowed":true,"error":null}]`,
		},
		{
			name: "invalid json",
			fields: fields{
				permissions: []config.Permission{},
				roles:       []string{"admin"},
				jsonBody:    `{"tuples": []`,
			},
			status: http.StatusInternalServerError,
			body:   `{"error": "unexpected end of JSON input"}`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			h := &CheckHandler{
				Checker: &authz.Checker{
					PermissionProvider: &authz.StaticPermissionProvider{
						Permissions: tt.fields.permissions,
					},
				},
			}
			handler := h.CheckBatchHandler()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(
				"POST",
				"/check/batch",
				bytes.NewReader([]byte(tt.fields.jsonBody)),
			)
			request = request.WithContext(context.WithValue(request.Context(), authz.ContextUserRolesKey, tt.fields.roles))
			handler.ServeHTTP(recorder, request)
			assert.Equal(t, tt.status, recorder.Code)
			assert.Equal(t, tt.body, util.NormalizeString(recorder.Body.String()))
		})
	}
}
