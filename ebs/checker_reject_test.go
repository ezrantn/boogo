package ebs

import (
	"testing"

	"github.com/ezrantn/boogo/boogie"
)

// ========================
// Helpers
// ========================

func mustReject(t *testing.T, p *boogie.Program) {
	t.Helper()
	if err := Check(p); err == nil {
		t.Fatalf("expected program to be rejected, but it was accepted")
	}
}

// ========================
// Reject Tests
// ========================

// ❌ Recursive procedure
func TestRejectRecursiveCall(t *testing.T) {
	p := &boogie.Program{
		Procs: []*boogie.Procedure{
			{
				Name: "f",
				Body: []boogie.Stmt{
					&boogie.Call{
						Name: "f",
						Args: nil,
						Rets: nil,
					},
				},
			},
		},
	}

	mustReject(t, p)
}

// ❌ Call to unknown procedure
func TestRejectUnknownCall(t *testing.T) {
	p := &boogie.Program{
		Procs: []*boogie.Procedure{
			{
				Name: "main",
				Body: []boogie.Stmt{
					&boogie.Call{
						Name: "unknown",
						Args: nil,
						Rets: nil,
					},
				},
			},
		},
	}

	mustReject(t, p)
}

// ❌ Type mismatch in assignment
func TestRejectTypeMismatchAssign(t *testing.T) {
	x := boogie.Var{Name: "x", Ty: boogie.IntType{}}

	p := &boogie.Program{
		Procs: []*boogie.Procedure{
			{
				Name:   "main",
				Locals: []boogie.Var{x},
				Body: []boogie.Stmt{
					&boogie.Assign{
						Lhs: &boogie.VarExpr{V: x},
						Rhs: &boogie.BoolLit{Value: true},
					},
				},
			},
		},
	}

	mustReject(t, p)
}

// ❌ Non-boolean condition in if
func TestRejectIfConditionNotBool(t *testing.T) {
	p := &boogie.Program{
		Procs: []*boogie.Procedure{
			{
				Name: "main",
				Body: []boogie.Stmt{
					&boogie.If{
						Cond: &boogie.IntLit{Value: 42},
						Then: nil,
						Else: nil,
					},
				},
			},
		},
	}

	mustReject(t, p)
}

// ❌ Heap read on non-ref
func TestRejectHeapReadNonRef(t *testing.T) {
	x := boogie.Var{Name: "x", Ty: boogie.IntType{}}

	p := &boogie.Program{
		Procs: []*boogie.Procedure{
			{
				Name:   "main",
				Locals: []boogie.Var{x},
				Body: []boogie.Stmt{
					&boogie.Assign{
						Lhs: &boogie.VarExpr{V: x},
						Rhs: &boogie.HeapRead{
							Obj:   &boogie.VarExpr{V: x}, // ❌ not RefType
							Field: "f",
							Ty:    boogie.IntType{},
						},
					},
				},
			},
		},
	}

	mustReject(t, p)
}

// ❌ Binary operator with mismatched operand types
func TestRejectBinOpTypeMismatch(t *testing.T) {
	p := &boogie.Program{
		Procs: []*boogie.Procedure{
			{
				Name: "main",
				Body: []boogie.Stmt{
					&boogie.Assign{
						Lhs: &boogie.VarExpr{
							V: boogie.Var{Name: "x", Ty: boogie.IntType{}},
						},
						Rhs: &boogie.BinOp{
							Op:    boogie.Add,
							Left:  &boogie.IntLit{Value: 1},
							Right: &boogie.BoolLit{Value: true}, // ❌
							Ty:    boogie.IntType{},
						},
					},
				},
			},
		},
	}

	mustReject(t, p)
}

// ❌ Return with ill-typed expression
func TestRejectBadReturn(t *testing.T) {
	p := &boogie.Program{
		Procs: []*boogie.Procedure{
			{
				Name: "main",
				Rets: []boogie.Var{
					{Name: "r", Ty: boogie.IntType{}},
				},
				Body: []boogie.Stmt{
					&boogie.Return{
						Values: []boogie.Expr{
							&boogie.BoolLit{Value: true}, // ❌
						},
					},
				},
			},
		},
	}

	mustReject(t, p)
}

func TestEraseRemovesAssert(t *testing.T) {
	p := &boogie.Program{
		Procs: []*boogie.Procedure{
			{
				Name: "main",
				Body: []boogie.Stmt{
					&boogie.Assert{Cond: &boogie.BoolLit{Value: true}},
					&boogie.Return{Values: nil},
				},
			},
		},
	}

	e := Erase(p)
	if len(e.Procs[0].Body) != 1 {
		t.Fatalf("assert not erased")
	}
}
