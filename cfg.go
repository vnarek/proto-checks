package protochecks

import (
	"go/ast"
	"go/token"

	"golang.org/x/tools/go/cfg"
)

type cfgNode struct {
	ast  ast.Node
	succ map[Node]struct{}
	pred map[Node]struct{}
}

func newCfg(ast ast.Node) cfgNode {
	return cfgNode{
		ast:  ast,
		succ: make(map[Node]struct{}),
		pred: make(map[Node]struct{}),
	}
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

func NewStartNode() *StartNode {
	return &StartNode{
		cfgNode: newCfg(nil),
	}
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

func NewRefNode(lhs, rhs Variable, ast ast.Node) *RefNode {
	return &RefNode{
		cfgNode: newCfg(ast),
		lhs:     lhs,
		rhs:     rhs,
	}
}

// X_1 = X_2
type AssignNode struct {
	cfgNode
	lhs Variable
	rhs Variable
}

func NewAssignNode(lhs, rhs Variable, ast ast.Node) *AssignNode {
	return &AssignNode{
		cfgNode: newCfg(ast),
		lhs:     lhs,
		rhs:     rhs,
	}
}

// X_1 = *X_2
type PointerNode struct {
	cfgNode
	lhs Variable
	rhs Variable
}

func NewPointerNode(lhs, rhs Variable, ast ast.Node) *PointerNode {
	return &PointerNode{
		cfgNode: newCfg(ast),
		lhs:     lhs,
		rhs:     rhs,
	}
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

func NewNullNode(lhs Variable, ast ast.Node) *NullNode {
	return &NullNode{
		cfgNode: newCfg(ast),
		lhs:     lhs,
	}
}

func ToString(node Node) string {
	switch n := node.(type) {
	case *StartNode:
		return "[START]"
	case *RefNode:
		return "[" + n.lhs + " = &" + n.rhs + "]"
	case *AssignNode:
		return "[" + n.lhs + " = " + n.rhs + "]"
	case *PointerNode:
		return "[" + n.lhs + " = *" + n.rhs + "]"
	case *DerefNode:
		return "[*" + n.lhs + " = " + n.rhs + "]"
	case *NullNode:
		return "[" + n.lhs + " = null]"
	default:
		panic("unimplemented print")
	}
}

type Node interface {
	AST() ast.Node
	Succ() map[Node]struct{}
	Pred() map[Node]struct{}
}

type Builder struct {
	BlockNode map[*cfg.Block]Node
}

func NewBuilder() *Builder {
	return &Builder{
		BlockNode: make(map[*cfg.Block]Node),
	}
}

/*
def blockToNode(block: Block, pred: CfgNode): Unit = {
  //currPred = pred
  //for ast in Nodes
    //currPred = astToNode(ast, currPred)

  //for succ in Succs
    //je succ ve struktuře blocks?
      //ANO: napoj currPred na blocks[succ]
      //NE: res = blockToNode(succ, currPred); blocks[succ] = res;
}
*/

func (b *Builder) BlockToNode(block *cfg.Block, pred Node) {
	first := pred
	currPred := pred
	for _, astNode := range block.Nodes {
		f, l := b.astToNode(astNode, currPred)
		if f != nil {
			if first == pred {
				first = f
			}
			currPred = l
		}
	}
	//don't store blocks, that had no nodes
	if first != pred {
		b.BlockNode[block] = first
	}
	for _, suc := range block.Succs {
		n, ok := b.BlockNode[suc]
		if ok {
			connect(currPred, n)
			continue
		}
		b.BlockToNode(suc, currPred)
	}
}

func (b *Builder) astToNode(a ast.Node, pred Node) (first, last Node) {
	switch a := a.(type) {
	case *ast.AssignStmt:
		first, last = b.assignStmtToNode(a)
	case *ast.ReturnStmt:
	}
	if first != nil {
		connect(pred, first)
	}
	return first, last
}

func (b *Builder) assignStmtToNode(stmt *ast.AssignStmt) (first, last Node) {
	switch lhs := stmt.Lhs[0].(type) { // TODO: multivariable
	case *ast.Ident:
		switch rhs := stmt.Rhs[0].(type) {
		case *ast.Ident:
			if rhs.Name == "nil" {
				n := NewNullNode(lhs.Name, stmt)
				return n, n
			} else {
				n := NewAssignNode(lhs.Name, rhs.Name, stmt)
				return n, n
			}
		case *ast.UnaryExpr:
			if rhs.Op == token.AND { //&
				id, ok := rhs.X.(*ast.Ident)
				if !ok {
					panic("expected ident")
				}
				n := NewRefNode(lhs.Name, id.Name, stmt)
				return n, n
			}
		case *ast.StarExpr:
			switch id := rhs.X.(type) {
			case *ast.Ident:
				n := NewPointerNode(lhs.Name, id.Name, stmt)
				return n, n
			}
		}

	default:
		panic("wtf?")
	}
	return nil, nil
}
