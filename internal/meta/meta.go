package meta

import (
	"fmt"
	"log"
	"net/http"

	"golang.org/x/net/html"
	"kdex.dev/proxy/internal/config"
	"kdex.dev/proxy/internal/dom"
	"kdex.dev/proxy/internal/store/session"
	"kdex.dev/proxy/internal/transform"
)

type MetaTransformer struct {
	transform.Transformer
	Config        *config.Config
	SessionHelper *session.SessionHelper
}

func NewMetaTransformer(config *config.Config, sessionHelper *session.SessionHelper) *MetaTransformer {
	return &MetaTransformer{
		Config:        config,
		SessionHelper: sessionHelper,
	}
}

func (m *MetaTransformer) Transform(r *http.Response, doc *html.Node) error {
	isLoggedIn, err := m.SessionHelper.IsLoggedIn(r)
	if err != nil {
		log.Printf("Error getting session status: %v", err)
	}

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
				{Key: "data-logged-in", Val: fmt.Sprintf("%t", isLoggedIn)},
			},
		}
		headNode.AppendChild(metaNode)
	}

	return nil
}

func (m *MetaTransformer) ShouldTransform(r *http.Response) bool {
	return transform.HtmlTransformCheck(r)
}
