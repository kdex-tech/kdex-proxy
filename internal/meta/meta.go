package meta

import (
	"net/http"

	"golang.org/x/net/html"
	"kdex.dev/proxy/internal/config"
	"kdex.dev/proxy/internal/dom"
	"kdex.dev/proxy/internal/transform"
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
				{Key: "data-path-separator", Val: m.Config.Proxy.PathSeparator},
				{Key: "data-login-path", Val: m.Config.Authn.Login.Path},
				{Key: "data-login-label", Val: m.Config.Authn.Login.Label},
				{Key: "data-login-css-query", Val: m.Config.Authn.Login.Query},
				{Key: "data-logout-path", Val: m.Config.Authn.Logout.Path},
				{Key: "data-logout-label", Val: m.Config.Authn.Logout.Label},
				{Key: "data-logout-css-query", Val: m.Config.Authn.Logout.Query},
				{Key: "data-state-endpoint", Val: m.Config.State.Endpoint},
			},
		}
		headNode.AppendChild(metaNode)
	}

	return nil
}

func (m *MetaTransformer) ShouldTransform(r *http.Response) bool {
	return transform.HtmlTransformCheck(r)
}
