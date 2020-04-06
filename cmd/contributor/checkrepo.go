package main

import (
	"fmt"
	"go/token"
	"log"

	"github.com/segflow/contribuehub/pkg/ast"
	"github.com/segflow/contribuehub/pkg/checker"
	"github.com/segflow/contribuehub/pkg/codechange"
	"github.com/segflow/contribuehub/pkg/repository"
)

func repoProcessChanDirection(repo *repository.Repository) (int, error) {
	fset := token.NewFileSet()
	checker := checker.NewChanDirectionChecker(fset)
	pkgs := ast.ParseDirPackages(fset, repo.LocalDirectory)
	if len(pkgs) == 0 {
		log.Printf("No packages found in %q", repo.LocalDirectory)
		return 0, nil
	}

	checker.SetPackages(pkgs)
	reports := checker.CodeChanges()

	changes := make(map[string][]codechange.CodeChange)
	for _, report := range reports {
		filename := report.Filename
		changes[filename] = append(changes[filename], report)
	}

	if len(changes) != 0 {
		fmt.Printf("Applying %d changes to %q\n", len(changes), repo.LocalDirectory)
	}

	err := applyChanges(changes)
	if err != nil {
		return 0, err
	}

	return len(changes), nil
}

func applyChanges(changes map[string][]codechange.CodeChange) error {

	for filename, fchanges := range changes {
		err := codechange.FileApplyChangesInplace(filename, fchanges)
		if err != nil {
			return err
		}
	}
	return nil
}
