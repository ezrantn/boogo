package ok

import (
	"testing"

	"github.com/ezrantn/boogo/boogie"
	"github.com/ezrantn/boogo/cmd/boogo"
	"github.com/ezrantn/boogo/ebs"
)

func TestArith(t *testing.T) {
	prog := &boogie.Program{
		Procs: []*boogie.Procedure{
			{
				Name: "inc",
				Params: []boogie.Var{
					{Name: "x", Ty: boogie.IntType{}},
				},
				Rets: []boogie.Var{
					{Name: "y", Ty: boogie.IntType{}},
				},
				Body: []boogie.Stmt{
					&boogie.Assign{
						Lhs: &boogie.VarExpr{V: boogie.Var{Name: "y"}},
						Rhs: &boogie.BinOp{
							Op:    boogie.Add,
							Left:  &boogie.VarExpr{V: boogie.Var{Name: "x"}},
							Right: &boogie.IntLit{Value: 1},
						},
					},
					&boogie.Return{
						Values: []boogie.Expr{
							&boogie.VarExpr{V: boogie.Var{Name: "y"}},
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
