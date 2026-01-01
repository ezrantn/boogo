package ebs

import (
	"fmt"

	"github.com/ezrantn/boogo/boogie"
)

func Check(p *boogie.Program) error {
	procMap := make(map[string]*boogie.Procedure)
	for _, proc := range p.Procs {
		if _, ok := procMap[proc.Name]; ok {
			return fmt.Errorf("duplicate procedure: %s", proc.Name)
		}

		procMap[proc.Name] = proc
	}

	for _, proc := range p.Procs {
		if err := checkProcedure(proc, procMap); err != nil {
			return fmt.Errorf("procedure %s: %w", proc.Name, err)
		}
	}

	return nil
}

// ========================
// Procedure Checking
// ========================

func checkProcedure(
	proc *boogie.Procedure,
	procMap map[string]*boogie.Procedure,
) error {

	// Reject recursion (direct)
	for _, stmt := range proc.Body {
		if callsSelf(stmt, proc.Name) {
			return fmt.Errorf("recursive call is not allowed in EBS v1")
		}
	}

	// Check body
	for _, stmt := range proc.Body {
		if err := checkStmt(stmt, proc, procMap); err != nil {
			return err
		}
	}

	return nil
}

// ========================
// Statement Checking
// ========================

func checkStmt(
	s boogie.Stmt,
	proc *boogie.Procedure,
	procMap map[string]*boogie.Procedure,
) error {

	switch st := s.(type) {

	case *boogie.Assign:
		return checkAssign(st)

	case *boogie.If:
		if err := checkExprBool(st.Cond); err != nil {
			return fmt.Errorf("if condition: %w", err)
		}
		for _, t := range st.Then {
			if err := checkStmt(t, proc, procMap); err != nil {
				return err
			}
		}
		for _, e := range st.Else {
			if err := checkStmt(e, proc, procMap); err != nil {
				return err
			}
		}
		return nil

	case *boogie.While:
		if err := checkExprBool(st.Cond); err != nil {
			return fmt.Errorf("while condition: %w", err)
		}
		for _, b := range st.Body {
			if err := checkStmt(b, proc, procMap); err != nil {
				return err
			}
		}
		return nil

	case *boogie.Call:
		return checkCall(st, procMap)

	case *boogie.Return:
		if len(st.Values) != len(proc.Rets) {
			return fmt.Errorf("return arity mismatch: expected %d values, got %d", len(proc.Rets), len(st.Values))
		}

		for i, v := range st.Values {
			if err := checkExpr(v); err != nil {
				return err
			}
			if !sameType(v.Type(), proc.Rets[i].Ty) {
				return fmt.Errorf("return type mismatch at index %d: expected %T, got %T", i, proc.Rets[i].Ty, v.Type())
			}
		}

		return nil

	case *boogie.Assert:
		// Verification-only, but must be well-typed
		return checkExprBool(st.Cond)

	case *boogie.HeapWrite:
		return checkHeapWrite(st)

	default:
		return fmt.Errorf("unsupported statement in EBS v1: %T", s)
	}
}

// ========================
// Expression Checking
// ========================

func checkExpr(e boogie.Expr) error {
	switch ex := e.(type) {

	case *boogie.VarExpr:
		return nil

	case *boogie.IntLit:
		return nil

	case *boogie.BoolLit:
		return nil

	case *boogie.BinOp:
		if err := checkExpr(ex.Left); err != nil {
			return err
		}
		if err := checkExpr(ex.Right); err != nil {
			return err
		}
		return checkBinOp(ex)

	case *boogie.UnOp:
		if err := checkExpr(ex.X); err != nil {
			return err
		}
		return checkUnOp(ex)

	case *boogie.HeapRead:
		return checkHeapRead(ex)

	default:
		return fmt.Errorf("unsupported expression in EBS v1: %T", e)
	}
}

func checkExprBool(e boogie.Expr) error {
	if err := checkExpr(e); err != nil {
		return err
	}
	if _, ok := e.Type().(boogie.BoolType); !ok {
		return fmt.Errorf("expected bool expression")
	}
	return nil
}

// ========================
// Specific Checks
// ========================

func checkAssign(a *boogie.Assign) error {
	if err := checkExpr(a.Lhs); err != nil {
		return err
	}
	if err := checkExpr(a.Rhs); err != nil {
		return err
	}

	if !sameType(a.Lhs.Type(), a.Rhs.Type()) {
		return fmt.Errorf("type mismatch in assignment")
	}
	return nil
}

func checkCall(c *boogie.Call, procMap map[string]*boogie.Procedure) error {
	target, ok := procMap[c.Name]
	if !ok {
		return fmt.Errorf("call to unknown procedure: %s", c.Name)
	}

	if len(c.Args) != len(target.Params) {
		return fmt.Errorf("arity mismatch in call to %s", c.Name)
	}

	for i, arg := range c.Args {
		if err := checkExpr(arg); err != nil {
			return err
		}
		if !sameType(arg.Type(), target.Params[i].Ty) {
			return fmt.Errorf("argument %d type mismatch in call to %s", i, c.Name)
		}
	}

	if len(c.Rets) != len(target.Rets) {
		return fmt.Errorf("return arity mismatch in call to %s", c.Name)
	}

	return nil
}

func checkBinOp(b *boogie.BinOp) error {
	switch b.Op {
	case boogie.Add, boogie.Sub, boogie.Mul:
		if err := requireInt(b.Left); err != nil {
			return err
		}

		if err := requireInt(b.Right); err != nil {
			return err
		}

		return requireType(b.Ty, boogie.IntType{})

	case boogie.Eq, boogie.Lt, boogie.Le:
		if !sameType(b.Left.Type(), b.Right.Type()) {
			return fmt.Errorf("binary op operands must have same type")
		}

		return requireType(b.Ty, boogie.BoolType{})

	case boogie.And, boogie.Or:
		if err := requireBool(b.Left); err != nil {
			return err
		}

		if err := requireBool(b.Right); err != nil {
			return err
		}

		return requireType(b.Ty, boogie.BoolType{})

	default:
		return fmt.Errorf("unsupported binary operator")
	}
}

func checkUnOp(u *boogie.UnOp) error {
	switch u.Op {
	case boogie.Not:
		requireBool(u.X)
		requireType(u.Ty, boogie.BoolType{})
		return nil
	case boogie.Neg:
		requireInt(u.X)
		requireType(u.Ty, boogie.IntType{})
		return nil
	default:
		return fmt.Errorf("unsupported unary operator")
	}
}

func checkHeapRead(h *boogie.HeapRead) error {
	if err := checkExpr(h.Obj); err != nil {
		return err
	}

	if _, ok := h.Obj.Type().(boogie.RefType); !ok {
		return fmt.Errorf("heap object must be ref type")
	}

	return nil
}

func checkHeapWrite(h *boogie.HeapWrite) error {
	if err := checkExpr(h.Obj); err != nil {
		return err
	}
	if _, ok := h.Obj.Type().(boogie.RefType); !ok {
		return fmt.Errorf("heap object must be ref type")
	}
	if err := checkExpr(h.Value); err != nil {
		return err
	}
	return nil
}

// ========================
// Utilities
// ========================

func sameType(a, b boogie.Type) bool {
	return fmt.Sprintf("%T", a) == fmt.Sprintf("%T", b)
}

func requireInt(e boogie.Expr) error {
	if _, ok := e.Type().(boogie.IntType); !ok {
		return fmt.Errorf("expected int expression, got %T", e.Type())
	}

	return nil
}

func requireBool(e boogie.Expr) error {
	if _, ok := e.Type().(boogie.BoolType); !ok {
		return fmt.Errorf("expected bool expression, got %T", e.Type())
	}

	return nil
}

func requireType(got boogie.Type, want boogie.Type) error {
	if fmt.Sprintf("%T", got) != fmt.Sprintf("%T", want) {
		return fmt.Errorf("unexpected type: got %T, want %T", got, want)
	}

	return nil
}

func callsSelf(s boogie.Stmt, name string) bool {
	switch st := s.(type) {
	case *boogie.Call:
		return st.Name == name
	case *boogie.If:
		for _, t := range st.Then {
			if callsSelf(t, name) {
				return true
			}
		}
		for _, e := range st.Else {
			if callsSelf(e, name) {
				return true
			}
		}
	case *boogie.While:
		for _, b := range st.Body {
			if callsSelf(b, name) {
				return true
			}
		}
	}

	return false
}
