package app

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"kdex.dev/proxy/internal/config"
)

func TestApps_GetAppsForPage(t *testing.T) {
	type fields struct {
		apps *config.Apps
	}
	type args struct {
		targetPath string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   []config.App
	}{
		{
			name: "no apps found",
			fields: fields{
				apps: &config.Apps{},
			},
			args: args{targetPath: "/posts"},
			want: nil,
		},
		{
			name: "one app found among one app",
			fields: fields{
				apps: &config.Apps{
					{Address: "sample-app", Targets: []config.Target{{Path: "/posts", Container: "main"}}},
				},
			},
			args: args{targetPath: "/posts"},
			want: []config.App{
				{Address: "sample-app", Targets: []config.Target{{Path: "/posts", Container: "main"}}},
			},
		},
		{
			name: "one app found among two apps",
			fields: fields{
				apps: &config.Apps{
					{Address: "sample-app", Targets: []config.Target{{Path: "/posts", Container: "main"}}},
					{Address: "sample-app-2", Targets: []config.Target{{Path: "/other", Container: "main"}}},
				},
			},
			args: args{targetPath: "/posts"},
			want: []config.App{
				{Address: "sample-app", Targets: []config.Target{{Path: "/posts", Container: "main"}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := tt.fields.apps.GetAppsForTargetPath(tt.args.targetPath)
			assert.Equal(t, tt.want, got)
		})
	}
}

func TestValidateApp(t *testing.T) {
	type args struct {
		app *config.App
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "missing address",
			args: args{
				app: &config.App{},
			},
			wantErr: true,
		},
		{
			name: "missing element",
			args: args{
				app: &config.App{Address: "sample-app"},
			},
			wantErr: true,
		},
		{
			name: "missing path",
			args: args{
				app: &config.App{Address: "sample-app", Element: "sample-element"},
			},
			wantErr: true,
		},
		{
			name: "missing targets",
			args: args{
				app: &config.App{Address: "sample-app", Element: "sample-element", Path: "sample-path"},
			},
			wantErr: true,
		},
		{
			name: "missing targets",
			args: args{
				app: &config.App{Address: "sample-app", Element: "sample-element", Path: "sample-path", Targets: []config.Target{}},
			},
			wantErr: true,
		},
		{
			name: "missing targets page",
			args: args{
				app: &config.App{Address: "sample-app", Element: "sample-element", Path: "sample-path", Targets: []config.Target{{Container: "sample-container"}}},
			},
			wantErr: true,
		},
		{
			name: "valid app",
			args: args{
				app: &config.App{Address: "sample-app", Element: "sample-element", Path: "sample-path", Targets: []config.Target{{Path: "sample-page", Container: "sample-container"}}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := tt.args.app.Validate(); (err != nil) != tt.wantErr {
				t.Errorf("ValidateApp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
