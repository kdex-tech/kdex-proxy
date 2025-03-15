package check

import (
	"bytes"
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/stretchr/testify/assert"
	"kdex.dev/proxy/internal/config"
	kctx "kdex.dev/proxy/internal/context"
	"kdex.dev/proxy/internal/permission"
	"kdex.dev/proxy/internal/util"
)

func TestChecker_Check(t *testing.T) {
	type fields struct {
		PermissionProvider permission.PermissionProvider
	}
	type args struct {
		ctx      context.Context
		resource string
		action   string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    bool
		wantErr error
	}{
		{
			name: "no permissions",
			fields: fields{
				PermissionProvider: &permission.StaticPermissionProvider{},
			},
			args: args{
				ctx:      context.WithValue(context.Background(), kctx.UserRolesKey, []string{"admin"}),
				resource: "page:/",
				action:   "read",
			},
			want:    false,
			wantErr: permission.ErrNoPermissions,
		},
		{
			name: "has permissions but no roles",
			fields: fields{
				PermissionProvider: &permission.StaticPermissionProvider{
					Permissions: []config.Permission{{Resource: "page:/", Action: "read"}},
				},
			},
			args: args{
				ctx:      context.Background(),
				resource: "page:/",
				action:   "read",
			},
			want:    false,
			wantErr: ErrNoRoles,
		},
		{
			name: "has permissions and roles",
			fields: fields{
				PermissionProvider: &permission.StaticPermissionProvider{
					Permissions: []config.Permission{{Resource: "page:/", Action: "read", Principal: "admin"}},
				},
			},
			args: args{
				ctx:      context.WithValue(context.Background(), kctx.UserRolesKey, []string{"admin"}),
				resource: "page:/",
				action:   "read",
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "has permissions and roles but no intersection",
			fields: fields{
				PermissionProvider: &permission.StaticPermissionProvider{
					Permissions: []config.Permission{{Resource: "page:/", Action: "read", Principal: "admin"}},
				},
			},
			args: args{
				ctx:      context.WithValue(context.Background(), kctx.UserRolesKey, []string{"user"}),
				resource: "page:/",
				action:   "read",
			},
			want:    false,
			wantErr: nil,
		},
		{
			name: "permission is a identifier glob",
			fields: fields{
				PermissionProvider: &permission.StaticPermissionProvider{
					Permissions: []config.Permission{{Resource: "page:/*", Action: "read", Principal: "admin"}},
				},
			},
			args: args{
				ctx:      context.WithValue(context.Background(), kctx.UserRolesKey, []string{"admin"}),
				resource: "page:/users/foo",
				action:   "read",
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "permission is a type glob",
			fields: fields{
				PermissionProvider: &permission.StaticPermissionProvider{
					Permissions: []config.Permission{{Resource: "*:/foo", Action: "read", Principal: "admin"}},
				},
			},
			args: args{
				ctx:      context.WithValue(context.Background(), kctx.UserRolesKey, []string{"admin"}),
				resource: "page:/foo",
				action:   "read",
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "permission is a type and identifier glob",
			fields: fields{
				PermissionProvider: &permission.StaticPermissionProvider{
					Permissions: []config.Permission{{Resource: "*:*", Action: "read", Principal: "admin"}},
				},
			},
			args: args{
				ctx:      context.WithValue(context.Background(), kctx.UserRolesKey, []string{"admin"}),
				resource: "page:/foo",
				action:   "read",
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "permission is a type, identifier and action glob",
			fields: fields{
				PermissionProvider: &permission.StaticPermissionProvider{
					Permissions: []config.Permission{{Resource: "*:*", Action: "*", Principal: "admin"}},
				},
			},
			args: args{
				ctx:      context.WithValue(context.Background(), kctx.UserRolesKey, []string{"admin"}),
				resource: "page:/foo",
				action:   "read",
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "permission is an identifier prefix glob",
			fields: fields{
				PermissionProvider: &permission.StaticPermissionProvider{
					Permissions: []config.Permission{{Resource: "page:/foo*", Action: "read", Principal: "admin"}},
				},
			},
			args: args{
				ctx:      context.WithValue(context.Background(), kctx.UserRolesKey, []string{"admin"}),
				resource: "page:/foo/bar",
				action:   "read",
			},
			want:    true,
			wantErr: nil,
		},
		{
			name: "permission is a principal prefix glob",
			fields: fields{
				PermissionProvider: &permission.StaticPermissionProvider{
					Permissions: []config.Permission{{Resource: "page:/foo", Action: "read", Principal: "adm*"}},
				},
			},
			args: args{
				ctx:      context.WithValue(context.Background(), kctx.UserRolesKey, []string{"admin"}),
				resource: "page:/foo",
				action:   "read",
			},
			want:    true,
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Checker{
				PermissionProvider: tt.fields.PermissionProvider,
			}
			got, err := c.Check(tt.args.ctx, tt.args.resource, tt.args.action)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestChecker_CheckBatch(t *testing.T) {
	type fields struct {
		PermissionProvider permission.PermissionProvider
	}
	type args struct {
		ctx    context.Context
		tuples []CheckBatchTuples
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []CheckBatchResult
		wantErr error
	}{
		{
			name: "no roles",
			fields: fields{
				PermissionProvider: &permission.StaticPermissionProvider{},
			},
			args: args{
				ctx: context.Background(),
				tuples: []CheckBatchTuples{
					{Resource: "page:/", Action: "read"},
				},
			},
			want:    nil,
			wantErr: ErrNoRoles,
		},
		{
			name: "no permissions",
			fields: fields{
				PermissionProvider: &permission.StaticPermissionProvider{
					Permissions: []config.Permission{},
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), kctx.UserRolesKey, []string{"admin"}),
				tuples: []CheckBatchTuples{
					{Resource: "page:/", Action: "read"},
					{Resource: "page:/users/", Action: "read"},
				},
			},
			want: []CheckBatchResult{
				{Resource: "page:/", Allowed: false, Error: permission.ErrNoPermissions},
				{Resource: "page:/users/", Allowed: false, Error: permission.ErrNoPermissions},
			},
			wantErr: nil,
		},
		{
			name: "has some permissions",
			fields: fields{
				PermissionProvider: &permission.StaticPermissionProvider{
					Permissions: []config.Permission{
						{Resource: "page:/", Action: "read", Principal: "admin"},
					},
				},
			},
			args: args{
				ctx: context.WithValue(context.Background(), kctx.UserRolesKey, []string{"admin"}),
				tuples: []CheckBatchTuples{
					{Resource: "page:/", Action: "read"},
					{Resource: "page:/users/", Action: "read"},
				},
			},
			want: []CheckBatchResult{
				{Resource: "page:/", Allowed: true, Error: nil},
				{Resource: "page:/users/", Allowed: false, Error: permission.ErrNoPermissions},
			},
			wantErr: nil,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Checker{
				PermissionProvider: tt.fields.PermissionProvider,
			}
			got, err := c.CheckBatch(tt.args.ctx, tt.args.tuples)
			assert.Equal(t, tt.wantErr, err)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestChecker_SingleHandler(t *testing.T) {
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
			h := &Checker{
				PermissionProvider: &permission.StaticPermissionProvider{
					Permissions: tt.fields.permissions,
				},
			}
			handler := h.SingleHandler()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(
				"GET",
				"/check?resource="+tt.fields.resource+"&action="+tt.fields.action,
				nil,
			)
			request = request.WithContext(context.WithValue(request.Context(), kctx.UserRolesKey, tt.fields.roles))
			handler.ServeHTTP(recorder, request)
			assert.Equal(t, recorder.Code, tt.status)
			assert.Equal(t, recorder.Body.String(), tt.body)
		})
	}
}

func TestChecker_BatchHandler(t *testing.T) {
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
			h := &Checker{
				PermissionProvider: &permission.StaticPermissionProvider{
					Permissions: tt.fields.permissions,
				},
			}
			handler := h.BatchHandler()
			recorder := httptest.NewRecorder()
			request := httptest.NewRequest(
				"POST",
				"/check/batch",
				bytes.NewReader([]byte(tt.fields.jsonBody)),
			)
			request = request.WithContext(context.WithValue(request.Context(), kctx.UserRolesKey, tt.fields.roles))
			handler.ServeHTTP(recorder, request)
			assert.Equal(t, tt.status, recorder.Code)
			assert.Equal(t, tt.body, util.NormalizeString(recorder.Body.String()))
		})
	}
}
