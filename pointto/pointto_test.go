package pointto

import (
	"go/ast"
	"go/parser"
	"go/token"
	"testing"

	normalizeCfg "github.com/vnarek/proto-checks/cfg"
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
	ref := make(Result)
	ref["p"] = []string{"alloc-1", "y", "z"}
	ref["q"] = []string{"y"}

	runTest(t, tt, ref)
}

func TestCycleCase(t *testing.T) {
	tt := `
package main

func main() {
	p := new(int)
	x := p
	p = x
}`
	ref := make(Result)
	ref["p"] = []string{"alloc-1"}
	ref["x"] = []string{"alloc-1"}

	runTest(t, tt, ref)
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
	ref := make(Result)
	ref["p1"] = []string{"a", "b", "c", "d"}
	ref["p2"] = []string{"b", "d"}
	ref["p3"] = []string{"a", "b", "c", "d"}
	ref["r"] = []string{"p1"}
	ref["_t1"] = []string{"c"}

	runTest(t, tt, ref)
}

func runTest(t *testing.T, tt string, ref Result) {
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

	b := normalizeCfg.NewBuilder()
	b.Build(funDecl)

	sol := Build(b)

	//check sizes of solution
	if len(sol) != len(ref) {
		t.Logf("solution size: %d", len(sol))
		t.Logf("ref solution size: %d", len(ref))
		t.Fail()
		return
	}

	//for each variable in ref solution
	for k, v := range ref {
		if len(sol[k]) != len(v) {
			t.Logf("solution size of variable '%s': %d", k, len(sol[k]))
			t.Logf("ref solution size of variable '%s': %d", k, len(v))
			t.Fail()
			return
		}

		//for each variable v points to
		for _, target := range v {
			if !contains(sol[k], target) {
				t.Logf("missing '%s points to %s' relation", k, target)
				t.Fail()
				return
			}
		}
	}
}

func contains(hay []string, needle string) bool {
	for _, v := range hay {
		if v == needle {
			return true
		}
	}
	return false
}