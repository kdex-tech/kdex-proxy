package navigation

import (
	"bytes"
	"net/http/httptest"
	"testing"
	"text/template"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
	"kdex.dev/proxy/internal/config"
	"kdex.dev/proxy/internal/store/session"
	"kdex.dev/proxy/internal/util"
)

func TestNavigationTransformer_Transform(t *testing.T) {
	type fields struct {
		NavItemsQuery   string
		NavItemFields   map[string]string
		NavItemTemplate string
		ProtectedPaths  []string
		TemplatePaths   []config.TemplatePath
	}
	type args struct {
		doc *html.Node
	}
	tests := []struct {
		name     string
		fields   fields
		args     args
		wantErr  bool
		wantBody string
	}{
		{
			name: "transform navigation",
			fields: fields{
				NavItemsQuery: `//nav//li[not(contains(@class,'Banner-item--title')) and contains(@class,'Banner-item')]`,
				NavItemFields: map[string]string{"href": `a/@href`, "label": `a/text()`},
				NavItemTemplate: `
		<li class="Banner-item">
			<a class="Banner-link u-clickable" href="{{ .href }}">{{ .label }}</a>
		</li>
`,
				ProtectedPaths: []string{"/private"},
				TemplatePaths: []config.TemplatePath{
					{
						Href:     "/foo",
						Label:    "Foo",
						Template: "/template1",
						Weight:   0.5,
					},
					{
						Href:     "/admin",
						Label:    "Admin",
						Template: "/template1",
						Weight:   10.5,
					},
				},
			},
			args: args{
				doc: util.ToDoc(`<nav>
	<ul>
		<li class="Banner-item Banner-item--title">
			<a class="Banner-link u-clickable" href="/home">Home</a>
		</li>
		<li class="Banner-item">
			<a class="Banner-link u-clickable" href="/about">About</a>
		</li>
		<li class="Banner-item">
			<a class="Banner-link u-clickable" href="/private">Private</a>
		</li>
		<li class="Banner-item">
			<a class="Banner-link u-clickable" href="/other">Other</a>
		</li>
	</ul>
</nav>`),
			},
			wantBody: `<html><head></head><body><nav>
	<ul>
		<li class="Banner-item Banner-item--title">
			<a class="Banner-link u-clickable" href="/home">Home</a>
		</li>
		<li class="Banner-item">
			<a class="Banner-link u-clickable" href="/about">About</a>
		</li>
		<li class="Banner-item">
			<a class="Banner-link u-clickable" href="/foo">Foo</a>
		</li>
		<li class="Banner-item">
			<a class="Banner-link u-clickable" href="/other">Other</a>
		</li>
		<li class="Banner-item">
			<a class="Banner-link u-clickable" href="/admin">Admin</a>
		</li>
	</ul>
</nav></body></html>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			defaultConfig := config.DefaultConfig()
			defaultConfig.Navigation.NavItemsQuery = tt.fields.NavItemsQuery
			defaultConfig.Navigation.NavItemFields = tt.fields.NavItemFields
			defaultConfig.Navigation.ProtectedPaths = tt.fields.ProtectedPaths
			defaultConfig.Navigation.TemplatePaths = tt.fields.TemplatePaths
			defaultConfig.Session.CookieName = "session_id"

			sessionStore := session.NewMemorySessionStore(&defaultConfig.Session)
			sessionHelper := session.SessionHelper{
				Config:       &defaultConfig,
				SessionStore: &sessionStore,
			}

			tr := &NavigationTransformer{
				Config:        &defaultConfig,
				navTmpl:       template.Must(template.New("Navigation").Parse(tt.fields.NavItemTemplate)),
				SessionHelper: &sessionHelper,
			}

			rec := httptest.NewRecorder()
			res := rec.Result()
			res.Request = httptest.NewRequest("GET", "/", nil)
			if err := tr.Transform(res, tt.args.doc); (err != nil) != tt.wantErr {
				t.Errorf("NavigationTransformer.Transform() error = %v, wantErr %v", err, tt.wantErr)
			}
			var buf bytes.Buffer
			html.Render(&buf, tt.args.doc)
			assert.Equal(t, util.NormalizeString(tt.wantBody), util.NormalizeString(buf.String()))
		})
	}
}
