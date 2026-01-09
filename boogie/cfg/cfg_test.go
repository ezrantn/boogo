package cfg

import (
	"testing"

	"github.com/ezrantn/boogo/boogie"
)

func id(i int) BlockID { return BlockID(i) }

func intVar(name string) boogie.Var {
	return boogie.Var{Name: name, Ty: boogie.IntType{}}
}

func varExpr(v boogie.Var) *boogie.VarExpr {
	return &boogie.VarExpr{Name: v.Name, V: v}
}

func intLit(n int) *boogie.IntLit {
	return &boogie.IntLit{Value: n}
}

func lt(x, y boogie.Expr) *boogie.BinOp {
	return &boogie.BinOp{
		Op:    boogie.Lt,
		Left:  x,
		Right: y,
		Ty:    boogie.BoolType{},
	}
}

func neg(x boogie.Expr) *boogie.UnOp {
	return &boogie.UnOp{
		Op: boogie.Neg,
		X:  x,
		Ty: boogie.IntType{},
	}
}

func TestBuildCFGLinear(t *testing.T) {
	b0 := &Block{
		ID:    id(0),
		Stmts: nil,
		Term:  &Goto{Targets: []BlockID{id(1)}},
	}

	b1 := &Block{
		ID:    id(1),
		Stmts: nil,
		Term:  &Return{},
	}

	cfg := BuildCFG([]*Block{b0, b1}, id(0))

	if len(cfg.Succ[id(0)]) != 1 || cfg.Succ[id(0)][0] != id(1) {
		t.Fatalf("unexpected successors for block 0: %v", cfg.Succ[id(0)])
	}

	if len(cfg.Pred[id(1)]) != 1 || cfg.Pred[id(1)][0] != id(0) {
		t.Fatalf("unexpected predecessors for block 1: %v", cfg.Pred[id(1)])
	}
}

func TestStructureLinear(t *testing.T) {
	b0 := &Block{
		ID:    id(0),
		Stmts: []boogie.Stmt{},
		Term:  &Goto{Targets: []BlockID{id(1)}},
	}
	b1 := &Block{
		ID:    id(1),
		Stmts: []boogie.Stmt{},
		Term:  &Return{},
	}

	cfg := BuildCFG([]*Block{b0, b1}, id(0))

	stmts, err := Structure(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stmts) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(stmts))
	}

	if _, ok := stmts[0].(*boogie.Return); !ok {
		t.Fatalf("expected Return, got %T", stmts[0])
	}
}

func TestLowerProcTerminatorsTotal(t *testing.T) {
	x := intVar("x")
	y := intVar("y")

	proc := &boogie.Procedure{
		Name:   "abs",
		Params: []boogie.Var{x},
		Rets:   []boogie.Var{y},
		Body: []boogie.Stmt{
			&boogie.If{
				Cond: lt(varExpr(x), intLit(0)),
				Then: []boogie.Stmt{
					&boogie.Assign{
						Lhs: varExpr(y),
						Rhs: neg(varExpr(x)),
					},
				},
				Else: []boogie.Stmt{
					&boogie.Assign{
						Lhs: varExpr(y),
						Rhs: varExpr(x),
					},
				},
			},
			&boogie.Return{},
		},
	}

	blocks, _, err := LowerProcToCFG(proc)
	if err != nil {
		t.Fatalf("lowering failed: %v", err)
	}

	for _, b := range blocks {
		if b.Term == nil {
			t.Fatalf("block %d has nil terminator", b.ID)
		}
	}
}

func TestBuildCFGIfEdges(t *testing.T) {
	// entry -> then / else -> join -> return
	blocks := []*Block{
		{ID: 0, Term: &If{Then: 1, Else: 2}},
		{ID: 1, Term: &Goto{Targets: []BlockID{3}}},
		{ID: 2, Term: &Goto{Targets: []BlockID{3}}},
		{ID: 3, Term: &Return{}},
	}

	cfg := BuildCFG(blocks, 0)

	if len(cfg.Succ[0]) != 2 {
		t.Fatalf("expected 2 successors from entry, got %d", len(cfg.Succ[0]))
	}

	if len(cfg.Pred[3]) != 2 {
		t.Fatalf("expected 2 predecessors to join, got %d", len(cfg.Pred[3]))
	}
}

func TestStructureIf(t *testing.T) {
	x := intVar("x")

	proc := &boogie.Procedure{
		Name: "test",
		Body: []boogie.Stmt{
			&boogie.If{
				Cond: lt(varExpr(x), intLit(0)),
				Then: []boogie.Stmt{
					&boogie.Assign{Lhs: varExpr(x), Rhs: intLit(1)},
				},
				Else: []boogie.Stmt{
					&boogie.Assign{Lhs: varExpr(x), Rhs: intLit(2)},
				},
			},
			&boogie.Return{},
		},
	}

	blocks, entry, err := LowerProcToCFG(proc)
	if err != nil {
		t.Fatalf("lowering failed: %v", err)
	}

	cfg := BuildCFG(blocks, entry)
	stmts, err := Structure(cfg)
	if err != nil {
		t.Fatalf("structuring failed: %v", err)
	}

	// One structured If, return is absorbed into branches
	if len(stmts) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(stmts))
	}

	ifStmt, ok := stmts[0].(*boogie.If)
	if !ok {
		t.Fatalf("expected If stmt, got %T", stmts[0])
	}

	// Then branch ends in return
	if _, ok := ifStmt.Then[len(ifStmt.Then)-1].(*boogie.Return); !ok {
		t.Fatalf("expected return in then branch")
	}

	// Else branch ends in return
	if _, ok := ifStmt.Else[len(ifStmt.Else)-1].(*boogie.Return); !ok {
		t.Fatalf("expected return in else branch")
	}
}

func TestLowerWhileBackEdge(t *testing.T) {
	x := intVar("x")

	proc := &boogie.Procedure{
		Name: "loop",
		Body: []boogie.Stmt{
			&boogie.While{
				Cond: lt(varExpr(x), intLit(10)),
				Body: []boogie.Stmt{
					&boogie.Assign{Lhs: varExpr(x), Rhs: intLit(1)},
				},
			},
			&boogie.Return{},
		},
	}

	blocks, entry, err := LowerProcToCFG(proc)
	if err != nil {
		t.Fatalf("lowering failed: %v", err)
	}

	cfg := BuildCFG(blocks, entry)

	foundBackEdge := false
	for from, succs := range cfg.Succ {
		for _, to := range succs {
			if to < from {
				foundBackEdge = true
			}
		}
	}

	if !foundBackEdge {
		t.Fatalf("expected back-edge in while loop CFG")
	}
}

func TestStructureRejectsCycle(t *testing.T) {
	cfg := &CFG{
		Entry: 0,
		Blocks: map[BlockID]*Block{
			0: {ID: 0, Term: &Goto{Targets: []BlockID{0}}},
		},
	}

	_, err := Structure(cfg)
	if err == nil {
		t.Fatalf("expected structuring failure on cycle")
	}
}
