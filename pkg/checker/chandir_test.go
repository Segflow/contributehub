package checker

import (
	"go/token"
	"testing"

	"github.com/davecgh/go-spew/spew"
	"github.com/segflow/contribuehub/pkg/ast"
	"github.com/stretchr/testify/assert"
)

func TestRecvOnlyChannel(t *testing.T) {
	code := `
	package test
	func A(a chan int) {
		a <- 2
	} 
	`

	checker := NewChanDirectionChecker(token.NewFileSet())
	checker.SetPackages(ast.PackagesFromCode(code))

	// if
	reports := checker.CodeChanges()

	// then
	assert.Len(t, reports, 1)
}

func TestSendOnlyChannel(t *testing.T) {
	code := `
	package test
	func A(a chan int) {
		b := <-a
	}
	`

	checker := NewChanDirectionChecker(token.NewFileSet())
	checker.SetPackages(ast.PackagesFromCode(code))

	// if
	reports := checker.CodeChanges()

	// then
	assert.Len(t, reports, 1)
}

func TestSendRecvChannel(t *testing.T) {
	code := `
	package test
	func A(a chan int) {
		a <- 2
		b := <-a
	}
	`

	checker := NewChanDirectionChecker(token.NewFileSet())
	checker.SetPackages(ast.PackagesFromCode(code))

	// if
	reports := checker.CodeChanges()

	// then
	assert.Len(t, reports, 0)
}

func TestReadCloseChannel(t *testing.T) {
	code := `
	package test
	func A(a chan int) {
		b := <-1
		close(a)
	}
	`

	checker := NewChanDirectionChecker(token.NewFileSet())
	checker.SetPackages(ast.PackagesFromCode(code))

	// if
	reports := checker.CodeChanges()

	// then
	assert.Len(t, reports, 0)
}

func TestReadCustomCloseChannel(t *testing.T) {
	// Since we mark any channel used in any local function call as bidrectional for now,
	// any close(CH) will mark CH as bidirectional
	code := `
	package test

	func close(interface {}) {}
	func A(a chan int) {
		a <- 2
		close(a)
	}
	`

	checker := NewChanDirectionChecker(token.NewFileSet())
	checker.SetPackages(ast.PackagesFromCode(code))

	// if
	reports := checker.CodeChanges()

	// then
	assert.Len(t, reports, 0)
}

func TestSelectChannel(t *testing.T) {
	code := `
	package test

	func A(a chan int) {
		select {
		case <-a:
		}
	}
	`

	checker := NewChanDirectionChecker(token.NewFileSet())
	checker.SetPackages(ast.PackagesFromCode(code))

	// if
	reports := checker.CodeChanges()

	// then
	assert.Len(t, reports, 1)
}

func TestRangeChannel(t *testing.T) {
	code := `
	package test

	func A(a chan int) {
		for b := range a {}
	}
	`

	checker := NewChanDirectionChecker(token.NewFileSet())
	checker.SetPackages(ast.PackagesFromCode(code))

	// if
	reports := checker.CodeChanges()

	// then
	assert.Len(t, reports, 1)
}

func TestNotUsedChannel(t *testing.T) {
	code := `
	package test

	func A(a chan int) {
		
	}
	`

	checker := NewChanDirectionChecker(token.NewFileSet())
	checker.SetPackages(ast.PackagesFromCode(code))

	// if
	reports := checker.CodeChanges()

	// then
	assert.Len(t, reports, 0)
}

func TestClosureFuncSendChannel(t *testing.T) {
	code := `
	package test

	func A(a chan int) {
		func (){
			a <- 2
		}()
	}
	`

	checker := NewChanDirectionChecker(token.NewFileSet())
	checker.SetPackages(ast.PackagesFromCode(code))

	// if
	reports := checker.CodeChanges()
	spew.Dump(reports)

	// then
	assert.Len(t, reports, 1)
}

func TestClosureFuncSendRecvChannel(t *testing.T) {
	code := `
	package test

	func A(a chan int) {
		func (){
			a <- 2
		}
		v := <- a
	}
	`

	checker := NewChanDirectionChecker(token.NewFileSet())
	checker.SetPackages(ast.PackagesFromCode(code))

	// if
	reports := checker.CodeChanges()
	spew.Dump(reports)

	// then
	assert.Len(t, reports, 0)
}

func TestChanInFuncCallChannel(t *testing.T) {
	code := `
	package test

	func A(a chan int) {
		func (b chan int){
			b <- 2
		}(a)
		v := <- a
	}
	`

	checker := NewChanDirectionChecker(token.NewFileSet())
	checker.SetPackages(ast.PackagesFromCode(code))

	// if
	reports := checker.CodeChanges()
	spew.Dump(reports)

	// then
	assert.Len(t, reports, 0)
}
