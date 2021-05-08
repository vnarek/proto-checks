package protochecks

import (
	"fmt"
	"go/ast"

	"github.com/vnarek/proto-checks/cfg"
	"github.com/vnarek/proto-checks/nilness"
	"github.com/vnarek/proto-checks/testdata/src/basic"
	"golang.org/x/tools/go/analysis"
)

var Analyzer = &analysis.Analyzer{
	Name: "protochecks",
	Doc:  "check usage of protocol buffers v1 api in go packages",
	Run: func(p *analysis.Pass) (interface{}, error) {
		a := analyzer{}
		return a.run(p)
	},
	Requires: []*analysis.Analyzer{},
}

var _ = basic.File_basic_basic_proto

//TODO: implement the analyzer
type analyzer struct {
}

func (a *analyzer) run(pass *analysis.Pass) (interface{}, error) {
	for _, f := range pass.Files {
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				b := cfg.NewBuilder()
				b.Build(d)

				a.check(pass, b.Nodes(), nilness.Build(b))
			}
		}
	}
	return nil, nil
}

func (a *analyzer) check(pass *analysis.Pass, nodes []cfg.Node, cfgMap nilness.CfgMap) {
	for _, n := range nodes {
		switch n := n.(type) {
		case *cfg.DerefNode:
			if cfgMap.Get(n).Get(n.Lhs) == nilness.PN {
				report(pass, n, fmt.Sprintf("%s could be nil", n.Lhs))
			}
		case *cfg.PointerNode:
			if cfgMap.Get(n).Get(n.Rhs) == nilness.PN {
				report(pass, n, fmt.Sprintf("%s could be nil", n.Rhs))
			}
		case *cfg.SingleDerefNode:
			if cfgMap.Get(n).Get(n.Lhs) == nilness.PN {
				report(pass, n, fmt.Sprintf("%s could be nil", n.Lhs))
			}
		}
	}
}

func report(pass *analysis.Pass, n cfg.Node, msg string) {
	pass.Report(analysis.Diagnostic{
		Pos:      n.AST().Pos(),
		End:      n.AST().End(),
		Category: "propable nil dereference",
		Message:  msg,
	})
}
