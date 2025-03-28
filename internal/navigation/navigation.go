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

package navigation

import (
	"bytes"
	"fmt"
	"log"
	"net/http"
	"sort"
	"strings"
	"text/template"

	"github.com/antchfx/htmlquery"
	"golang.org/x/net/html"
	"kdex.dev/proxy/internal/config"
	"kdex.dev/proxy/internal/transform"
	"kdex.dev/proxy/internal/util"
)

type NavigationTransformer struct {
	transform.Transformer
	Config  *config.Config
	navTmpl *template.Template
}

func NewNavigationTransformer(config *config.Config) *NavigationTransformer {
	tmpl := template.Must(template.New("Navigation").Parse(config.Navigation.NavItemTemplate))
	return &NavigationTransformer{
		Config:  config,
		navTmpl: tmpl,
	}
}

func (t *NavigationTransformer) Transform(r *http.Response, doc *html.Node) error {
	if t.Config.Navigation.NavItemsQuery == "" {
		return nil
	}

	nodes, err := htmlquery.QueryAll(doc, t.Config.Navigation.NavItemsQuery)
	if err != nil {
		return fmt.Errorf(`error querying navigation items: %w`, err)
	}

	if len(nodes) == 0 {
		log.Printf("No navigation items found matching query: %s", t.Config.Navigation.NavItemsQuery)
		return nil
	}

	navItems := make([]map[string]interface{}, 0, len(nodes))
	navItemFields := t.Config.Navigation.NavItemFields
	fieldKeys := util.Keys(navItemFields)

	for index, node := range nodes {
		item := make(map[string]interface{})

		for _, fieldKey := range fieldKeys {
			field := navItemFields[fieldKey]
			fieldValue, err := htmlquery.Query(node, field)
			if err != nil {
				return fmt.Errorf(`error querying navigation item field %s: %w`, fieldKey, err)
			}

			if fieldValue != nil && fieldValue.Type == html.ElementNode {
				item[fieldKey] = fieldValue.FirstChild.Data
			} else {
				item[fieldKey] = fieldValue.Data
			}
		}

		item["weight"] = float64(index)

		navItems = append(navItems, item)
	}

	// insert the template_paths from the config into the navItems
	for _, templatePath := range t.Config.Navigation.TemplatePaths {
		navItems = append(navItems, map[string]interface{}{
			"href":   templatePath.Href,
			"label":  templatePath.Label,
			"weight": templatePath.Weight,
		})
	}

	sort.Slice(navItems, func(i, j int) bool {
		return navItems[i]["weight"].(float64) < navItems[j]["weight"].(float64)
	})

	var output bytes.Buffer

	for _, item := range navItems {
		if t.Config.Proxy.AlwaysAppendSlash {
			item["href"] = strings.TrimRight(item["href"].(string), "/") + "/"
		}

		err = t.navTmpl.Execute(&output, item)
		if err != nil {
			return fmt.Errorf(`error executing navigation item template: %w`, err)
		}
	}

	navNode := nodes[0].Parent

	for _, node := range nodes {
		navNode.RemoveChild(node)
	}

	reader := bytes.NewReader(output.Bytes())
	newNodes, err := html.ParseFragment(reader, navNode)
	if err != nil {
		return fmt.Errorf(`error parsing navigation template: %w`, err)
	}

	for _, node := range newNodes {
		navNode.AppendChild(node)
	}

	return nil
}
