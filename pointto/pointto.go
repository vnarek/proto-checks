package pointto

import (
	"fmt"

	"github.com/vnarek/proto-checks/cfg"
)

type Result = map[string]Constrains

var fil = struct{}{}

type Cells map[string]cfg.Node

func (c Cells) Merge(other Cells) {
	for o, v := range other {
		c[o] = v
	}
}

type Points map[string]struct{}

func (p Points) Merge(other Points) {
	for o := range other {
		p[o] = fil
	}
}

type Constrains struct {
	Cells  Cells
	Points Points
}

func (c Constrains) Merge(other Constrains) {
	c.Cells.Merge(other.Cells)
	c.Points.Merge(other.Points)
}

func Build(build *cfg.Builder) Result {
	nodes := build.Nodes()

	res := contrains(nodes)
}

func contrains(nodes []cfg.Node) Result {
	var res Result

	for _, n := range nodes {
		switch x := n.(type) {
		case *cfg.AllocNode:
			// maybe better representation?
			allocI := fmt.Sprintf("alloc-%s", x.Id())
			res[x.Lhs].Cells[allocI] = x
		case *cfg.RefNode:
			res[x.Lhs].Points[x.Rhs] = fil
		case *cfg.AssignNode:
			res[x.Lhs].Merge(res[x.Rhs])
		case *cfg.PointerNode:
		case *cfg.DerefNode:
		case *cfg.NullNode:
			//noop
		}
	}

	return res
}
