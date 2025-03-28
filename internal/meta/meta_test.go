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

package meta

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"golang.org/x/net/html"
	"kdex.dev/proxy/internal/config"
	"kdex.dev/proxy/internal/store/session"
	"kdex.dev/proxy/internal/util"
)

func TestMetaTransformer_Transform(t *testing.T) {
	defaultConfig := config.DefaultConfig()
	defaultConfig.Proxy.PathSeparator = "/_/"
	defaultConfig.Authn.Login.Path = "/~/oauth/login"
	defaultConfig.Authn.Login.Label = "Login"
	defaultConfig.Authn.Login.Query = `nav a[href="/signin/"]`
	defaultConfig.Authn.Logout.Path = "/~/oauth/logout"
	defaultConfig.Authn.Logout.Label = "Logout"
	defaultConfig.Authn.Logout.Query = `nav a[href="/signin/"]`
	defaultConfig.State.Endpoint = "/~/state"

	type fields struct {
		Config        *config.Config
		SessionHelper *session.SessionHelper
	}
	tests := []struct {
		name    string
		fields  fields
		doc     *html.Node
		wantErr bool
		want    string
	}{
		{
			name: "no head node",
			fields: fields{
				Config: &config.Config{},
			},
			doc:     &html.Node{},
			wantErr: false,
			want:    "",
		},
		{
			name: "head node",
			fields: fields{
				Config: defaultConfig,
			},
			doc: &html.Node{
				Type: html.ElementNode,
				Data: "head",
			},
			wantErr: false,
			want:    `<head><meta name="kdex-ui" data-check-single-endpoint="/~/check/single" data-check-batch-endpoint="/~/check/batch" data-login-path="/~/oauth/login" data-login-label="Login" data-login-css-query="nav a[href=&#34;/signin/&#34;]" data-logout-path="/~/oauth/logout" data-logout-label="Logout" data-logout-css-query="nav a[href=&#34;/signin/&#34;]" data-path-separator="/_/" data-state-endpoint="/~/state"/></head>`,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			m := &MetaTransformer{
				Config: tt.fields.Config,
			}
			err := m.Transform(nil, tt.doc)
			assert.Equal(t, tt.wantErr, err != nil)
			assert.Equal(t, tt.want, util.FromDoc(tt.doc))
		})
	}
}
