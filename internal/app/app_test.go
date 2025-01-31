package app

import (
	"reflect"
	"testing"
)

func TestNewAppManagerFromEnv(t *testing.T) {
	tests := []struct {
		name string
		want *AppManager
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAppManagerFromEnv(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAppManagerFromEnv() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppManager_GetApps(t *testing.T) {
	type fields struct {
		apps Apps
	}
	tests := []struct {
		name   string
		fields fields
		want   Apps
	}{
		{
			name: "test",
			fields: fields{
				apps: Apps{
					{Address: "sample-app"},
				},
			},
			want: Apps{
				{Address: "sample-app"},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &AppManager{
				apps: tt.fields.apps,
			}
			if got := m.GetApps(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AppManager.GetApps() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestAppManager_GetAppsForPage(t *testing.T) {
	type fields struct {
		apps Apps
	}
	type args struct {
		page string
	}
	tests := []struct {
		name   string
		fields fields
		args   args
		want   Apps
	}{
		{
			name: "no apps found",
			fields: fields{
				apps: Apps{},
			},
			args: args{page: "/posts"},
			want: nil,
		},
		{
			name: "one app found among one app",
			fields: fields{
				apps: Apps{
					{Address: "sample-app", Targets: []Target{{Page: "/posts", Container: "main"}}},
				},
			},
			args: args{page: "/posts"},
			want: Apps{
				{Address: "sample-app", Targets: []Target{{Page: "/posts", Container: "main"}}},
			},
		},
		{
			name: "one app found among two apps",
			fields: fields{
				apps: Apps{
					{Address: "sample-app", Targets: []Target{{Page: "/posts", Container: "main"}}},
					{Address: "sample-app-2", Targets: []Target{{Page: "/other", Container: "main"}}},
				},
			},
			args: args{page: "/posts"},
			want: Apps{
				{Address: "sample-app", Targets: []Target{{Page: "/posts", Container: "main"}}},
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &AppManager{
				apps: tt.fields.apps,
			}
			if got := m.GetAppsForPage(tt.args.page); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("AppManager.GetAppsForPage() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestValidateApp(t *testing.T) {
	type args struct {
		app App
	}
	tests := []struct {
		name    string
		args    args
		wantErr bool
	}{
		{
			name: "missing address",
			args: args{
				app: App{},
			},
			wantErr: true,
		},
		{
			name: "missing element",
			args: args{
				app: App{Address: "sample-app"},
			},
			wantErr: true,
		},
		{
			name: "missing path",
			args: args{
				app: App{Address: "sample-app", Element: "sample-element"},
			},
			wantErr: true,
		},
		{
			name: "missing targets",
			args: args{
				app: App{Address: "sample-app", Element: "sample-element", Path: "sample-path"},
			},
			wantErr: true,
		},
		{
			name: "missing targets",
			args: args{
				app: App{Address: "sample-app", Element: "sample-element", Path: "sample-path", Targets: []Target{}},
			},
			wantErr: true,
		},
		{
			name: "missing targets page",
			args: args{
				app: App{Address: "sample-app", Element: "sample-element", Path: "sample-path", Targets: []Target{{Container: "sample-container"}}},
			},
			wantErr: true,
		},
		{
			name: "valid app",
			args: args{
				app: App{Address: "sample-app", Element: "sample-element", Path: "sample-path", Targets: []Target{{Page: "sample-page", Container: "sample-container"}}},
			},
			wantErr: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if err := ValidateApp(tt.args.app); (err != nil) != tt.wantErr {
				t.Errorf("ValidateApp() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
