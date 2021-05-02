package cfg

import (
	"go/ast"
	"go/token"
	"strconv"
)

var cnt = 0

type cfgNode struct {
	ast  ast.Node
	id   int
}

func newCfg(ast ast.Node) cfgNode {
	cnt++
	return cfgNode{
		ast:  ast,
		id:   cnt - 1,
	}
}

type Variable = string

func (c *cfgNode) AST() ast.Node {
	return c.ast
}

func (c *cfgNode) Id() int {
	return c.id
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
		rhs:     rhs,
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
	Id() int
}

type Builder struct {
	FreshVarCnt int
}

func NewBuilder() *Builder {
	return &Builder{
		FreshVarCnt: 0,
	}
}

func (b *Builder) sameFreshVar() string {
	return "_t" + strconv.Itoa(b.FreshVarCnt)
}

func (b *Builder) nextFreshVar() string {
	b.FreshVarCnt++
	return b.sameFreshVar()
}

func (b *Builder) GetNodes(a ast.Node) []Node {
	var nodes []Node
	return b.astToNode(a, nodes)
}


func (b *Builder) astToNode(a ast.Node, nodes []Node) []Node {
	switch a := a.(type) {
	case *ast.DeclStmt:
		nodes = b.astToNode(a.Decl, nodes)
	case *ast.GenDecl:
		for _, subStmt := range a.Specs {
			nodes = b.astToNode(subStmt, nodes)
		}
	case *ast.LabeledStmt:
		nodes = b.astToNode(a.Stmt, nodes)
	case *ast.BlockStmt:
		for _, subStmt := range a.List {
			nodes = b.astToNode(subStmt, nodes)
		}
	case *ast.IfStmt:
		if a.Init != nil {
			nodes = b.astToNode(a.Init, nodes)
		}
		nodes = b.astToNode(a.Body, nodes)
		if a.Else != nil {
			nodes = b.astToNode(a.Else, nodes)
		}
	case *ast.CaseClause:
		for _, subStmt := range a.Body {
			nodes = b.astToNode(subStmt, nodes)
		}
	case *ast.SwitchStmt:
		if a.Init != nil {
			nodes = b.astToNode(a.Init, nodes)
		}
		nodes = b.astToNode(a.Body, nodes)
	case *ast.TypeSwitchStmt:
		if a.Init != nil {
			nodes = b.astToNode(a.Init, nodes)
		}
		nodes = b.astToNode(a.Assign, nodes)
		nodes = b.astToNode(a.Body, nodes)
	case *ast.CommClause:
		if a.Comm != nil {
			nodes = b.astToNode(a.Comm, nodes)
		}
		for _, subStmt := range a.Body {
			nodes = b.astToNode(subStmt, nodes)
		}
	case *ast.SelectStmt:
		nodes = b.astToNode(a.Body, nodes)
	case *ast.ForStmt:
		if a.Init != nil {
			nodes = b.astToNode(a.Init, nodes)
		}
		if a.Post != nil {
			nodes = b.astToNode(a.Post, nodes)
		}
		nodes = b.astToNode(a.Body, nodes)
	case *ast.RangeStmt:
		nodes = b.astToNode(a.Body, nodes)
	case *ast.AssignStmt:
		nodes = b.assignLhsToNode(a.Lhs[0], a.Rhs[0], nodes)
	case *ast.ValueSpec: //for example [var int* x] or [var int* x = new(1)]
		nodes = b.declToNode(a, nodes)
	}
	return nodes
}

//decomposes lhs to string
func (b *Builder) assignLhsToNode(lhsExp ast.Expr, rhsExp ast.Expr, nodes []Node) []Node {
	switch lhs := lhsExp.(type) {
	case *ast.ParenExpr:
		nodes = b.assignLhsToNode(lhs.X, rhsExp, nodes)
	case *ast.Ident:
		nodes = b.assignRhsToNode(lhs.Name, rhsExp, nodes)
	case *ast.StarExpr:
		//check if we need to normalize lhs StarExpr even more
		switch id := lhs.X.(type) {
		case *ast.Ident: //no need to normalize
			//now we'll peek the rhs
			switch rhs := rhsExp.(type) {
			case *ast.Ident:
				nodes = append(nodes, NewDerefNode(id.Name, rhs.Name, rhsExp))
				return nodes
			}
			freshVar := b.nextFreshVar()
			prevLen := len(nodes)
			nodes = b.assignRhsToNode(freshVar, rhsExp, nodes)
			//only if new nodes were created we will append DerefNode
			if prevLen != len(nodes) {
				nodes = append(nodes, NewDerefNode(id.Name, freshVar, lhsExp))
			}
		default: //we need to normalize lhs StarExpr
			freshVar := b.nextFreshVar()
			//this will normalize lhs and store it in the freshVar
			nodes = b.assignRhsToNode(freshVar, id, nodes)
			//this will set lhs ast to freshVar (which now represents normalized lhs)
			lhs.X = ast.NewIdent(freshVar)
			nodes = b.assignLhsToNode(lhs, rhsExp, nodes)
		}
	}
	return nodes
}

func (b *Builder) assignRhsToNode(lhs string, rhsExp ast.Expr, nodes []Node) []Node {
	switch rhs := rhsExp.(type) {
	case *ast.ParenExpr:
		nodes = b.assignRhsToNode(lhs, rhs.X, nodes)
	case *ast.Ident:
		if rhs.Name == "nil" {
			nodes = append(nodes, NewNullNode(lhs, rhsExp))
		} else {
			nodes = append(nodes, NewAssignNode(lhs, rhs.Name, rhsExp))
		}
	case *ast.UnaryExpr:
		if rhs.Op == token.AND { //&
			switch id := rhs.X.(type) {
			case *ast.Ident:
				nodes = append(nodes, NewRefNode(lhs, id.Name, rhsExp))
			default: //recursive normalization for * and &
				freshVar := b.nextFreshVar()
				nodes = b.assignRhsToNode(freshVar, rhs.X, nodes)
				nodes = append(nodes, NewRefNode(lhs, freshVar, rhsExp))
			}
		}
	case *ast.StarExpr:
		switch id := rhs.X.(type) {
		case *ast.Ident:
			nodes = append(nodes, NewPointerNode(lhs, id.Name, rhsExp))
		default: //recursive normalization for * and &
			freshVar := b.nextFreshVar()
			nodes = b.assignRhsToNode(freshVar, rhs.X, nodes)
			nodes = append(nodes, NewPointerNode(lhs, freshVar, rhsExp))
		}
	case *ast.FuncLit:
		//TODO: see CallExpr, basically the same problem
		nodes = append(nodes, NewAllocNode(lhs, rhsExp))
	case *ast.CallExpr:
		//TODO: dunno what to do here..
		//		assume there's no normalization needed and create AllocNode (what if the function returns double pointer)?
		//		I guess for now, yes..
		nodes = append(nodes, NewAllocNode(lhs, rhsExp))
	}
	return nodes
}

func (b *Builder) declToNode(spec *ast.ValueSpec, nodes []Node) []Node {
	//first, we need to count the number of * references of spec's type
	refCount := 1 //if type is undefined, we assume that there's one * reference
	stop := false
	if spec.Type != nil {
		specType := spec.Type
		for i := 0; !stop; i++ {
			switch expr := specType.(type) {
			case *ast.StarExpr:
				specType = expr.X
			default:
				refCount = i
				stop = true
			}
		}
	}

	//only create nodes if there is at least one level of * reference
	if refCount > 0 {
		//normalize more than 1 * reference
		currVar := spec.Names[0].Name
		var tempNodes []Node
		for i := 1; i < refCount; i++ {
			freshVar := b.nextFreshVar()
			//prepending each new node
			tempNodes = append([]Node{NewRefNode(currVar, freshVar, spec)}, tempNodes...)
			currVar = freshVar
		}
		if spec.Values == nil {
			nodes = append(nodes, NewNullNode(currVar, spec))
		} else {
			nodes = b.assignRhsToNode(currVar, spec.Values[0], nodes)
		}
		//now we will append normalization nodes
		nodes = append(nodes, tempNodes...)
	}
	return nodes
}
