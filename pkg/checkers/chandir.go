package checker

import (
	"go/ast"
	"go/token"

	"github.com/segflow/contribuehub/pkg/codechange"
)

const (
	biDirectionalChan = ast.SEND | ast.RECV // 3
)

var (
	exampleFunc = `
	package test

	func A(a chan int) {
		f := func() {
			b := <-a
		}
	}
`
)

type ChanDirectionChecker struct {
	// funcsWithBidirChan holds the list of functions with bidirectional channels in params
	// Keys can be either *ast.FuncDecl or *ast.FuncLit, value is a map of parameterName -> *ast.Field
	funcsWithBidirChan map[ast.Node][]*ast.Field

	// localFuncCalls holds the list of callExpr
	localFuncCalls map[ast.Node][]*ast.CallExpr

	pkgs []*ast.Package
	fset *token.FileSet
}

type chanDirChange struct {
	ParameterName string
	NewDirection  ast.ChanDir
	Position      token.Position
	Offset        int
}

func (c chanDirChange) toCodeChange() codechange.CodeChange {
	delta := len(c.ParameterName) + 1
	if c.NewDirection == ast.SEND {
		delta += len("chan")
	}

	return codechange.CodeChange{
		Filename: c.Position.Filename,
		Line:     c.Position.Line,
		Column:   c.Position.Column + delta,
		Offset:   c.Position.Offset + delta,
		Add:      []byte("<-"),
	}
}

func NewChanDirectionChecker(fset *token.FileSet) *ChanDirectionChecker {
	return &ChanDirectionChecker{
		funcsWithBidirChan: make(map[ast.Node][]*ast.Field),
		localFuncCalls:     make(map[ast.Node][]*ast.CallExpr),
		fset:               fset,
	}
}

func (c *ChanDirectionChecker) CodeChanges() []codechange.CodeChange {
	// Step 1: Get all functions/methods with at least one bidirectional channels parameter
	for _, pkg := range c.pkgs {
		funcs := c.biDirChanFuncs(pkg)
		if len(funcs) == 0 {
			continue
		}
		for node, fields := range funcs {
			c.funcsWithBidirChan[node] = fields
		}
	}

	var reports []codechange.CodeChange
	for fn, params := range c.funcsWithBidirChan {
		chansUsage := c.funcsChanParamsUsage(fn, params)
		for field, usage := range chansUsage {
			if usage == biDirectionalChan {
				continue
			}

			// Channel is a send only channel
			if usage == ast.SEND {
				pos := field.Pos()
				reports = append(reports, chanDirChange{
					NewDirection:  ast.SEND,
					ParameterName: field.Names[0].Name,
					Position:      c.fset.Position(pos),
					Offset:        c.fset.Position(pos).Offset,
				}.toCodeChange())
			}

			// Channel is a recv only channel
			if usage == ast.RECV {
				pos := field.Pos()
				reports = append(reports, chanDirChange{
					NewDirection:  ast.RECV,
					ParameterName: field.Names[0].Name,
					Position:      c.fset.Position(pos),
					Offset:        c.fset.Position(pos).Offset,
				}.toCodeChange())
			}
		}
	}

	return reports
}

func (c *ChanDirectionChecker) SetPackages(pkgs []*ast.Package) {
	c.pkgs = pkgs
}

// biDirChanParams returns the names of bi directional channel found in func param.
func (c *ChanDirectionChecker) biDirChanParams(fn *ast.FuncDecl) []*ast.Field {
	params := fn.Type.Params.List
	var biDirChans []*ast.Field
	for _, param := range params {
		t, ok := param.Type.(*ast.ChanType) // If not a chan type we ignore it
		if !ok {
			continue
		}

		if t.Dir != biDirectionalChan {
			continue
		}

		if len(param.Names) == 0 { // unnamed chan param
			continue
		}

		biDirChans = append(biDirChans, param)
	}

	return biDirChans
}

func (c *ChanDirectionChecker) biDirChanFuncs(pkg *ast.Package) map[*ast.FuncDecl][]*ast.Field {
	bidirchanFuncs := make(map[*ast.FuncDecl][]*ast.Field)

	for _, file := range pkg.Files {
		for _, decl := range file.Decls {
			fn, ok := decl.(*ast.FuncDecl) // We only care for functions declaration
			if !ok {
				continue
			}

			chans := c.biDirChanParams(fn)
			if len(chans) == 0 {
				continue
			}

			bidirchanFuncs[fn] = chans
		}
	}

	return bidirchanFuncs
}

func objInParams(obj *ast.Object, params []*ast.Field) bool {
	if obj == nil {
		return false
	}

	field, ok := obj.Decl.(*ast.Field)
	if !ok {
		return false
	}

	return fieldInParams(field, params)
}

func fieldInParams(field *ast.Field, params []*ast.Field) bool {
	for _, f := range params {
		if field == f {
			return true
		}
	}
	return false
}

// paramsUsedArgs returns the list of params used directly or inderectly by args.
func paramsUsedInArgs(params []*ast.Field, args []ast.Expr) []*ast.Field {
	var fields []*ast.Field

	walk := func(node ast.Node) bool {
		if id, ok := node.(*ast.Ident); ok && objInParams(id.Obj, params) {
			field := id.Obj.Decl.(*ast.Field)
			fields = append(fields, field)
		}
		return true
	}

	for _, arg := range args {
		ast.Inspect(arg, walk)
	}

	return fields
}

// funcsChanParamsUsage returns a mapping of how chan parameters are being used inside function fn.
func (c *ChanDirectionChecker) funcsChanParamsUsage(fn ast.Node, params []*ast.Field) map[*ast.Field]ast.ChanDir {
	m := make(map[*ast.Field]ast.ChanDir)

	walkFunc := func(node ast.Node) bool {
		// Mark any channel parameter used in function call as bidirectional
		if callExpr, ok := node.(*ast.CallExpr); ok {
			for _, field := range paramsUsedInArgs(params, callExpr.Args) {
				if fieldInParams(field, params) {
					m[field] = m[field] | biDirectionalChan
				}
			}
		}

		// Send to channel
		if sendStmt, ok := node.(*ast.SendStmt); ok {
			if id, ok := sendStmt.Chan.(*ast.Ident); ok && objInParams(id.Obj, params) { // We only care when the channel is an identifier
				field := id.Obj.Decl.(*ast.Field)
				m[field] = m[field] | ast.SEND
			}
		}

		// Read from channel
		if unaryExpr, ok := node.(*ast.UnaryExpr); ok {
			op := unaryExpr.Op.String()
			if op == "<-" {
				for _, id := range unaryExprReadChannels(unaryExpr) {
					if objInParams(id.Obj, params) {
						field := id.Obj.Decl.(*ast.Field)
						m[field] = m[field] | ast.RECV
					}
				}
			}
		}

		// Range over a channel
		if rngStmt, ok := node.(*ast.RangeStmt); ok {
			if ident, ok := rngStmt.X.(*ast.Ident); ok && objInParams(ident.Obj, params) {
				field := ident.Obj.Decl.(*ast.Field)
				m[field] = m[field] | ast.RECV
			}
		}

		// Close of a channel
		if callExpr, ok := node.(*ast.CallExpr); ok && isBuiltinCloseCall(callExpr) {
			if ident, ok := callExpr.Args[0].(*ast.Ident); ok && objInParams(ident.Obj, params) {
				field := ident.Obj.Decl.(*ast.Field)
				m[field] = m[field] | biDirectionalChan
			}
		}

		return true
	}

	fnBody := fn.(*ast.FuncDecl).Body
	ast.Inspect(fnBody, walkFunc)

	return m
}

// isBuiltinCloseCall checks if the call to `close` is a call to the builtin `call` function
func isBuiltinCloseCall(call *ast.CallExpr) bool {
	ident, ok := call.Fun.(*ast.Ident)
	if !ok { // Maybe anon function call
		return false
	}

	return ident.Name == "close" && ident.Obj == nil // if obj is not nil, then close is overwritten by another one.
}

// unaryExprReadChannels returns the list of channel params a read (<-) unary expression depends on.
func unaryExprReadChannels(unaryExpr *ast.UnaryExpr) []*ast.Ident {
	switch x := unaryExpr.X.(type) {
	case *ast.Ident:
		// e.g: `<-channel`
		return []*ast.Ident{x}
	case *ast.CallExpr:
		// e.g: `<-fn()`
		// todo: handle
	}
	return nil
}
