package app

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
	"kdex.dev/proxy/internal/config"
	"kdex.dev/proxy/internal/dom"
	"kdex.dev/proxy/internal/store/session"
	"kdex.dev/proxy/internal/transform"
	"kdex.dev/proxy/internal/util"
)

const (
	KDEX_UI_APP_CONTAINER_ID = "kdex-ui-app-container"
)

type AppTransformer struct {
	transform.Transformer
	Apps              *config.Apps
	Login             *config.LoginConfig
	Logout            *config.LogoutConfig
	PathSeparator     string
	SessionCookieName string
	SessionStore      *session.SessionStore
}

func (t *AppTransformer) Transform(r *http.Response, doc *html.Node) error {
	t.applyMetadata(r, doc)

	targetPath := strings.TrimSuffix(r.Request.URL.Path, "/")
	log.Printf("Looking for apps for %s", targetPath)
	apps := t.Apps.GetAppsForTargetPath(targetPath)

	if len(apps) == 0 {
		return nil
	}

	appAlias := r.Request.Header.Get("X-Kdex-Proxy-App-Alias")
	appPath := r.Request.Header.Get("X-Kdex-Proxy-App-Path")

	bodyNode := dom.FindElementByName("body", doc, nil)

	for _, app := range apps {
		for _, target := range app.Targets {
			appContainerNode := dom.FindElementByName(KDEX_UI_APP_CONTAINER_ID, doc, func(n *html.Node) bool {
				foundId := false
				for _, a := range n.Attr {
					if a.Key == "id" {
						foundId = true
					}
					if a.Key == "id" && a.Val == target.Container {
						return true
					}
				}

				// If the containerId is not found, but the containerId is empty, we can assume that this is the default container
				if !foundId && (target.Container == "" || target.Container == "main") {
					return true
				}

				return false
			})

			if appContainerNode == nil {
				log.Printf("App container for element %s targeting %s/%s not found", app.Element, target.Path, target.Container)
				continue
			}

			log.Printf("App container for element %s targeting %s/%s found", app.Element, target.Path, target.Container)

			// append a new custom element node to element
			customElement := &html.Node{
				Type: html.ElementNode,
				Data: app.Element,
				Attr: []html.Attribute{
					{Key: "id", Val: app.Alias},
				},
			}

			if app.Alias == appAlias && appPath != "" {
				customElement.Attr = append(customElement.Attr, html.Attribute{Key: "route-path", Val: appPath})
			}

			// remove all children of the app container node
			for c := appContainerNode.FirstChild; c != nil; c = c.NextSibling {
				appContainerNode.RemoveChild(c)
			}

			appContainerNode.AppendChild(customElement)
		}

		if bodyNode != nil {
			scriptNode := &html.Node{
				Type: html.ElementNode,
				Data: "script",
				Attr: []html.Attribute{
					{Key: "type", Val: "module"},
					{Key: "src", Val: fmt.Sprintf(
						"%s://%s/%s",
						util.GetScheme(r.Request),
						app.Address,
						strings.TrimPrefix(app.Path, "/"),
					)},
				},
			}

			bodyNode.AppendChild(scriptNode)
		}
	}

	return nil
}

func (t *AppTransformer) ShouldTransform(r *http.Response) bool {
	// Check if response is HTML and not streaming
	contentType := r.Header.Get("Content-Type")
	isHTML := strings.Contains(contentType, "text/html")
	isStreaming := r.Header.Get("Transfer-Encoding") == "chunked"

	if !isHTML || isStreaming {
		return false
	}

	return true
}

func (t *AppTransformer) applyMetadata(r *http.Response, doc *html.Node) {
	headNode := dom.FindElementByName("head", doc, nil)

	isLoggedIn, err := t.getSessionStatus(r)
	if err != nil {
		log.Printf("Error getting session status: %v", err)
	}

	if headNode != nil {
		metaNode := &html.Node{
			Type: html.ElementNode,
			Data: "meta",
			Attr: []html.Attribute{
				{Key: "name", Val: "kdex-ui"},
				{Key: "data-path-separator", Val: t.PathSeparator},
				{Key: "data-login-path", Val: t.Login.Path},
				{Key: "data-login-label", Val: t.Login.Label},
				{Key: "data-login-css-query", Val: t.Login.CSSQuery},
				{Key: "data-logout-path", Val: t.Logout.Path},
				{Key: "data-logout-label", Val: t.Logout.Label},
				{Key: "data-logout-css-query", Val: t.Logout.CSSQuery},
				{Key: "data-logged-in", Val: fmt.Sprintf("%t", isLoggedIn)},
			},
		}
		headNode.AppendChild(metaNode)
	}
}

func (t *AppTransformer) getSessionStatus(r *http.Response) (bool, error) {
	sessionCookie, err := r.Request.Cookie(t.SessionCookieName)
	if err != nil {
		return false, err
	}

	isLoggedIn, err := (*t.SessionStore).IsLoggedIn(r.Request.Context(), sessionCookie.Value)
	if err != nil {
		return false, err
	}

	return isLoggedIn, nil
}
