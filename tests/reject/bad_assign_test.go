package reject

import (
	"testing"

	"github.com/ezrantn/boogo/boogie"
	"github.com/ezrantn/boogo/ebs"
)

func TestBadAssign(t *testing.T) {
	prog := &boogie.Program{
		Procs: []*boogie.Procedure{
			{
				Name: "bad",
				Locals: []boogie.Var{
					{Name: "x", Ty: boogie.IntType{}},
				},
				Body: []boogie.Stmt{
					&boogie.Assign{
						Lhs: &boogie.VarExpr{
							V: boogie.Var{Name: "x", Ty: boogie.IntType{}},
						},
						Rhs: &boogie.BoolLit{Value: true}, // ❌
					},
				},
			},
		},
	}

	if err := ebs.Check(prog); err == nil {
		t.Fatalf("expected checker to reject program")
	}
}
