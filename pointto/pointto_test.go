package pointto

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	normalizeCfg "github.com/vnarek/proto-checks/cfg"
	"golang.org/x/tools/go/cfg"
)

func TestBookCase(t *testing.T) {
	tt := `
package main

func main() {
	p := new(int)
	x := y
	x := z
	*p = z
	p = q
	q = &y
	x = *p
	p = &z
}`
	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "", tt, 0)
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

	sol := Build(b)

	t.Logf("%#v\n", sol)
}

func TestCycleCase(t *testing.T) {
	tt := `
package main

func main() {
	p := new(int)
	x := p
	p = x
}`
	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "", tt, 0)
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

	sol := Build(b)

	t.Logf("%#v\n", sol)
}

//source: http://pages.cs.wisc.edu/~fischer/cs701.f08/lectures/Lecture26.4up.pdf
func TestWiscCase(t *testing.T) {
	tt := `
package main

func main() {
	p1 = &a;
	p2 = &b;
	p1 = p2;
	r = &p1;
	*r = &c
	p3 = *r;
	p2 = &d;
}`
	// Create the AST by parsing src.
	fset := token.NewFileSet() // positions are relative to fset
	f, err := parser.ParseFile(fset, "", tt, 0)
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

	sol := Build(b)

	/* Should return:
	 * p1 = a, b, c, d
	 * p2 = b, d
	 * p3 = a, b, c, d
	 * r = p1
	 */
	t.Logf("%#v\n", sol)
}
