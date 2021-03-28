package protochecks

import (
	"golang.org/x/tools/go/analysis"
	"golang.org/x/tools/go/analysis/passes/buildssa"
)

var Analyer = &analysis.Analyzer{
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

//TODO: implement the analyzer
type analyzer struct {
}

func (a *analyzer) run(p *analysis.Pass) (interface{}, error) {
	return nil, nil
}
