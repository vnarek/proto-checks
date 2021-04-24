package main

import (
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"strings"

	protochecks "github.com/vnarek/proto-checks"
	"golang.org/x/tools/go/cfg"
)

func printNodes(n protochecks.Node) {
	succPrint(n, 0, make(map[protochecks.Node]struct{}))
}

func succPrint(n protochecks.Node, depth int, printed map[protochecks.Node]struct{}) {
	printed[n] = struct{}{}
	fmt.Print(strings.Repeat("  ", depth))
	fmt.Println(protochecks.ToString(n))
	for k := range n.Succ() {
		if _, ok := k.Pred()[n]; !ok {
			panic("panic")
		}
		if _, ok := printed[k]; ok {
			fmt.Print(strings.Repeat("  ", depth + 1))
			fmt.Println("[connects to: " + protochecks.ToString(k) + "]")
		} else {
			succPrint(k, depth + 1, printed)
		}
	}
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
	for (false) {
		a := *z
	}
	z := nil
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
	printNodes(start)

	//fmt.Println(c.Format(fset))
	//fmt.Println("==================")
	// Print the AST.
	//ast.Print(fset, f)
}
