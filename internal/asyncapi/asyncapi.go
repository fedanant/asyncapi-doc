package asyncapi

import (
	"go/ast"
	"go/parser"
	"go/token"
	"log"
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
		// The @url annotation belongs to Operation
		case urlAttr:
			return true
		}
	}
	return false
}

func ParseFolder(srcDir string) ([]byte, error) {
	pathExec, err := os.Getwd()
	if err != nil {
		log.Panicln(err)
	}
	fset := token.NewFileSet()

	pkgs := map[string]*ast.Package{}
	pkgsMain, err := parser.ParseDir(fset, srcDir, nil, parser.ParseComments)
	if err != nil {
		panic(err)
	}

	for key, v := range pkgsMain {
		pkgs[key] = v
	}

	packagesFile, err := listPackages(srcDir, nil, "-deps")

	if err != nil {
		panic(err)
	}

	for _, pkg := range packagesFile {
		filename := pkg.Dir
		if strings.HasPrefix(filename, pathExec) && pkgs[filename] == nil {
			packages, err := parser.ParseDir(fset, filename, nil, parser.ParseComments)
			if err != nil {
				log.Panicln(err)
			}
			for _, pack := range packages {
				pkgs[filename] = pack
			}
		}
	}

	p := NewParser()

	for _, pkg := range pkgs {
		parseComments(p, pkg)
	}

	yaml, err := p.MarshalYAML()
	if err != nil {
		log.Fatalln(err)
	}

	return yaml, err
}

func Gen(filename string, outFile string) {
	srcDir := filepath.Dir(filename)
	yaml, err := ParseFolder(srcDir)
	if err != nil {
		log.Fatalln(err)
	}
	os.WriteFile(outFile, yaml, 0o600)
}
