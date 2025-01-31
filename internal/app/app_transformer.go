package app

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
	"kdex.dev/proxy/internal/dom"
	"kdex.dev/proxy/internal/transform"
	"kdex.dev/proxy/internal/util"
)

const (
	KDEX_UI_APP_CONTAINER_ID = "kdex-ui-app-container"
)

type AppTransformer struct {
	transform.Transformer
	AppManager *AppManager
}

func (t *AppTransformer) Transform(r *http.Response, body *[]byte) error {
	path := strings.TrimSuffix(r.Request.URL.Path, "/")

	log.Printf("Looking for apps for %s", path)

	apps := t.AppManager.GetAppsForPage(path)

	if len(apps) == 0 {
		return nil
	}

	doc, err := html.Parse(bytes.NewReader(*body))
	if err != nil {
		return err
	}

	// <meta name="path-separator" content="/_/">
	headNode := dom.FindElementByName("head", doc, nil)

	if headNode != nil {
		metaNode := &html.Node{
			Type: html.ElementNode,
			Data: "meta",
			Attr: []html.Attribute{{Key: "name", Val: "path-separator"}, {Key: "content", Val: t.AppManager.pathSeparator}},
		}
		headNode.AppendChild(metaNode)
	}

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
				if !foundId && target.Container == "" {
					return true
				}

				return false
			})

			if appContainerNode == nil {
				log.Printf("App container for element %s targeting %s/%s not found", app.Element, target.Page, target.Container)
				continue
			}

			log.Printf("App container for element %s targeting %s/%s found", app.Element, target.Page, target.Container)

			// append a new custom element node to element
			customElement := &html.Node{
				Type: html.ElementNode,
				Data: app.Element,
				Attr: []html.Attribute{
					{Key: "id", Val: app.Alias},
				},
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
					{Key: "src", Val: fmt.Sprintf("%s://%s/%s", util.GetScheme(r.Request), app.Address, app.Path)},
				},
			}

			bodyNode.AppendChild(scriptNode)
		}
	}

	var buf bytes.Buffer
	if err := html.Render(&buf, doc); err != nil {
		log.Printf("Error rendering modified HTML: %v", err)
		return err
	}
	*body = buf.Bytes()

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
