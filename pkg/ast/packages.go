package ast

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"log"
	"os"
	"path/filepath"
)

func PackagesFromCode(codes ...string) []*ast.Package {
	fset := token.NewFileSet()

	var pkgs []*ast.Package
	for i, code := range codes {
		fname := fmt.Sprintf("filename-%d", i)
		file, _ := parser.ParseFile(fset, fname, code, parser.ParseComments)
		pkg, _ := ast.NewPackage(fset, map[string]*ast.File{
			"file": file,
		}, nil, nil)

		pkgs = append(pkgs, pkg)
	}

	return pkgs
}

func ParseDirPackages(fset *token.FileSet, dir string) []*ast.Package {
	var allPkgs []*ast.Package
	var walk = func(path string, info os.FileInfo, err error) error {
		if err != nil {
			fmt.Printf("error %v at a path %q\n", err, path)
			return err
		}

		if !info.IsDir() {
			return nil
		}

		lastDir := filepath.Base(path)
		if lastDir == "testdata" || lastDir == "vendor" {
			return filepath.SkipDir
		}

		pkgs, err := parser.ParseDir(fset, path, nil, parser.ParseComments)
		if err != nil {
			log.Printf("error parsing dir %q: %s\n", dir, err)
			return nil
		}

		for _, pkg := range pkgs {
			allPkgs = append(allPkgs, pkg)
		}

		return nil
	}

	filepath.Walk(dir, walk)
	return allPkgs
}
