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
	ModuleDir      string
	ModuleImports  map[string]string
	ModulePrefix   string
	PreloadModules []string
}

func NewImportMapTransformer(c *config.ImportmapConfig, moduleDir string, modulePrefix string) *ImportMapTransformer {
	transformer := &ImportMapTransformer{
		ModuleDir:      moduleDir,
		ModulePrefix:   modulePrefix,
		PreloadModules: c.PreloadModules,
	}

	if !slices.Contains(transformer.PreloadModules, "@kdex/ui") {
		transformer.PreloadModules = append(transformer.PreloadModules, "@kdex/ui")
	}

	if err := transformer.ScanForImports(); err != nil {
		log.Fatal(err)
	}

	return transformer
}

func (t *ImportMapTransformer) ScanForImports() error {
	s := scanner.NewScanner(t.ModuleDir)

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
	importMapInstance.WithPreloadModules(t.PreloadModules)
	importMapInstance.Mutate()

	return nil
}

func (t *ImportMapTransformer) Mutator() ImportMapMutator {
	return func(im *ImportMap) {
		for key, value := range t.ModuleImports {
			im.Imports[key] = t.ModulePrefix + value
		}
	}
}
