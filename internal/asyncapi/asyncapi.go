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

func sortedFiles(files []*ast.File, fileNames map[*ast.File]string) []file {
	result := make([]file, 0, len(files))

	for _, f := range files {
		name := filepath.Base(fileNames[f])
		result = append(result, file{file: f, name: name})
	}

	// Sort files in lexicographic order, except give priority to main.go
	sort.Slice(result, func(i, j int) bool {
		ni, nj := result[i].name, result[j].name
		if ni == "main.go" {
			return true
		}

		if nj == "main.go" {
			return false
		}

		return ni < nj
	})

	return result
}

func extractComment(cgrp *ast.CommentGroup) []string {
	s := cgrp.Text()
	comments := strings.Split(s, "\n")
	return comments
}

func parseComments(p *Parser, files []file, tc *TypeChecker) {
	for _, f := range files {
		for _, c := range f.file.Comments {
			comments := extractComment(c)
			if isGeneralAPIComment(comments) {
				p.ParseMain(comments)
			} else {
				p.ParseOperation(comments, tc)
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

//nolint:gocyclo // Complex folder parsing logic is intentionally centralized
func ParseFolder(srcDir string, verbose bool, excludeDirs string) ([]byte, error) {
	// Validate that the source directory exists
	if _, err := os.Stat(srcDir); os.IsNotExist(err) {
		return nil, fmt.Errorf("source directory does not exist: %s", srcDir)
	}

	pathExec, err := os.Getwd()
	if err != nil {
		return nil, fmt.Errorf("failed to get working directory: %w", err)
	}
	fset := token.NewFileSet()

	// Parse excluded directories list
	excludeMap := make(map[string]bool)
	if excludeDirs != "" {
		for _, dir := range strings.Split(excludeDirs, ",") {
			excludeMap[strings.TrimSpace(dir)] = true
		}
	}

	// Create filter function to exclude directories
	filter := func(info os.FileInfo) bool {
		if info.IsDir() && excludeMap[info.Name()] {
			if verbose {
				fmt.Printf("Excluding directory: %s\n", info.Name())
			}
			return false
		}
		return true
	}

	// Parse all files in the directory
	pkgs, err := parser.ParseDir(fset, srcDir, filter, parser.ParseComments)
	if err != nil {
		return nil, fmt.Errorf("failed to parse directory %s: %w", srcDir, err)
	}

	// Collect all type checkers by package
	typeCheckers := make(map[string]*TypeChecker)

	for pkgName, pkg := range pkgs {
		// Convert ast.Package to []*ast.File
		var files []*ast.File
		for _, f := range pkg.Files {
			files = append(files, f)
		}

		tc, err := NewTypeChecker(fset, files, pkgName)
		if err != nil {
			if verbose {
				fmt.Printf("Warning: failed to create type checker for package %s: %v\n", pkgName, err)
			}
			continue
		}
		typeCheckers[pkgName] = tc
	}

	// Parse additional dependency packages
	packagesFile, err := listPackages(srcDir, nil, "-deps")
	if err != nil {
		return nil, fmt.Errorf("failed to list packages: %w", err)
	}

	for _, pkgInfo := range packagesFile {
		filename := pkgInfo.Dir
		if strings.HasPrefix(filename, pathExec) && typeCheckers[pkgInfo.Name] == nil {
			packages, err := parser.ParseDir(fset, filename, nil, parser.ParseComments)
			if err != nil {
				if verbose {
					fmt.Printf("Warning: failed to parse package directory %s: %v\n", filename, err)
				}
				continue
			}

			for pkgName, pkg := range packages {
				var files []*ast.File
				for _, f := range pkg.Files {
					files = append(files, f)
				}

				tc, err := NewTypeChecker(fset, files, pkgName)
				if err != nil {
					if verbose {
						fmt.Printf("Warning: failed to create type checker for package %s: %v\n", pkgName, err)
					}
					continue
				}
				typeCheckers[pkgName] = tc
			}
		}
	}

	p := NewParser()

	if verbose {
		fmt.Printf("Parsing %d package(s)...\n", len(pkgs))
	}

	// Parse comments from main packages
	for pkgName, pkg := range pkgs {
		if verbose {
			fmt.Printf("  - Parsing package: %s\n", pkgName)
		}

		tc := typeCheckers[pkgName]
		if tc == nil {
			if verbose {
				fmt.Printf("Warning: no type checker for package %s\n", pkgName)
			}
			continue
		}

		// Create file list with names
		var files []*ast.File
		fileNames := make(map[*ast.File]string)
		for name, f := range pkg.Files {
			files = append(files, f)
			fileNames[f] = name
		}

		sortedFileList := sortedFiles(files, fileNames)
		parseComments(p, sortedFileList, tc)
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
			len(p.asyncAPI.Channels), len(p.asyncAPI.Operations))
	}

	return yaml, nil
}

func Gen(filename, outFile string) error {
	srcDir := filepath.Dir(filename)
	yaml, err := ParseFolder(srcDir, false, "")
	if err != nil {
		return fmt.Errorf("failed to parse folder: %w", err)
	}

	if err := os.WriteFile(outFile, yaml, 0o600); err != nil {
		return fmt.Errorf("failed to write output file: %w", err)
	}

	return nil
}
