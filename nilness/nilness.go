package nilness

import (
	"fmt"
	"github.com/vnarek/proto-checks/cfg"
	"github.com/vnarek/proto-checks/pointto"
)

type Result = CfgMap

func Build(build *cfg.Builder) Result {
	start := build.GetCfg()

	pt := pointto.Build(build)

	vars := make(map[string]struct{})
	for k := range pt {
		vars[k] = struct{}{}
	}

	nodeState := NewDeclMapLattice()
	lattice := NewCfgMapLattice(nodeState)

	a := Analysis{
		cfg:          start,
		cfgNodes:     build.Nodes(),
		declaredVars: vars,
		nodeState:    nodeState,
		lattice:      lattice,
		pt:           pt,
	}

	return a.analyze()
}

type Analysis struct {
	cfg 			*cfg.StartNode
	cfgNodes		[]cfg.Node
	declaredVars 	map[string]struct{}
	nodeState		DeclMapLattice
	lattice			CfgMapLattice
	pt 				pointto.Result
}

func (a Analysis) store(state DeclMap, lhs string, rhs string) DeclMap {
	for _, v := range a.pt[lhs] {
		if state.Get(rhs) == PN {
			state.Set(v, PN)
		}
	}
	return state
}

func (a Analysis) load(state DeclMap, lhs string, rhs string) DeclMap {
	lub := NN
	for _, v := range a.pt[rhs] {
		if state.Get(v) == PN {
			lub = PN
			break
		}
	}
	state.Set(lhs, lub)
	return state
}

func (a Analysis) transferFun(n cfg.Node, state DeclMap) DeclMap {
	newState := state.Clone()
	switch node := n.(type) {
	case *cfg.AllocNode:
		allocI := fmt.Sprint("alloc-", node.Id())
		newState.Set(node.Lhs, NN)
		newState.Set(allocI, PN)
	case *cfg.RefNode:
		newState.Set(node.Lhs, NN)
	case *cfg.AssignNode:
		newState.Set(node.Lhs, newState.Get(node.Rhs))
	case *cfg.PointerNode:
		return a.load(newState, node.Lhs, node.Rhs)
	case *cfg.DerefNode:
		return a.store(newState, node.Lhs, node.Rhs)
	case *cfg.NullNode:
		newState.Set(node.Lhs, PN)
	}
	return newState
}

func (a Analysis) join(n cfg.Node, x CfgMap) DeclMap {
	states := make([]DeclMap, 0)
	for k := range n.Pred() {
		states = append(states, x.Get(k))
	}

	res := a.nodeState.Bot().Clone()
	for _, v := range states {
		res = a.nodeState.Lub(res, v)
	}
	return res
}

func (a Analysis) funOne(n cfg.Node, x CfgMap) DeclMap {
	return a.transferFun(n, a.join(n, x))
}

func (a Analysis) fun(x CfgMap) CfgMap {
	res := a.lattice.Bot().Clone()
	for _, n := range a.cfgNodes {
		res.Set(n, a.funOne(n, x))
	}
	return res
}

func (a Analysis) analyze() Result {
	x := a.lattice.Bot()
	t := x
	for {
		t = x
		x = a.fun(x)
		if x.Eq(t) {
			break
		}
	}
	return x
}
