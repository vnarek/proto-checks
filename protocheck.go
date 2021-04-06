package protochecks

import (
	"fmt"
	"go/token"

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

type LaticeVals int

const (
	NN LaticeVals = iota
	Top
)

type MapLatice struct {
}

func runFunc(pass *analysis.Pass, fn *ssa.Function) {
	if fn.Name() != "SayHello" { //for testing purposes
		return
	}
	fmt.Println(fn.Signature.String())
	for _, ins := range fn.Blocks[0].Instrs {
		switch ins := ins.(type) {
		case *ssa.FieldAddr:
			fmt.Println("fieldAddr", ins.Field, ins.X)
		case *ssa.Alloc:
			fmt.Println("alloc", ins.Heap, ins.String())
		case *ssa.Field:
			fmt.Println("field", ins.Field, ins.X)
		case *ssa.UnOp:
			if ins.Op == token.MUL {
				fmt.Println("*", ins.X)
			}
		case *ssa.Call:
			fmt.Println("call", ins.Call.Value, ins.Call.Method)
		}
	}
}
