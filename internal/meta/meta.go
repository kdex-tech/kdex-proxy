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
	Config       *config.Config
	SessionStore *session.SessionStore
}

func (m *MetaTransformer) Transform(r *http.Response, doc *html.Node) error {
	isLoggedIn, err := m.getSessionStatus(r)
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

func (m *MetaTransformer) getSessionStatus(r *http.Response) (bool, error) {
	sessionCookie, err := r.Request.Cookie(m.Config.Session.CookieName)
	if err != nil {
		return false, err
	}

	if m.SessionStore == nil {
		return false, nil
	}

	isLoggedIn, err := (*m.SessionStore).IsLoggedIn(r.Request.Context(), sessionCookie.Value)
	if err != nil {
		return false, err
	}

	return isLoggedIn, nil
}
