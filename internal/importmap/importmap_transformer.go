package importmap

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"

	"golang.org/x/net/html"
	"kdex.dev/proxy/internal/scanner"
	"kdex.dev/proxy/internal/transform"
)

const (
	DefaultModuleDir              = "/modules"
	DefaultModuleBodyPath         = "/etc/kdex/module_body"
	DefaultModuleBody             = ""
	DefaultModuleDependenciesPath = "/etc/kdex/module_dependencies"
	DefaultModulesPrefix          = "/~/m/"
)

type ImportMapTransformer struct {
	transform.Transformer
	ModuleDir          string
	ModuleDependencies map[string]string
	ModuleImports      map[string]string
	ModuleBody         string
	ModulePrefix       string
}

func NewImportMapTransformerFromEnv() *ImportMapTransformer {
	moduleDir := os.Getenv("MODULE_DIR")
	if moduleDir == "" {
		moduleDir = DefaultModuleDir
		log.Printf("Defaulting module_dir to %s", moduleDir)
	}

	moduleDependenciesPath := os.Getenv("MODULE_DEPENDENCIES_PATH")
	if moduleDependenciesPath == "" {
		moduleDependenciesPath = DefaultModuleDependenciesPath
		log.Printf("Defaulting module_dependencies_path to %s", moduleDependenciesPath)
	}

	var dependencies map[string]string
	if _, err := os.Stat(moduleDependenciesPath); err == nil {
		moduleDependenciesBytes, err := os.ReadFile(moduleDependenciesPath)
		if err == nil {
			err := json.Unmarshal(moduleDependenciesBytes, &dependencies)
			if err != nil {
				log.Fatal(err)
			}
		}
	}

	moduleBodyPath := os.Getenv("MODULE_BODY_PATH")
	if moduleBodyPath == "" {
		moduleBodyPath = DefaultModuleBodyPath
		log.Printf("Defaulting module_body_path to %s", moduleBodyPath)
	}

	var moduleBody string
	if info, err := os.Stat(moduleBodyPath); err != nil || info.IsDir() {
		moduleBody = DefaultModuleBody
		log.Printf("Defaulting module_body to %s", moduleBody)
	} else {
		moduleBodyBytes, err := os.ReadFile(moduleBodyPath)

		if err != nil {
			log.Fatal(err)
		}

		moduleBody = string(moduleBodyBytes)
	}

	transformer := &ImportMapTransformer{
		ModuleDir:          moduleDir,
		ModuleDependencies: dependencies,
		ModuleBody:         moduleBody,
		ModulePrefix:       DefaultModulesPrefix,
	}

	if err := transformer.ScanForImports(); err != nil {
		log.Fatal(err)
	}

	return transformer
}

func (t *ImportMapTransformer) ScanForImports() error {
	s := scanner.NewScanner(t.ModuleDir)

	if t.ModuleDependencies != nil {
		if err := s.ScanDependencies(t.ModuleDependencies); err != nil {
			return err
		}
	} else {
		if err := s.ScanRootDir(); err != nil {
			return err
		}
	}

	s.ValidateImports()

	t.ModuleImports = s.GetImports()

	return nil
}

func (t *ImportMapTransformer) WithModulePrefix(modulePrefix string) *ImportMapTransformer {
	t.ModulePrefix = modulePrefix
	return t
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
