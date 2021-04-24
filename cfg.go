package protochecks

import (
	"go/ast"

	"golang.org/x/tools/go/cfg"
)

type cfgNode struct {
	ast  ast.Node
	succ map[Node]struct{}
	pred map[Node]struct{}
}

type Variable = string

func (c *cfgNode) AST() ast.Node {
	return c.ast
}

func (c *cfgNode) Succ() map[Node]struct{} {
	return c.succ
}

func (c *cfgNode) Pred() map[Node]struct{} {
	return c.pred
}

func connect(from, to Node) {
	from.Succ()[to] = struct{}{}
	to.Pred()[from] = struct{}{}
}

type StartNode struct {
	cfgNode
}

// X = alloc P
type AllocNode struct {
	cfgNode
	lhs Variable
}

// X_1 = &X_2
type RefNode struct {
	cfgNode
	lhs Variable
	rhs Variable
}

// X_1 = X_2
type AssignNode struct {
	cfgNode
	lhs Variable
	rhs Variable
}

// X_1 = *X_2
type PointerNode struct {
	cfgNode
	lhs Variable
	rhs Variable
}

// *X_1 = X_2
type DerefNode struct {
	cfgNode
	lhs Variable
	rhs Variable
}

// X = null
type NullNode struct {
	cfgNode
	lhs Variable
}

type Node interface {
	AST() ast.Node
	Succ() map[Node]struct{}
	Pred() map[Node]struct{}
}

type Builder struct {
}

func (b *Builder) blockToNode(block *cfg.Block, pred Node) {

}
