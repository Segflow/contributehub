package checker

import (
	"go/ast"

	"github.com/segflow/contribuehub/pkg/codechange"
)

// Checker in the interface to be implemented by all checkers.
type Checker interface {
	SetPackages([]*ast.Package)
	CodeChanges() []codechange.CodeChange
}
