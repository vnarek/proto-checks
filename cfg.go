package protochecks

import (
	"go/ast"
	"go/token"
	"strconv"

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
	if from != nil && to != nil {
		from.Succ()[to] = struct{}{}
		to.Pred()[from] = struct{}{}
	}
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

func NewAllocNode(lhs Variable, ast ast.Node) *AllocNode {
	return &AllocNode{
		cfgNode: newCfg(ast),
		lhs:     lhs,
	}
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

func NewDerefNode(lhs, rhs Variable, ast ast.Node) *DerefNode {
	return &DerefNode{
		cfgNode: newCfg(ast),
		lhs:     lhs,
		rhs:	 rhs,
	}
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
	case *AllocNode:
		return "[" + n.lhs + " = alloc]"
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
	FreshVarCnt int
}

func NewBuilder() *Builder {
	return &Builder{
		BlockNode: make(map[*cfg.Block]Node),
		FreshVarCnt: 0,
	}
}

func (b *Builder) SameFreshVar() string {
	return "_t" + strconv.Itoa(b.FreshVarCnt)
}

func (b *Builder) NextFreshVar() string {
	b.FreshVarCnt++
	return b.SameFreshVar()
}

func (b *Builder) BlockToNode(block *cfg.Block, pred Node) {
	var blockFirst Node = nil //represents first node of this block
	currLast := pred
	for _, astNode := range block.Nodes {
		first, last := b.astToNode(astNode)
		if first != nil {
			connect(currLast, first)
			currLast = last
			//if blockFirst is nil, that means this is first node of this block
			if blockFirst == nil {
				blockFirst = first
			}
		}
	}
	//if current block hasn't created any new node, we skip this part
	if blockFirst != nil {
		b.BlockNode[block] = blockFirst
		connect(pred, blockFirst)
	}
	for _, suc := range block.Succs {
		n, ok := b.BlockNode[suc]
		if ok {
			connect(currLast, n)
			continue
		}
		b.BlockToNode(suc, currLast)
	}
}

//appends node to first-last sequence of nodes and returns a new first-last sequence
func (b *Builder) appendNode(n Node, f Node, l Node) (Node, Node) {
	if f == nil {
		return n, n
	}
	connect(l, n)
	return f, n
}

//returns first and last node created by this AST node
func (b *Builder) astToNode(a ast.Node) (first, last Node) {
	switch a := a.(type) {
	case *ast.AssignStmt:
		//extends current first-last sequence with new nodes and returns new first-last sequence
		first, last = b.assignLhsToNode(a.Lhs[0], a.Rhs[0], first, last)
	case *ast.ValueSpec: //for example [var int* x] or [var int* x = new(1)]
		first, last = b.declToNode(a, first, last)
	}
	return first, last
}

//decomposes lhs to string
func (b *Builder) assignLhsToNode(lhsExp ast.Expr, rhsExp ast.Expr, f Node, l Node) (first, last Node) {
	switch lhs := lhsExp.(type) {
	case *ast.Ident:
		first, last = b.assignRhsToNode(lhs.Name, rhsExp, f, l)
	case *ast.StarExpr:
		//check if we need to normalize lhs StarExpr even more
		switch id := lhs.X.(type) {
		case *ast.Ident: //no need to normalize
			//now we'll peek the rhs
			switch rhs := rhsExp.(type) {
			case *ast.Ident:
				if rhs.Name != "nil" {
					return b.appendNode(NewDerefNode(id.Name, rhs.Name, rhsExp), f, l)
				}
				//if it's nil, we will need to normalize it
			}
			freshVar := b.NextFreshVar()
			f, l = b.assignRhsToNode(freshVar, rhsExp, f, l)
			first, last = b.appendNode(NewDerefNode(id.Name, freshVar, lhsExp), f, l)
		default: //we need to normalize lhs StarExpr
			freshVar := b.NextFreshVar()
			//this will normalize lhs and store it in the freshVar
			f, l = b.assignRhsToNode(freshVar, id, f ,l)
			//this will set lhs ast to freshVar (which now represents normalized lhs)
			lhs.X = ast.NewIdent(freshVar)
			first, last = b.assignLhsToNode(lhs, rhsExp, f, l)
		}
	case *ast.UnaryExpr:
		//TODO: normalization of the & operator
		panic("unknown unary expr")
	default:
		panic("unknown assign lhs")
	}
	return first, last
}

func (b *Builder) assignRhsToNode(lhs string, rhsExp ast.Expr, f Node, l Node) (first, last Node) {
	switch rhs := rhsExp.(type) {
	case *ast.Ident:
		if rhs.Name == "nil" {
			first, last = b.appendNode(NewNullNode(lhs, rhsExp), f, l)
		} else {
			first, last = b.appendNode(NewAssignNode(lhs, rhs.Name, rhsExp), f, l)
		}
	case *ast.UnaryExpr:
		if rhs.Op == token.AND { //&
			switch id := rhs.X.(type) {
			case *ast.Ident:
				first, last = b.appendNode(NewRefNode(lhs, id.Name, rhsExp), f, l)
			default: //recursive normalization for * and &
				freshVar := b.NextFreshVar()
				f, l = b.assignRhsToNode(freshVar, rhs.X, f, l)
				first, last = b.appendNode(NewRefNode(lhs, freshVar, rhsExp), f, l)
			}
		}
	case *ast.StarExpr:
		switch id := rhs.X.(type) {
		case *ast.Ident:
			first, last = b.appendNode(NewPointerNode(lhs, id.Name, rhsExp), f, l)
		default: //recursive normalization for * and &
			freshVar := b.NextFreshVar()
			f, l = b.assignRhsToNode(freshVar, rhs.X, f, l)
			first, last = b.appendNode(NewPointerNode(lhs, freshVar, rhsExp), f, l)
		}
	default:
		panic("unknown assign rhs")
	}
	return first, last
}

func (b *Builder) declToNode(decl *ast.ValueSpec, f Node, l Node) (first, last Node) {
	switch exp := decl.Type.(type) {
	case *ast.StarExpr: //we only care about pointer decls
		switch id := exp.X.(type) {
		case *ast.Ident:
			if decl.Values == nil {
				first, last = b.appendNode(NewNullNode(id.Name, decl), f, l)
			} else {
				first, last = b.appendNode(NewAllocNode(id.Name, decl), f, l)
			}
		case *ast.StarExpr:
			//we'll probably have to rewrite this function because of this case
			//first, we will count how many levels of indirection there is and then normalize it appropriately
			panic("unhandled double pointer declaration")
		}
	}
	return first, last
}
