package main

import (
	"go/ast"
	"go/parser"
	"go/token"

	normalizeCfg "github.com/vnarek/proto-checks/cfg"
)

func main() {
	// src is the input for which we want to print the AST.
	src := `
package main

func main() {
	var x *int = nil
	if 2+2 == 4 {
		y := new(int)
		var z **int = &y
		x = *z
	} else {
		x = new(int)
	}
	z := x
}
`

	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "", src, 0)
	if err != nil {
		panic(err)
	}

	funDecl, ok := f.Decls[0].(*ast.FuncDecl)
	if !ok {
		panic("not funDecl")
	}

	b := normalizeCfg.NewBuilder()
	nodes := b.GetNodes(funDecl.Body)
	normalizeCfg.Print(nodes)

	// Print the AST.
	//ast.Print(fset, f)
}
