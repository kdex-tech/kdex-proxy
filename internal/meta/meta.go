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
	"net/http"

	"golang.org/x/net/html"
	"kdex.dev/proxy/internal/config"
	"kdex.dev/proxy/internal/dom"
)

type MetaTransformer struct {
	Config *config.Config
}

func NewMetaTransformer(config *config.Config) *MetaTransformer {
	return &MetaTransformer{
		Config: config,
	}
}

func (m *MetaTransformer) Transform(r *http.Response, doc *html.Node) error {
	if headNode := dom.FindElementByName("head", doc, nil); headNode != nil {
		metaNode := &html.Node{
			Type: html.ElementNode,
			Data: "meta",
			Attr: []html.Attribute{
				{Key: "name", Val: "kdex-ui"},
				{Key: "data-check-single-endpoint", Val: m.Config.Authz.Endpoints.Single},
				{Key: "data-check-batch-endpoint", Val: m.Config.Authz.Endpoints.Batch},
				{Key: "data-login-path", Val: m.Config.Authn.Login.Path},
				{Key: "data-login-label", Val: m.Config.Authn.Login.Label},
				{Key: "data-login-css-query", Val: m.Config.Authn.Login.Query},
				{Key: "data-logout-path", Val: m.Config.Authn.Logout.Path},
				{Key: "data-logout-label", Val: m.Config.Authn.Logout.Label},
				{Key: "data-logout-css-query", Val: m.Config.Authn.Logout.Query},
				{Key: "data-path-separator", Val: m.Config.Proxy.PathSeparator},
				{Key: "data-state-endpoint", Val: m.Config.State.Endpoint},
			},
		}
		headNode.AppendChild(metaNode)
	}

	return nil
}
