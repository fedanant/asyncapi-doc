package asyncapi

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"sort"
	"strings"
)

type file struct {
	file *ast.File
	name string
}

func sortedFiles(pkg *ast.Package) []file {
	files := make([]file, 0, len(pkg.Files))

	for name, f := range pkg.Files {
		files = append(files, file{file: f, name: filepath.Base(name)})
	}

	// Sort files in lexicographic order, except give priority to main.go
	sort.Slice(files, func(i, j int) bool {
		ni, nj := files[i].name, files[j].name
		if ni == "main.go" {
			return true
		}

		if nj == "main.go" {
			return false
		}

		return ni < nj
	})

	return files
}

func extractComment(cgrp *ast.CommentGroup) []string {
	s := cgrp.Text()
	comments := strings.Split(s, "\n")
	return comments
}

func parseComments(p *Parser, pkg *ast.Package) {
	files := sortedFiles(pkg)

	for _, f := range files {
		for _, c := range f.file.Comments {
			comments := extractComment(c)
			if isGeneralAPIComment(comments) {
				p.ParseMain(comments)
			} else {
				p.ParseOperation(comments, pkg)
			}
		}
	}
}

func isGeneralAPIComment(comments []string) bool {
	for _, commentLine := range comments {
		attribute := strings.ToLower(strings.Split(commentLine, " ")[0])
		switch attribute {
		case titleAttr, versionAttr, protocolAttr, urlAttr, hostAttr:
			return true
		}
	}
	return false
}

func ParseFolder(srcDir string, verbose bool) ([]byte, error) {
	// Validate that the source directory exists
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("source directory does not exist: %s", srcDir)
	}

	pathExec, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	fset := token.NewFileSet()

	pkgs := map[string]*ast.Package{}
	pkgsMain, err := parser.ParseDir(fset, srcDir, nil, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse directory %s: %w", srcDir, err)
	}

	for key, v := range pkgsMain {
		pkgs[key] = v
	}

	packagesFile, err := listPackages(srcDir, nil, "-deps")
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	for _, pkg := range packagesFile {
		filename := pkg.Dir
		if strings.HasPrefix(filename, pathExec) && pkgs[filename] == nil {
			packages, err := parser.ParseDir(fset, filename, nil, parser.ParseComments)
			if err != nil {
				return nil, fmt.Errorf("failed to parse package directory %s: %w", filename, err)
			}
			for _, pack := range packages {
				pkgs[filename] = pack
			}
		}
	}

	p := NewParser()

	if verbose {
		fmt.Printf("Parsing %d package(s)...\n", len(pkgs))
	}

	for _, pkg := range pkgs {
		if verbose {
			fmt.Printf("  - Parsing package: %s\n", pkg.Name)
		}
		parseComments(p, pkg)
	}

	// Validate that we found required API information
	if err := p.Validate(); err != nil {
		return nil, fmt.Errorf("validation failed: %w", err)
	}

	yaml, err := p.MarshalYAML()
	if err != nil {
		return nil, fmt.Errorf("failed to marshal YAML: %w", err)
	}

	if verbose {
		fmt.Printf("Generated %d channel(s) and %d operation(s)\n",
			len(p.asyncApi.Channels), len(p.asyncApi.Operations))
	}

	return yaml, nil
}

func Gen(filename string, outFile string) error {
	srcDir := filepath.Dir(filename)
	yaml, err := ParseFolder(srcDir, false)
	if err != nil {
		return fmt.Errorf("failed to parse folder: %w", err)
	}

	if err := os.WriteFile(outFile, yaml, 0o600); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
