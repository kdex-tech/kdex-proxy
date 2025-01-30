package scanner

import (
	"encoding/json"
	"fmt"
	"log"
	"os"
	"path/filepath"
	"regexp"
	"strings"
)

const (
	ES_DIR  = "es"
	ESM_DIR = "esm"
	MJS_DIR = "mjs"
)

var modulePackageDirs = [...]string{
	ESM_DIR,
	ES_DIR,
	MJS_DIR,
}

type PackageDependencies map[string]string

type PackageJSON struct {
	Browser      string              `json:"browser"`
	Dependencies PackageDependencies `json:"dependencies"`
	// exports can be a string, a map[string]interface{}, or a map[string]string
	Exports interface{} `json:"exports"`
	Main    string      `json:"main"`
	Module  string      `json:"module"`
	Type    string      `json:"type"`
}

type Scanner struct {
	ModuleDir string
	Imports   map[string]string
	Scanned   map[string]bool
}

func (s *Scanner) ScanRootDir() error {
	entries, err := os.ReadDir(s.ModuleDir)
	if err != nil {
		return err
	}

	for _, e := range entries {
		if e.IsDir() {
			s.ScanPackage(e.Name())
		}
	}

	return nil
}

func (s *Scanner) ScanDependencies(dependencies PackageDependencies) error {
	for pkgName := range dependencies {
		if err := s.ScanPackage(pkgName); err != nil {
			return err
		}
	}

	return nil
}

func NewScanner(moduleDir string) *Scanner {
	return &Scanner{
		ModuleDir: moduleDir,
		Imports:   make(map[string]string),
		Scanned:   make(map[string]bool),
	}
}

func (s *Scanner) ScanPackage(packageName string) error {
	if s.Scanned[packageName] {
		return nil
	}

	packagePath := filepath.Join(s.ModuleDir, packageName)

	packageJsonPath := filepath.Join(packagePath, "package.json")

	_, err := os.Stat(packageJsonPath)
	if err != nil {
		// look at subdirectories
		entries, err := os.ReadDir(packagePath)
		if err != nil {
			return err
		}

		for _, e := range entries {
			if e.IsDir() {
				s.ScanPackage(filepath.Join(packageName, e.Name()))
			}
		}

		return nil
	}

	// Read package.json
	pkgData, err := os.ReadFile(packageJsonPath)
	if err != nil {
		return fmt.Errorf("reading package.json: %w", err)
	}

	var pkg PackageJSON
	if err := json.Unmarshal(pkgData, &pkg); err != nil {
		return fmt.Errorf("parsing package.json: %w", err)
	}

	// Check for ES modules
	if pkg.Type == "module" || pkg.Module != "" || pkg.Browser != "" {
		// Add main module entry
		moduleEntry := pkg.Main
		if pkg.Browser != "" {
			moduleEntry = pkg.Browser
		}
		if pkg.Module != "" {
			moduleEntry = pkg.Module
		}
		if moduleEntry != "" {
			s.Imports[packageName] = filepath.Join(packageName, moduleEntry)
		}

		// Handle exports field
		if pkg.Exports != nil {
			if targetStr, ok := pkg.Exports.(string); ok {
				if strings.HasSuffix(targetStr, ".mjs") || strings.HasSuffix(targetStr, ".js") {
					s.Imports[packageName] = filepath.Join(packageName, targetStr)
				}
			} else if targetMap, ok := pkg.Exports.(map[string]interface{}); ok {
				for key, value := range targetMap {
					if key == "import" {
						if valueStr, ok := value.(string); ok {
							s.Imports[packageName] = filepath.Join(packageName, valueStr)
						}
					}
				}
			}
		}
	} else {
		for _, dir := range modulePackageDirs {
			dirPath := filepath.Join(packagePath, dir)
			if _, err := os.Stat(dirPath); err == nil {
				mainFile := filepath.Join(dir, "index.js")
				if _, err := os.Stat(filepath.Join(packagePath, mainFile)); err == nil {
					s.Imports[packageName] = filepath.Join(packageName, mainFile)
					break
				}
			}
		}
	}

	s.Scanned[packageName] = true

	s.ScanDependencies(pkg.Dependencies)

	return nil
}

func (s *Scanner) GetImports() map[string]string {
	return s.Imports
}

func (s *Scanner) ValidateImports() {
	for importName, importPath := range s.Imports {
		s.ProcessImport(importName, importPath)
	}
}

func (s *Scanner) ProcessImport(importName string, importPath string) {
	if s.Scanned[importPath] {
		return
	}

	s.Scanned[importPath] = true

	// load the file and read all the javascript module usedImports statements and if any of them are not in the s.Imports map, remove the importName from the s.Imports map
	usedImports, err := s.LoadImports(importPath)

	if err != nil {
		delete(s.Imports, importName)
		log.Printf("error loading imports for %s: %v", importName, err)
		return
	}

	for _, usedImport := range usedImports {
		if strings.HasPrefix(usedImport, "./") {
			usedImport = filepath.Join(filepath.Dir(importPath), usedImport)
			s.ProcessImport(importName, usedImport)
		} else if strings.HasPrefix(usedImport, "../") {
			usedImport = filepath.Join(filepath.Dir(importPath), usedImport)
			s.ProcessImport(importName, usedImport)
		} else {
			if _, ok := s.Imports[usedImport]; !ok {
				fullPath := filepath.Join(s.ModuleDir, usedImport)
				info, err := os.Stat(fullPath)
				if err != nil {
					info, err = os.Stat(fullPath + ".js")
					if err == nil && !info.IsDir() {
						s.ProcessImport(importName, usedImport+".js")
						s.Imports[usedImport] = usedImport + ".js"
						return
					}
					delete(s.Imports, importName)
					log.Printf("missing imports for %s: %s", importName, usedImport)
					return
				}
				if info.IsDir() {
					err = s.ScanPackage(usedImport)
					if err != nil {
						delete(s.Imports, importName)
						log.Printf("missing imports for %s: %s", importName, usedImport)
						return
					}
				} else {
					s.ProcessImport(importName, usedImport)
				}
			}
		}
	}
}

func (s *Scanner) LoadImports(importPath string) ([]string, error) {
	filePath := filepath.Join(s.ModuleDir, importPath)

	_, err := os.Stat(filePath)
	if err != nil && !strings.HasSuffix(importPath, ".js") {
		filePath = filepath.Join(s.ModuleDir, importPath+".js")
		if _, err = os.Stat(filePath); err != nil {
			err = s.ScanPackage(importPath)
			if err != nil {
				return nil, err
			}
			return nil, nil
		}
	}

	content, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	return matchImport(string(content), filePath)
}

func matchImport(content string, filePath string) ([]string, error) {
	singleLineCommentRegex := regexp.MustCompile(`(?mU)//[^\n]*\n`)
	slrm := singleLineCommentRegex.FindAllStringSubmatch(content, -1)
	for _, match := range slrm {
		content = strings.ReplaceAll(content, match[0], "")
	}
	// if strings.Contains(filePath, "acorn") {
	// 	os.WriteFile("acorn.txt.js", []byte(content), 0644)
	// }

	// strip all the javascript comments from the content
	multiListCommentRegex := regexp.MustCompile(`(?s)/\*[\s\S]*?\*/`)
	mlrm := multiListCommentRegex.FindAllStringSubmatch(content, -1)
	for _, match := range mlrm {
		content = strings.ReplaceAll(content, match[0], "")
	}

	exportOrImportRegex := regexp.MustCompile(`(export|import)\s+(?:(?:{[^}]*}|\*\s+as\s+\w+|\w+)\s+from\s+)?['"]([^'"]+)['"]`)
	matches := exportOrImportRegex.FindAllStringSubmatch(content, -1)

	imports := make([]string, 0)
	for _, match := range matches {
		if len(match) > 1 {
			imports = append(imports, match[2])
		}
	}

	return imports, nil
}
