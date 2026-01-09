package frontend

import (
	"testing"

	"github.com/ezrantn/boogo/boogie"
)

type parseTestCase struct {
	name        string
	src         string
	check       func(t *testing.T, prog *boogie.Program)
	expectError bool
}

func TestParseExecutableProcedure(t *testing.T) {
	tests := []parseTestCase{
		{
			name: "simple executable procedure",
			src: `
procedure p(x: int) returns (y: int) {
  var z: int;
  z := x + 1;
  if (z > 0) {
    y := z;
  } else {
    y := 0;
  }
  return y;
}
`,
			check: checkSimpleProcedure,
		},
		{
			name:  "operator precedence",
			src:   `procedure p(x: int) { var z: int; z := x + 2 * 3; }`,
			check: checkPrecedence,
		},
		{
			name:  "unary negation",
			src:   `procedure p(x: int) { var z: int; z := -x; }`,
			check: checkUnaryNeg,
		},
		{
			name:        "reject goto",
			src:         `procedure p() { goto L; }`,
			expectError: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			prog, err := Parse([]byte(tt.src))

			if tt.expectError {
				if err == nil {
					t.Fatalf("expected parse error, got nil")
				}

				return
			}

			if err != nil {
				t.Fatalf("Parse failed: %v", err)
			}
			if prog == nil {
				t.Fatal("program is nil")
			}

			if tt.check != nil {
				tt.check(t, prog)
			}
		})
	}

}

func checkSimpleProcedure(t *testing.T, prog *boogie.Program) {
	t.Helper()

	if len(prog.Procs) != 1 {
		t.Fatalf("expected 1 procedure, got %d", len(prog.Procs))
	}

	p := prog.Procs[0]

	// ----------------------
	// Header
	// ----------------------
	if p.Name != "p" {
		t.Fatalf("expected procedure name 'p', got %q", p.Name)
	}

	assertVarList(t, p.Params, []string{"x"}, boogie.IntType{})
	assertVarList(t, p.Rets, []string{"y"}, boogie.IntType{})

	// ----------------------
	// Body
	// ----------------------
	if len(p.Body) != 4 {
		t.Fatalf("expected 4 statements, got %d", len(p.Body))
	}

	assertLocalDecl(t, p.Body[0], "z", boogie.IntType{})
	assertAssignAdd(t, p.Body[1], "z")
	assertIfGtZero(t, p.Body[2])
	assertReturnVar(t, p.Body[3], "y")
}

func checkPrecedence(t *testing.T, prog *boogie.Program) {
	t.Helper()

	p := prog.Procs[0]
	assign := p.Body[1].(*boogie.Assign)

	add, ok := assign.Rhs.(*boogie.BinOp)
	if !ok || add.Op != boogie.Add {
		t.Fatalf("expected top-level Add binop")
	}

	// Left: x
	if _, ok := add.Left.(*boogie.VarExpr); !ok {
		t.Fatalf("expected left operand VarExpr")
	}

	// Right: 2 * 3
	mul, ok := add.Right.(*boogie.BinOp)
	if !ok || mul.Op != boogie.Mul {
		t.Fatalf("expected right operand Mul binop")
	}

	if _, ok := mul.Left.(*boogie.IntLit); !ok {
		t.Errorf("expected mul left IntLit")
	}

	if _, ok := mul.Right.(*boogie.IntLit); !ok {
		t.Errorf("expected mul right IntLit")
	}
}

func checkUnaryNeg(t *testing.T, prog *boogie.Program) {
	t.Helper()

	p := prog.Procs[0]
	assign := p.Body[1].(*boogie.Assign)

	un, ok := assign.Rhs.(*boogie.UnOp)
	if !ok {
		t.Fatalf("expected RHS UnOp, got %T", assign.Rhs)
	}

	if un.Op != boogie.Neg {
		t.Errorf("expected Neg operator, got %v", un.Op)
	}

	vx, ok := un.X.(*boogie.VarExpr)
	if !ok || vx.V.Name != "x" {
		t.Errorf("expected unary operand VarExpr x")
	}

}

func assertVarList(t *testing.T, vars []boogie.Var, names []string, ty boogie.Type) {
	t.Helper()

	if len(vars) != len(names) {
		t.Fatalf("expected %d vars, got %d", len(names), len(vars))
	}

	for i, v := range vars {
		if v.Name != names[i] {
			t.Errorf("expected var %q, got %q", names[i], v.Name)
		}
		if v.Ty == nil {
			t.Errorf("var %q has nil type", v.Name)
		}
	}
}

func assertLocalDecl(t *testing.T, stmt boogie.Stmt, name string, _ boogie.Type) {
	t.Helper()

	decl, ok := stmt.(*boogie.LocalDecl)
	if !ok {
		t.Fatalf("expected LocalDecl, got %T", stmt)
	}

	if decl.V.Name != name {
		t.Errorf("expected local var %q, got %q", name, decl.V.Name)
	}
}

func assertAssignAdd(t *testing.T, stmt boogie.Stmt, lhsName string) {
	t.Helper()

	assign, ok := stmt.(*boogie.Assign)
	if !ok {
		t.Fatalf("expected Assign, got %T", stmt)
	}

	lhs, ok := assign.Lhs.(*boogie.VarExpr)
	if !ok || lhs.V.Name != lhsName {
		t.Fatalf("expected LHS VarExpr %q, got %T (%v)", lhsName, assign.Lhs, lhs)
	}

	bin, ok := assign.Rhs.(*boogie.BinOp)
	if !ok {
		t.Fatalf("expected RHS BinOp, got %T", assign.Rhs)
	}

	if bin.Op != boogie.Add {
		t.Errorf("expected Add op, got %v", bin.Op)
	}

	if _, ok := bin.Left.(*boogie.VarExpr); !ok {
		t.Errorf("expected left operand VarExpr")
	}

	if _, ok := bin.Right.(*boogie.IntLit); !ok {
		t.Errorf("expected right operand IntLit")
	}
}

func assertIfGtZero(t *testing.T, stmt boogie.Stmt) {
	t.Helper()

	ifs, ok := stmt.(*boogie.If)
	if !ok {
		t.Fatalf("expected If, got %T", stmt)
	}

	cond, ok := ifs.Cond.(*boogie.BinOp)
	if !ok {
		t.Fatalf("expected condition BinOp, got %T", ifs.Cond)
	}

	if cond.Op != boogie.Gt {
		t.Errorf("expected '>' op, got %v", cond.Op)
	}

	if len(ifs.Then) != 1 || len(ifs.Else) != 1 {
		t.Errorf("expected 1 stmt in then and else branches")
	}
}

func assertReturnVar(t *testing.T, stmt boogie.Stmt, name string) {
	t.Helper()

	ret, ok := stmt.(*boogie.Return)
	if !ok {
		t.Fatalf("expected Return, got %T", stmt)
	}

	if len(ret.Values) != 1 {
		t.Fatalf("expected 1 return value, got %d", len(ret.Values))
	}

	v, ok := ret.Values[0].(*boogie.VarExpr)
	if !ok || v.V.Name != name {
		t.Fatalf("expected return VarExpr %q, got %T (%v)", name, ret.Values[0], v)
	}
}
