package main

import (
	"encoding/json"
	"fmt"
	"os"

	"kdex.dev/proxy/internal/scanner"
)

func main() {
	rootDir := os.Getenv("MODULE_DIR")
	if rootDir == "" {
		fmt.Println("MODULE_DIR is not set")
		return
	}

	s := scanner.NewScanner(rootDir)

	if err := s.ScanRootDir(); err != nil {
		fmt.Println(err)
		return
	}

	s.ValidateImports()

	// Generate import map
	imports := s.GetImports()

	bytes, err := json.MarshalIndent(imports, "", "  ")
	if err != nil {
		fmt.Println(err)
		return
	}

	fmt.Printf("%s\n", bytes)
}
