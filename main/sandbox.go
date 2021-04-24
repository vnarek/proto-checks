package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	protochecks "github.com/vnarek/proto-checks"
	"golang.org/x/tools/go/cfg"
)

func succPrint(n protochecks.Node) {
	fmt.Printf("%#v\n", n.AST())
	fmt.Println("childs")
	for k := range n.Succ() {
		if _, ok := k.Pred()[n]; !ok {
			panic("panic")
		}
		succPrint(k)
	}
	fmt.Println("end child")
}

func main() {
	// src is the input for which we want to print the AST.
	src := `
package main

func main() { 
	if (true) {
	  x := &y
	}
	y := &z
	z := &a
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
	b := protochecks.NewBuilder()
	start := protochecks.NewStartNode()
	b.BlockToNode(c.Blocks[0], start)
	succPrint(start)

	fmt.Println(c.Format(fset))
	fmt.Println("==================")
	// Print the AST.
	ast.Print(fset, f)
}
