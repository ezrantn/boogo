package ok

import (
	"testing"

	"github.com/ezrantn/boogo/boogie"
	"github.com/ezrantn/boogo/cmd/boogo"
	"github.com/ezrantn/boogo/ebs"
)

func TestControl(t *testing.T) {
	prog := &boogie.Program{
		Procs: []*boogie.Procedure{
			{
				Name: "count",
				Params: []boogie.Var{
					{Name: "n", Ty: boogie.IntType{}},
				},
				Rets: []boogie.Var{
					{Name: "r", Ty: boogie.IntType{}},
				},
				Locals: []boogie.Var{
					{Name: "i", Ty: boogie.IntType{}},
				},
				Body: []boogie.Stmt{
					&boogie.Assign{
						Lhs: &boogie.VarExpr{V: boogie.Var{Name: "i"}},
						Rhs: &boogie.IntLit{Value: 0},
					},
					&boogie.While{
						Cond: &boogie.BinOp{
							Op:    boogie.Lt,
							Left:  &boogie.VarExpr{V: boogie.Var{Name: "i"}},
							Right: &boogie.VarExpr{V: boogie.Var{Name: "n"}},
						},
						Body: []boogie.Stmt{
							&boogie.Assign{
								Lhs: &boogie.VarExpr{V: boogie.Var{Name: "i"}},
								Rhs: &boogie.BinOp{
									Op:    boogie.Add,
									Left:  &boogie.VarExpr{V: boogie.Var{Name: "i"}},
									Right: &boogie.IntLit{Value: 1},
								},
							},
						},
					},
					&boogie.Assign{
						Lhs: &boogie.VarExpr{V: boogie.Var{Name: "r"}},
						Rhs: &boogie.VarExpr{V: boogie.Var{Name: "i"}},
					},
					&boogie.Return{
						Values: []boogie.Expr{
							&boogie.VarExpr{V: boogie.Var{Name: "r"}},
						},
					},
				},
			},
		},
	}

	ebs.Check(prog)
	ep := ebs.Erase(prog)
	_ = boogo.EmitProgram(ep)
}
