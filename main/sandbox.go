package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"

	normalizeCfg "github.com/vnarek/proto-checks/cfg"
	"github.com/vnarek/proto-checks/nilness"
)

func main() {
	// src is the input for which we want to print the AST.
	src := `
package main

func main() {
	p1 = &a;
	p2 = &b;
	p1 = p2;
	r = &p1;
	*r = &c
	p3 = *r;
	p2 = &d;
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
	b.Build(funDecl)
	start := b.GetCfg()
	normalizeCfg.PrintNodes(start)

	nodes := b.Nodes()
	for _, n := range nodes {
		println(normalizeCfg.ToString(n))
	}

	res := nilness.Build(b)
	fmt.Println(res)

	//fmt.Println(c.Format(fset))
	//fmt.Println("==================")
	// Print the AST.
	//ast.Print(fset, f)
}
