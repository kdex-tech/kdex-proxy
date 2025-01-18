package main

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"kdex.dev/proxy/internal/scanner"
)

func main() {
	rootDir := os.Getenv("MODULE_DIR")
	if rootDir == "" {
		fmt.Println("MODULE_DIR is not set")
		return
	}

	s := scanner.NewScanner(rootDir)
	pkgData, err := os.ReadFile(filepath.Join(rootDir, "package.json"))
	if err != nil {
		fmt.Println(err)
		return
	}

	var pkg scanner.PackageJSON
	if err := json.Unmarshal(pkgData, &pkg); err != nil {
		fmt.Println(err)
		return
	}

	// Scan all dependencies
	if err := s.ScanDependencies(pkg.Dependencies); err != nil {
		fmt.Println(err)
		return
	}

	// Generate import map
	imports := s.GenerateImports()

	bytes, err := json.MarshalIndent(imports, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%s\n", bytes)
}
