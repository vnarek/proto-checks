package main

import (
	"fmt"
	normalizeCfg "github.com/vnarek/proto-checks/cfg"
	"github.com/vnarek/proto-checks/nilness"
	"go/ast"
	"go/parser"
	"go/token"
	"golang.org/x/tools/go/cfg"
)

func main() {
	// src is the input for which we want to print the AST.
	src := `
package main

func main(in *int) {
	r := nil
	if r != nil {
		q = &r
	}
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
	b.Build(c.Blocks[0])
	start := b.GetCfg()
	normalizeCfg.PrintNodes(start)
/*
	nodes := b.Nodes()
	for _, n := range nodes {
		println(normalizeCfg.ToString(n))
	}*/

	res := nilness.Build(b)

	fmt.Printf("%#v\n", res)

	//fmt.Println(c.Format(fset))
	//fmt.Println("==================")
	// Print the AST.
	//ast.Print(fset, f)
}
