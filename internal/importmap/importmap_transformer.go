package importmap

import (
	"log"
	"net/http"
	"strings"

	"golang.org/x/net/html"
	"kdex.dev/proxy/internal/config"
	"kdex.dev/proxy/internal/scanner"
	"kdex.dev/proxy/internal/transform"
)

type ImportMapTransformer struct {
	transform.Transformer
	ModuleDir     string
	ModuleImports map[string]string
	ModuleBody    string
	ModulePrefix  string
}

func NewImportMapTransformer(c *config.ImportmapConfig, moduleDir string) *ImportMapTransformer {
	transformer := &ImportMapTransformer{
		ModuleDir:    moduleDir,
		ModuleBody:   c.ModuleBody,
		ModulePrefix: c.ModulePrefix,
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
	// Check if response is HTML and not streaming
	contentType := r.Header.Get("Content-Type")
	isHTML := strings.Contains(contentType, "text/html")
	isStreaming := r.Header.Get("Transfer-Encoding") == "chunked"

	if !isHTML || isStreaming {
		return false
	}

	return true
}

func (t *ImportMapTransformer) Transform(r *http.Response, doc *html.Node) error {
	importMapInstance, err := Parse(doc)
	if err != nil {
		return err
	}

	importMapInstance.WithMutator(t.Mutator())
	importMapInstance.WithModuleBody(t.ModuleBody)
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
