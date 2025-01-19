package scanner

import (
	"reflect"
	"testing"
)

func TestScanner_ScanRootDir(t *testing.T) {
	type fields struct {
		ModuleDir string
		Imports   map[string]string
		Scanned   map[string]bool
	}
	tests := []struct {
		name    string
		fields  fields
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scanner{
				ModuleDir: tt.fields.ModuleDir,
				Imports:   tt.fields.Imports,
				Scanned:   tt.fields.Scanned,
			}
			if err := s.ScanRootDir(); (err != nil) != tt.wantErr {
				t.Errorf("Scanner.ScanRootDir() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScanner_ScanDependencies(t *testing.T) {
	type fields struct {
		ModuleDir string
		Imports   map[string]string
		Scanned   map[string]bool
	}
	type args struct {
		dependencies PackageDependencies
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scanner{
				ModuleDir: tt.fields.ModuleDir,
				Imports:   tt.fields.Imports,
				Scanned:   tt.fields.Scanned,
			}
			if err := s.ScanDependencies(tt.args.dependencies); (err != nil) != tt.wantErr {
				t.Errorf("Scanner.ScanDependencies() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestNewScanner(t *testing.T) {
	type args struct {
		moduleDir string
	}
	tests := []struct {
		name string
		args args
		want *Scanner
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := NewScanner(tt.args.moduleDir); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("NewScanner() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanner_ScanPackage(t *testing.T) {
	type fields struct {
		ModuleDir string
		Imports   map[string]string
		Scanned   map[string]bool
	}
	type args struct {
		packageName string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scanner{
				ModuleDir: tt.fields.ModuleDir,
				Imports:   tt.fields.Imports,
				Scanned:   tt.fields.Scanned,
			}
			if err := s.ScanPackage(tt.args.packageName); (err != nil) != tt.wantErr {
				t.Errorf("Scanner.ScanPackage() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestScanner_GetImports(t *testing.T) {
	type fields struct {
		ModuleDir string
		Imports   map[string]string
		Scanned   map[string]bool
	}
	tests := []struct {
		name   string
		fields fields
		want   map[string]string
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scanner{
				ModuleDir: tt.fields.ModuleDir,
				Imports:   tt.fields.Imports,
				Scanned:   tt.fields.Scanned,
			}
			if got := s.GetImports(); !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Scanner.GetImports() = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestScanner_ValidateImports(t *testing.T) {
	type fields struct {
		ModuleDir string
		Imports   map[string]string
		Scanned   map[string]bool
	}
	tests := []struct {
		name   string
		fields fields
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scanner{
				ModuleDir: tt.fields.ModuleDir,
				Imports:   tt.fields.Imports,
				Scanned:   tt.fields.Scanned,
			}
			s.ValidateImports()
		})
	}
}

func TestScanner_LoadImports(t *testing.T) {
	type fields struct {
		ModuleDir string
		Imports   map[string]string
		Scanned   map[string]bool
	}
	type args struct {
		importPath string
	}
	tests := []struct {
		name    string
		fields  fields
		args    args
		want    []string
		wantErr bool
	}{
		// TODO: Add test cases.
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			s := &Scanner{
				ModuleDir: tt.fields.ModuleDir,
				Imports:   tt.fields.Imports,
				Scanned:   tt.fields.Scanned,
			}
			got, err := s.LoadImports(tt.args.importPath)
			if (err != nil) != tt.wantErr {
				t.Errorf("Scanner.LoadImports() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("Scanner.LoadImports() = %v, want %v", got, tt.want)
			}
		})
	}
}
