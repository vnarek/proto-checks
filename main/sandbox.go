package main

import (
	"go/ast"
	"go/parser"
	"go/token"

	normalizeCfg "github.com/vnarek/proto-checks/cfg"
	"golang.org/x/tools/go/cfg"
)

func main() {
	// src is the input for which we want to print the AST.
	src := `
package main

func main() {
	var int *x = new(1)
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

	c := cfg.New(
		funDecl.Body,
		func(ce *ast.CallExpr) bool { return true },
	)
	b := normalizeCfg.NewBuilder()
	start := normalizeCfg.NewStartNode()
	b.BlockToNode(c.Blocks[0], start)
	normalizeCfg.PrintNodes(start)

	//fmt.Println(c.Format(fset))
	//fmt.Println("==================")
	// Print the AST.
	//ast.Print(fset, f)
}
