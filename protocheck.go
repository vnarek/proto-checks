package protochecks

import (
	"fmt"
	"go/ast"

	"github.com/vnarek/proto-checks/testdata/src/basic"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/ctrlflow"
	"golang.org/x/tools/go/cfg"
)

var Analyzer = &analysis.Analyzer{
	Name: "protochecks",
	Doc:  "check usage of protocol buffers v1 api in go packages",
	Run: func(p *analysis.Pass) (interface{}, error) {
		a := analyzer{}
		return a.run(p)
	},
	Requires: []*analysis.Analyzer{
		ctrlflow.Analyzer,
	},
}

var _ = basic.File_basic_basic_proto

//TODO: implement the analyzer
type analyzer struct {
}

func (a *analyzer) run(pass *analysis.Pass) (interface{}, error) {
	cfg := pass.ResultOf[ctrlflow.Analyzer].(*ctrlflow.CFGs)
	for _, f := range pass.Files {
		for _, decl := range f.Decls {
			switch d := decl.(type) {
			case *ast.FuncDecl:
				if d.Name.String() != "SayHello" {
					break
				}
				runFunc(pass, cfg.FuncDecl(d))
			}
		}
	}
	return nil, nil
}

func runFunc(pass *analysis.Pass, cfg *cfg.CFG) {
	fmt.Println(cfg.Format(pass.Fset))
}
