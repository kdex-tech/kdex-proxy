package app

import (
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
	"kdex.dev/proxy/internal/config"
	"kdex.dev/proxy/internal/dom"
	"kdex.dev/proxy/internal/transform"
	"kdex.dev/proxy/internal/util"
)

const (
	KDEX_UI_APP_CONTAINER_ID = "kdex-ui-app-container"
)

type AppTransformer struct {
	Config *config.Config
}

func NewAppTransformer(config *config.Config) *AppTransformer {
	return &AppTransformer{
		Config: config,
	}
}

func (t *AppTransformer) Transform(r *http.Response, doc *html.Node) error {
	targetPath := strings.TrimSuffix(r.Request.URL.Path, "/")
	log.Printf("Looking for apps for %s", targetPath)
	apps := t.Config.GetAppsForTargetPath(targetPath)

	if len(apps) == 0 {
		return nil
	}

	appAlias := r.Request.Header.Get("X-Kdex-Proxy-App-Alias")
	appPath := r.Request.Header.Get("X-Kdex-Proxy-App-Path")

	bodyNode := dom.FindElementByName("body", doc, nil)

	for _, app := range apps {
		var appContainerNode *html.Node
		for _, target := range app.Targets {
			appContainerNode = dom.FindElementByName(KDEX_UI_APP_CONTAINER_ID, doc, func(n *html.Node) bool {
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
			break
		}

		if bodyNode != nil && appContainerNode != nil {
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
	return transform.HtmlTransformCheck(r)
}
