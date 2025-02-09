package importmap

import (
	"log"
	"net/http"
	"slices"

	"golang.org/x/net/html"
	"kdex.dev/proxy/internal/config"
	"kdex.dev/proxy/internal/scanner"
	"kdex.dev/proxy/internal/transform"
)

type ImportMapTransformer struct {
	transform.Transformer
	Config        *config.Config
	ModuleImports map[string]string
}

func NewImportMapTransformer(c *config.Config) *ImportMapTransformer {
	transformer := &ImportMapTransformer{
		Config: c,
	}

	if !slices.Contains(transformer.Config.Importmap.PreloadModules, "@kdex/ui") {
		transformer.Config.Importmap.PreloadModules = append(transformer.Config.Importmap.PreloadModules, "@kdex/ui")
	}

	if err := transformer.ScanForImports(); err != nil {
		log.Fatal(err)
	}

	return transformer
}

func (t *ImportMapTransformer) ScanForImports() error {
	s := scanner.NewScanner(t.Config.ModuleDir)

	if err := s.ScanRootDir(); err != nil {
		return err
	}

	s.ValidateImports()

	t.ModuleImports = s.GetImports()

	return nil
}

func (t *ImportMapTransformer) ShouldTransform(r *http.Response) bool {
	return transform.HtmlTransformCheck(r)
}

func (t *ImportMapTransformer) Transform(r *http.Response, doc *html.Node) error {
	importMapInstance, err := Parse(doc)
	if err != nil {
		return err
	}

	importMapInstance.WithMutator(t.Mutator())
	importMapInstance.WithPreloadModules(t.Config.Importmap.PreloadModules)
	importMapInstance.Mutate()

	return nil
}

func (t *ImportMapTransformer) Mutator() ImportMapMutator {
	return func(im *ImportMap) {
		for key, value := range t.ModuleImports {
			im.Imports[key] = t.Config.Fileserver.Prefix + value
		}
	}
}
