package importmap

import (
	"encoding/json"

	"golang.org/x/net/html"
	"kdex.dev/proxy/internal/dom"
)

type ImportMapInstance struct {
	docNode    *html.Node
	mapNode    *html.Node
	importMap  ImportMap
	mutator    ImportMapMutator
	moduleBody string
}

func Parse(doc *html.Node) (*ImportMapInstance, error) {
	importMapInstance := Instance(doc)

	return importMapInstance, nil
}

func Instance(doc *html.Node) *ImportMapInstance {
	var importMapInstance ImportMapInstance
	importMapInstance.docNode = doc
	importMapInstance.mapNode = dom.FindElementByName("script", doc, func(n *html.Node) bool {
		for _, a := range n.Attr {
			if a.Key == "type" && a.Val == "importmap" {
				return true
			}
		}

		return false
	})
	return &importMapInstance
}

func (importMapInstance *ImportMapInstance) WithMutator(mutator func(importMap *ImportMap)) *ImportMapInstance {
	importMapInstance.mutator = mutator
	return importMapInstance
}

func (importMapInstance *ImportMapInstance) WithModuleBody(moduleBody string) *ImportMapInstance {
	importMapInstance.moduleBody = moduleBody
	return importMapInstance
}

func (importMapInstance *ImportMapInstance) Mutate() {
	if importMapInstance.mapNode == nil {
		headNode := dom.FindElementByName("head", importMapInstance.docNode, nil)
		if headNode == nil {
			return
		}

		importMapInstance.mapNode = &html.Node{
			Type: html.ElementNode,
			Data: "script",
			Attr: []html.Attribute{
				{Key: "type", Val: "importmap"},
			},
		}

		headNode.InsertBefore(importMapInstance.mapNode, headNode.FirstChild)
	}

	bytes := dom.GetNodeText(importMapInstance.mapNode)

	json.Unmarshal(bytes, &importMapInstance.importMap)

	if importMapInstance.importMap.Imports == nil {
		importMapInstance.importMap.Imports = make(map[string]string)
	}
	if importMapInstance.importMap.Integrity == nil {
		importMapInstance.importMap.Integrity = make(map[string]string)
	}
	if importMapInstance.importMap.Scopes == nil {
		importMapInstance.importMap.Scopes = make(map[string]map[string]string)
	}

	if importMapInstance.mutator != nil {
		importMapInstance.mutator(&importMapInstance.importMap)
	}

	mapBytes, _ := json.Marshal(importMapInstance.importMap)

	newTextNode := &html.Node{
		Type: html.TextNode,
		Data: string(mapBytes),
	}

	// Remove existing text nodes
	for c := importMapInstance.mapNode.FirstChild; c != nil; {
		next := c.NextSibling
		importMapInstance.mapNode.RemoveChild(c)
		c = next
	}

	// Add new text node
	importMapInstance.mapNode.AppendChild(newTextNode)

	// Append a new import script node to the bottom of the body
	if importMapInstance.moduleBody != "" {
		bodyNode := dom.FindElementByName("body", importMapInstance.docNode, nil)
		if bodyNode != nil {
			scriptNode := &html.Node{
				Type: html.ElementNode,
				Data: "script",
				Attr: []html.Attribute{
					{Key: "type", Val: "module"},
				},
			}

			scriptNode.AppendChild(&html.Node{
				Type: html.TextNode,
				Data: importMapInstance.moduleBody,
			})

			bodyNode.AppendChild(scriptNode)
		}
	}
}
