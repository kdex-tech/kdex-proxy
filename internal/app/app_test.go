package app

import (
	"context"
	"net/http"
	"net/http/httptest"
	"reflect"
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
	"kdex.dev/proxy/internal/config"
	kctx "kdex.dev/proxy/internal/context"
	"kdex.dev/proxy/internal/util"
)

func TestApps_GetAppsForPage(t *testing.T) {
	type fields struct {
		config *config.Config
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
				config: &config.Config{},
			},
			args: args{targetPath: "/posts"},
			want: nil,
		},
		{
			name: "one app found among one app",
			fields: fields{
				config: &config.Config{
					Apps: []config.App{
						{Address: "sample-app", Targets: []config.Target{{Path: "/posts", Container: "main"}}},
					},
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
				config: &config.Config{
					Apps: []config.App{
						{Address: "sample-app", Targets: []config.Target{{Path: "/posts", Container: "main"}}},
						{Address: "sample-app-2", Targets: []config.Target{{Path: "/other", Container: "main"}}},
					},
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
			got := tt.fields.config.GetAppsForTargetPath(tt.args.targetPath)
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

func TestAppTransformer_Transform(t *testing.T) {
	type args struct {
		proxiedParts kctx.ProxiedParts
		doc          *html.Node
	}
	tests := []struct {
		name    string
		Config  *config.Config
		args    args
		wantErr bool
		want    string
	}{
		{
			name: "no apps found",
			Config: &config.Config{
				Apps: []config.App{},
			},
			args: args{
				proxiedParts: kctx.ProxiedParts{
					AppAlias:    "",
					AppPath:     "",
					ProxiedPath: "/posts",
				},
				doc: util.ToDoc("<html><head></head><body></body></html>"),
			},
			wantErr: false,
			want:    "<html><head></head><body></body></html>",
		},
		{
			name: "one app found but no kdex-ui app container element",
			Config: &config.Config{
				Apps: []config.App{
					{
						Targets: []config.Target{{Path: "/posts", Container: "main"}},
					},
				},
			},
			args: args{
				proxiedParts: kctx.ProxiedParts{
					AppAlias:    "",
					AppPath:     "",
					ProxiedPath: "/posts",
				},
				doc: util.ToDoc("<html><head></head><body></body></html>"),
			},
			wantErr: false,
			want:    "<html><head></head><body></body></html>",
		},
		{
			name: "one app found and kdex-ui app container element",
			Config: &config.Config{
				Apps: []config.App{
					{
						Address: "test-app",
						Alias:   "ta",
						Element: "test-app",
						Path:    "/app.js",
						Targets: []config.Target{{Path: "/posts", Container: "main"}},
					},
				},
			},
			args: args{
				proxiedParts: kctx.ProxiedParts{
					AppAlias:    "",
					AppPath:     "",
					ProxiedPath: "/posts",
				},
				doc: util.ToDoc("<html><head></head><body><kdex-ui-app-container></kdex-ui-app-container></body></html>"),
			},
			wantErr: false,
			want:    `<html><head></head><body><kdex-ui-app-container><test-app id="ta"></test-app></kdex-ui-app-container><script type="module" src="http://test-app/app.js"></script></body></html>`,
		},
		{
			name: "one app found and kdex-ui app container element with alias and apppath",
			Config: &config.Config{
				Apps: []config.App{
					{
						Address: "test-app",
						Alias:   "ta",
						Element: "test-app",
						Path:    "/app.js",
						Targets: []config.Target{{Path: "/posts", Container: "main"}},
					},
				},
			},
			args: args{
				proxiedParts: kctx.ProxiedParts{
					AppAlias:    "ta",
					AppPath:     "/foo",
					ProxiedPath: "/posts",
				},
				doc: util.ToDoc("<html><head></head><body><kdex-ui-app-container></kdex-ui-app-container></body></html>"),
			},
			wantErr: false,
			want:    `<html><head></head><body><kdex-ui-app-container><test-app id="ta" route-path="/foo"></test-app></kdex-ui-app-container><script type="module" src="http://test-app/app.js"></script></body></html>`,
		},
		{
			name: "one app found and kdex-ui app container element by id",
			Config: &config.Config{
				Apps: []config.App{
					{Address: "test-app", Alias: "ta", Element: "test-app", Path: "/app.js", Targets: []config.Target{{Path: "/posts", Container: "foo"}}},
				},
			},
			args: args{
				proxiedParts: kctx.ProxiedParts{
					AppAlias:    "",
					AppPath:     "",
					ProxiedPath: "/posts",
				},
				doc: util.ToDoc(`<html><head></head><body><kdex-ui-app-container id="foo"></kdex-ui-app-container></body></html>`),
			},
			wantErr: false,
			want:    `<html><head></head><body><kdex-ui-app-container id="foo"><test-app id="ta"></test-app></kdex-ui-app-container><script type="module" src="http://test-app/app.js"></script></body></html>`,
		},
		{
			name: "one app found and kdex-ui app container element by id but wrong id",
			Config: &config.Config{
				Apps: []config.App{
					{Address: "test-app", Alias: "ta", Element: "test-app", Path: "/app.js", Targets: []config.Target{{Path: "/posts", Container: "foo"}}},
				},
			},
			args: args{
				proxiedParts: kctx.ProxiedParts{
					AppAlias:    "",
					AppPath:     "",
					ProxiedPath: "/posts",
				},
				doc: util.ToDoc(`<html><head></head><body><kdex-ui-app-container id="bar"></kdex-ui-app-container></body></html>`),
			},
			wantErr: false,
			want:    `<html><head></head><body><kdex-ui-app-container id="bar"></kdex-ui-app-container></body></html>`,
		},
		{
			name: "one app found and kdex-ui app container element with child element",
			Config: &config.Config{
				Apps: []config.App{
					{Address: "test-app", Alias: "ta", Element: "test-app", Path: "/app.js", Targets: []config.Target{{Path: "/posts"}}},
				},
			},
			args: args{
				proxiedParts: kctx.ProxiedParts{
					AppAlias:    "",
					AppPath:     "",
					ProxiedPath: "/posts",
				},
				doc: util.ToDoc(`<html><head></head><body><kdex-ui-app-container><div></div></kdex-ui-app-container></body></html>`),
			},
			wantErr: false,
			want:    `<html><head></head><body><kdex-ui-app-container><test-app id="ta"></test-app></kdex-ui-app-container><script type="module" src="http://test-app/app.js"></script></body></html>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &AppTransformer{
				Config: tt.Config,
			}
			ctx := context.WithValue(context.Background(), kctx.ProxiedPartsKey, tt.args.proxiedParts)
			r := httptest.NewRequestWithContext(ctx, "GET", "/", nil)
			r.URL.Scheme = "http"
			resp := &http.Response{
				Request: r,
			}
			err := tr.Transform(resp, tt.args.doc)
			assert.Equal(t, tt.wantErr, err != nil)
			if tt.wantErr {
				return
			}
			assert.Equal(t, util.NormalizeString(tt.want), util.NormalizeString(util.FromDoc(tt.args.doc)))
		})
	}
}

func TestAppTransformer_ShouldTransform(t *testing.T) {
	tests := []struct {
		name   string
		Config *config.Config
		r      *http.Response
		want   bool
	}{
		{
			name:   "should transform",
			Config: &config.Config{},
			r: &http.Response{
				Request: httptest.NewRequest("GET", "/", nil),
			},
			want: false,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tr := &AppTransformer{
				Config: tt.Config,
			}
			if got := tr.ShouldTransform(tt.r); got != tt.want {
				t.Errorf("AppTransformer.ShouldTransform() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestNewAppTransformer(t *testing.T) {
	type args struct {
		config *config.Config
	}
	tests := []struct {
		name string
		args args
		want *AppTransformer
	}{
		{
			name: "constructor",
			args: args{
				config: &config.Config{},
			},
			want: &AppTransformer{Config: &config.Config{}},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewAppTransformer(tt.args.config); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewAppTransformer() = %v, want %v", got, tt.want)
			}
		})
	}
}
