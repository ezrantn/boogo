package ok

import (
	"testing"

	"github.com/ezrantn/boogo/boogie"
	"github.com/ezrantn/boogo/cmd/boogo"
	"github.com/ezrantn/boogo/ebs"
)

func TestHeap(t *testing.T) {
	prog := &boogie.Program{
		Procs: []*boogie.Procedure{
			{
				Name: "heapTest",
				Params: []boogie.Var{
					{Name: "o", Ty: boogie.RefType{}},
				},
				Body: []boogie.Stmt{
					&boogie.HeapWrite{
						Obj:   &boogie.VarExpr{V: boogie.Var{Name: "o"}},
						Field: "f",
						Value: &boogie.IntLit{Value: 42},
					},
					&boogie.Assign{
						Lhs: &boogie.VarExpr{V: boogie.Var{Name: "x"}},
						Rhs: &boogie.HeapRead{
							Obj:   &boogie.VarExpr{V: boogie.Var{Name: "o"}},
							Field: "f",
						},
					},
					&boogie.Return{},
				},
			},
		},
	}

	ebs.Check(prog)
	ep := ebs.Erase(prog)
	_ = boogo.EmitProgram(ep)
}
