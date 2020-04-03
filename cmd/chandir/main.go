package main

import (
	"encoding/json"
	"fmt"
	"go/token"
	"log"
	"os"

	"github.com/segflow/contribuehub/pkg/ast"
	checker "github.com/segflow/contribuehub/pkg/checkers"
	"github.com/segflow/contribuehub/pkg/codechange"
	"github.com/spf13/cobra"
)

var rootCmd = &cobra.Command{
	Use:   "chandir PACKAGE",
	Short: "Check channel direction usage in a Go package.",
	Run:   chandircheck,
}

type result struct {
	Count   int
	Changes map[string][]codechange.CodeChange
}

func chandircheck(cmd *cobra.Command, args []string) {
	if len(args) != 1 {
		cmd.Usage()
		os.Exit(2)
	}

	fset := token.NewFileSet()
	checker := checker.NewChanDirectionChecker(fset)
	pkgs := ast.ParseDirPackages(fset, args[0])
	if len(pkgs) == 0 {
		log.Fatalf("No packages found in %q", args[0])
	}

	checker.SetPackages(pkgs)
	reports := checker.CodeChanges()

	changes := make(map[string][]codechange.CodeChange)
	for _, report := range reports {
		filename := report.Filename
		changes[filename] = append(changes[filename], report)
	}

	result := result{
		Count:   len(reports),
		Changes: changes,
	}

	if err := json.NewEncoder(os.Stdout).Encode(result); err != nil {
		log.Fatalf("Error encoding result: %v", err)
	}

	apply := cmd.Flags().Lookup("apply").Value.String() == "true"
	if apply {
		err := applyChanges(changes)
		if err != nil {
			log.Fatal(err)
		}
	}

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

func main() {
	if err := rootCmd.Execute(); err != nil {
		fmt.Println(err)
		os.Exit(1)
	}
}

func init() {
	rootCmd.PersistentFlags().Bool("apply", false, "apply changes")
}
