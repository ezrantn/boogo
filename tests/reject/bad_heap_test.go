package reject

import (
	"testing"

	"github.com/ezrantn/boogo/boogie"
	"github.com/ezrantn/boogo/ebs"
)

func TestBadHeap(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			t.Fatalf("expected checker to reject program")
		}
	}()

	prog := &boogie.Program{
		Procs: []*boogie.Procedure{
			{
				Name: "badHeap",
				Locals: []boogie.Var{
					{Name: "x", Ty: boogie.IntType{}},
				},
				Body: []boogie.Stmt{
					&boogie.HeapRead{
						Obj:   &boogie.VarExpr{V: boogie.Var{Name: "x"}},
						Field: "f",
					},
				},
			},
		},
	}

	ebs.Check(prog)
}
