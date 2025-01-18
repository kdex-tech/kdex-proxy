package scanner

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

type PackageDependencies map[string]string

type PackageJSON struct {
	Dependencies PackageDependencies `json:"dependencies"`
	// exports can be a string, a map[string]interface{}, or a map[string]string
	Exports interface{} `json:"exports"`
	Main    string      `json:"main"`
	Module  string      `json:"module"`
	Type    string      `json:"type"`
}

type Scanner struct {
	RootDir string
	Imports map[string]string
	Scanned map[string]bool
}

func (s *Scanner) ScanDependencies(dependencies PackageDependencies) error {
	for pkgName := range dependencies {
		if err := s.ScanPackage(pkgName); err != nil {
			return err
		}
	}

	return nil
}

func NewScanner(rootDir string) *Scanner {
	return &Scanner{
		RootDir: rootDir,
		Imports: make(map[string]string),
		Scanned: make(map[string]bool),
	}
}

func (s *Scanner) ScanPackage(packageName string) error {
	if s.Scanned[packageName] {
		return nil
	}

	packagePath := filepath.Join(s.RootDir, "node_modules", packageName)

	// Read package.json
	pkgData, err := os.ReadFile(filepath.Join(packagePath, "package.json"))
	if err != nil {
		return fmt.Errorf("reading package.json: %w", err)
	}

	var pkg PackageJSON
	if err := json.Unmarshal(pkgData, &pkg); err != nil {
		return fmt.Errorf("parsing package.json: %w", err)
	}

	// Check for ES modules
	if pkg.Type == "module" || pkg.Module != "" {
		// Add main module entry
		mainEntry := pkg.Module
		if mainEntry == "" {
			mainEntry = pkg.Main
		}
		if mainEntry != "" {
			s.Imports[packageName] = filepath.Join("/node_modules", packageName, mainEntry)
		}

		// Handle exports field
		if pkg.Exports != nil {
			if targetStr, ok := pkg.Exports.(string); ok {
				if strings.HasSuffix(targetStr, ".mjs") || strings.HasSuffix(targetStr, ".js") {
					s.Imports[packageName] = filepath.Join("/node_modules", packageName, targetStr)
				}
			} else if targetMap, ok := pkg.Exports.(map[string]interface{}); ok {
				for key, value := range targetMap {
					if key == "import" {
						if valueStr, ok := value.(string); ok {
							s.Imports[packageName] = filepath.Join("/node_modules", packageName, valueStr)
						}
					}
				}
			}
		}
	}

	s.Scanned[packageName] = true

	s.ScanDependencies(pkg.Dependencies)

	return nil
}

func (s *Scanner) GenerateImports() map[string]string {
	return s.Imports
}
