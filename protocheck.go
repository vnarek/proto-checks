package protochecks

import (
	"github.com/vnarek/proto-checks/testdata/src/basic"
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
	"golang.org/x/tools/go/ssa"
)

var Analyzer = &analysis.Analyzer{
	Name: "protochecks",
	Doc:  "check usage of protocol buffers v1 api in go packages",
	Run: func(p *analysis.Pass) (interface{}, error) {
		a := analyzer{}
		return a.run(p)
	},
	Requires: []*analysis.Analyzer{
		buildssa.Analyzer,
	},
}

var _ = basic.File_basic_basic_proto

//TODO: implement the analyzer
type analyzer struct {
}

func (a *analyzer) run(pass *analysis.Pass) (interface{}, error) {
	ssainput := pass.ResultOf[buildssa.Analyzer].(*buildssa.SSA)
	for _, fn := range ssainput.SrcFuncs {
		runFunc(pass, fn)
	}
	return nil, nil
}

func runFunc(pass *analysis.Pass, fn *ssa.Function) {
}
