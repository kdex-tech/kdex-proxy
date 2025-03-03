package authz

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
	"kdex.dev/proxy/internal/config"
	kctx "kdex.dev/proxy/internal/context"
)

func TestChecker_Check(t *testing.T) {
	type fields struct {
		PermissionProvider PermissionProvider
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
				PermissionProvider: &StaticPermissionProvider{},
			},
			args: args{
				ctx:      context.WithValue(context.Background(), kctx.UserRolesKey, []string{"admin"}),
				resource: "page:/",
				action:   "read",
			},
			want:    false,
			wantErr: ErrNoPermissions,
		},
		{
			name: "has permissions but no roles",
			fields: fields{
				PermissionProvider: &StaticPermissionProvider{
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
				PermissionProvider: &StaticPermissionProvider{
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
				PermissionProvider: &StaticPermissionProvider{
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
				PermissionProvider: &StaticPermissionProvider{
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
				PermissionProvider: &StaticPermissionProvider{
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
				PermissionProvider: &StaticPermissionProvider{
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
				PermissionProvider: &StaticPermissionProvider{
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
				PermissionProvider: &StaticPermissionProvider{
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
				PermissionProvider: &StaticPermissionProvider{
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
		PermissionProvider PermissionProvider
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
				PermissionProvider: &StaticPermissionProvider{},
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
				PermissionProvider: &StaticPermissionProvider{
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
				{Resource: "page:/", Allowed: false, Error: ErrNoPermissions},
				{Resource: "page:/users/", Allowed: false, Error: ErrNoPermissions},
			},
			wantErr: nil,
		},
		{
			name: "has some permissions",
			fields: fields{
				PermissionProvider: &StaticPermissionProvider{
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
				{Resource: "page:/users/", Allowed: false, Error: ErrNoPermissions},
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
