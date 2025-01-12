package importmap

import (
	"encoding/json"
	"log"
	"net/http"
	"os"
	"strings"
)

const (
	DefaultModuleBodyPath    = "/etc/kdex/module_body"
	DefaultModuleBody        = "import '@kdex-ui';"
	DefaultModuleImportsPath = "/etc/kdex/module_imports"
	DefaultModuleImports     = `{
		"imports": {
			"@kdex-ui": "@kdex-ui/index.js"
		}
	}`
	DefaultModulesPrefix = "/~/m/"
)

type Imports struct {
	Imports map[string]string `json:"imports"`
}

type ImportMapTransformer struct {
	ModuleImports Imports
	ModuleBody    string
	ModulePrefix  string
}

func NewImportMapTransformerFromEnv() *ImportMapTransformer {
	moduleImportsPath := os.Getenv("MODULE_IMPORTS_PATH")
	if moduleImportsPath == "" {
		moduleImportsPath = DefaultModuleImportsPath
		log.Printf("Defaulting module_imports_path to %s", moduleImportsPath)
	}

	var imports Imports
	var moduleImportsBytes []byte
	if _, err := os.Stat(moduleImportsPath); !os.IsNotExist(err) {
		moduleImportsBytes, err = os.ReadFile(moduleImportsPath)
		if err != nil {
			log.Fatal(err)
		}
	} else {
		moduleImportsBytes = []byte(DefaultModuleImports)
		log.Printf("Defaulting module_imports to %s", moduleImportsBytes)
	}

	if err := json.Unmarshal(moduleImportsBytes, &imports); err != nil {
		log.Fatal(err)
	}

	moduleBodyPath := os.Getenv("MODULE_BODY_PATH")
	if moduleBodyPath == "" {
		moduleBodyPath = DefaultModuleBodyPath
		log.Printf("Defaulting module_body_path to %s", moduleBodyPath)
	}

	var moduleBody string
	if _, err := os.Stat(moduleBodyPath); os.IsNotExist(err) {
		moduleBody = DefaultModuleBody
		log.Printf("Defaulting module_body to %s", moduleBody)
	} else {
		moduleBodyBytes, err := os.ReadFile(moduleBodyPath)

		if err != nil {
			log.Fatal(err)
		}

		moduleBody = string(moduleBodyBytes)
	}

	return &ImportMapTransformer{
		ModuleImports: imports,
		ModuleBody:    moduleBody,
		ModulePrefix:  DefaultModulesPrefix,
	}
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

func (t *ImportMapTransformer) Transform(body *[]byte) error {
	importMapInstance, err := Parse(body)
	if err != nil {
		return err
	}

	importMapInstance.WithMutator(t.Mutator())
	importMapInstance.WithModuleBody(t.ModuleBody)

	if !importMapInstance.Mutate() {
		return nil
	}

	if err := importMapInstance.Return(body); err != nil {
		return err
	}

	return nil
}

func (t *ImportMapTransformer) Mutator() ImportMapMutator {
	return func(im *ImportMap) {
		for key, value := range t.ModuleImports.Imports {
			im.Imports[key] = t.ModulePrefix + value
		}
	}
}
