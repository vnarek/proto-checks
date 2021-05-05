package pointto

import (
	"fmt"
	"go/ast"

	"github.com/vnarek/proto-checks/cfg"
)

var fil = struct{}{}

type Edges map[string]struct{}

type VarPair struct {
	from, to string
}

type Constrain struct {
	Valid bool
	Thens map[VarPair]struct{}
}

type Node struct {
	ast    ast.Node
	edges  Edges
	constr map[string]Constrain
}

//stands for directed acyclic graph
type DAG struct {
	nodes map[string]*Node

	cellsSet map[string]struct{}
	cells    []string
}

func NewDAG() *DAG {
	return &DAG{
		cellsSet: make(map[string]struct{}),
		nodes:    make(map[string]*Node),
	}
}

func (d *DAG) AddNode(name string, ast ast.Node) {
	_, ok := d.nodes[name]
	if ok {
		return
	}
	if d.cells != nil {
		panic("can't call AddNode after first GetCells")
	}
	d.cellsSet[name] = fil
	d.nodes[name] = &Node{
		ast:    ast,
		edges:  make(Edges),
		constr: make(map[string]Constrain),
	}
}

func (d *DAG) GetCells() []string {
	if d.cells == nil {
		for n := range d.cellsSet {
			d.cells = append(d.cells, n)
		}
	}
	return d.cells
}

// AddEdge adds new edge `from`->`to`
// and tries to normalize graph by merging
// cycles
func (d *DAG) AddEdge(from, to string) {
	d.nodes[from].edges[to] = fil
	d.MergeCycles(from, to)
}

// AddImpliedContrain adds contrain in the form
// t ∈ [[x]] -> [[y]] ⊆ [[z]]
func (d *DAG) AddImpliedConstrain(t, x, y, z string) {
	constr := d.nodes[x].constr[t]
	if constr.Valid {
		// If the contrain is valid already
		// just add the needed edge
		d.AddSubsetConstrain(y, z)
		return
	}

	// otherwise we add the contrain to `Thens`
	// so we can add it later if this contrain
	// becomes true
	if d.nodes[x].constr[t].Thens == nil {
		thens := make(map[VarPair]struct{})
		d.nodes[x].constr[t] = Constrain{
			Thens: thens,
		}
	}
	d.nodes[x].constr[t].Thens[VarPair{y, z}] = fil
}

// x2 is subset of x1 constrain
func (d *DAG) AddSubsetConstrain(x2, x1 string) {
	d.AddEdge(x2, x1)

	for name, c := range d.nodes[x2].constr {
		if c.Valid {
			for edge := range d.nodes[x2].edges {
				d.AddInConstrain(name, edge)
			}
		}
	}
}

func (d *DAG) MergeCycles(from, to string) {
	cycleCells := d.mergeCycle(to, from, []string{to})
	if len(cycleCells) == 0 {
		return
	}

	mergedNode := d.nodes[from]

	for _, n := range cycleCells {
		node := d.nodes[n]
		for name, c := range node.constr {
			if c.Valid {
				constr := mergedNode.constr[name]
				constr.Valid = true
				mergedNode.constr[name] = constr
			}
			for pair := range c.Thens {
				mergedNode.constr[name].Thens[pair] = fil
			}
		}
	}

	for _, c := range cycleCells {
		d.nodes[c] = mergedNode
	}

	for n := range mergedNode.edges {
		if d.nodes[n] == mergedNode {
			delete(mergedNode.edges, n)
		}
	}

	for name, c := range mergedNode.constr {
		if c.Valid {
			for edge := range mergedNode.edges {
				d.AddInConstrain(name, edge)
			}
		}
	}
}

func (d *DAG) mergeCycle(actual, start string, cycle []string) []string {
	if actual == start {
		return cycle
	}

	for n := range d.nodes[actual].edges {
		c := d.mergeCycle(n, start, append(cycle, n))
		if len(c) != 0 {
			return c
		}
	}

	return nil
}

func (d *DAG) AddInConstrain(t, x string) {
	constr := d.nodes[x].constr[t]
	constr.Valid = true

	for then := range constr.Thens {
		d.AddSubsetConstrain(then.from, then.to)
	}

	d.nodes[x].constr[t] = Constrain{
		Valid: true,
	}

	for x := range d.nodes[x].edges {
		d.AddInConstrain(t, x)
	}
}

type Result map[string][]string

func Build(build *cfg.Builder) Result {
	nodes := build.Nodes()
	sol := NewDAG()

	fillDAG(sol, nodes)

	solve(sol, nodes)

	return createResult(sol)
}

// Not sure about this one
func createResult(dag *DAG) Result {
	res := make(Result)

	for _, c := range dag.cells {
		for k, constr := range dag.nodes[c].constr {
			if constr.Valid {
				res[c] = append(res[c], k)
			}
		}
	}
	return res
}

func solve(dag *DAG, nodes []cfg.Node) {
	for _, n := range nodes {
		switch node := n.(type) {
		case *cfg.AllocNode:
			allocI := fmt.Sprint("alloc-", node.Id())
			dag.AddInConstrain(allocI, node.Lhs)
		case *cfg.RefNode:
			dag.AddInConstrain(node.Rhs, node.Lhs)
		case *cfg.AssignNode:
			dag.AddSubsetConstrain(node.Rhs, node.Lhs)
		case *cfg.PointerNode:
			for _, c := range dag.GetCells() {
				dag.AddImpliedConstrain(c, node.Rhs, c, node.Lhs)
			}
		case *cfg.DerefNode:
			for _, c := range dag.GetCells() {
				dag.AddImpliedConstrain(c, node.Lhs, node.Rhs, c)
			}
		}
	}
}

// fillDAG fills DAG with variables from cfg.Node
func fillDAG(dag *DAG, nodes []cfg.Node) {
	for _, n := range nodes {
		ast := n.AST()
		switch node := n.(type) {
		case *cfg.AllocNode:
			allocI := fmt.Sprint("alloc-", node.Id())
			dag.AddNode(allocI, ast)
			dag.AddNode(node.Lhs, ast)
		case *cfg.RefNode:
			dag.AddNode(node.Lhs, ast)
			dag.AddNode(node.Rhs, ast)
		case *cfg.AssignNode:
			dag.AddNode(node.Lhs, ast)
			dag.AddNode(node.Rhs, ast)
		case *cfg.PointerNode:
			dag.AddNode(node.Lhs, ast)
			dag.AddNode(node.Rhs, ast)
		case *cfg.DerefNode:
			dag.AddNode(node.Lhs, ast)
			dag.AddNode(node.Rhs, ast)
		}
	}
	dag.GetCells()
}
