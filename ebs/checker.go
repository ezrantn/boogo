package ebs

import (
	"fmt"

	"github.com/ezrantn/boogo/boogie"
)

func Check(p *boogie.Program) error {
	// 1. Map procedures for global lookup
	procMap := make(map[string]*boogie.Procedure)
	for _, proc := range p.Procs {
		if _, ok := procMap[proc.Name]; ok {
			return fmt.Errorf("duplicate procedure: %s", proc.Name)
		}
		procMap[proc.Name] = proc
	}

	// 2. Create the Type Context (Symbol Table)
	tcx := NewTyCtx()

	// 3. Run the checks
	for _, proc := range p.Procs {
		// Pass BOTH the procedure-specific tcx and the global procMap
		if err := checkProcedure(proc, tcx, procMap); err != nil {
			return fmt.Errorf("procedure %s: %w", proc.Name, err)
		}
	}

	return nil
}

// ========================
// Procedure Checking
// ========================

func checkProcedure(proc *boogie.Procedure, tcx *TyCtx, procMap map[string]*boogie.Procedure) error {
	res := NewResolver()

	// Register Parameters
	for i := range proc.Params {
		p := &proc.Params[i]
		if p.Name == "" {
			return fmt.Errorf("param %d has no name", i)
		}
		id := res.Define(p.Name)
		tcx.NodeTypes[id] = p.Ty
		tcx.Resolutions[id] = Definition{p.Name, "param"}
	}

	// Register Return Values (In Boogie, these are accessible like locals)
	for i := range proc.Rets {
		r := &proc.Rets[i]
		if r.Name == "" {
			return fmt.Errorf("return var %d has no name", i)
		}
		id := res.Define(r.Name)
		tcx.NodeTypes[id] = r.Ty
		tcx.Resolutions[id] = Definition{r.Name, "local"}
	}

	// Register Locals
	for i := range proc.Locals {
		l := &proc.Locals[i]
		if l.Name == "" {
			return fmt.Errorf("local %d has no name", i)
		}

		id := res.Define(l.Name)
		tcx.NodeTypes[id] = l.Ty
		tcx.Resolutions[id] = Definition{l.Name, "local"}
	}

	// Check Body
	// Inside checkProcedure...
	for _, stmt := range proc.Body {
		if err := resolveAndCheckStmt(stmt, res, tcx, procMap, proc); err != nil {
			return err
		}
	}

	return nil
}

func resolveAndCheckExpr(e boogie.Expr, res *Resolver, tcx *TyCtx) error {
	switch ex := e.(type) {
	case *boogie.VarExpr:
		nameToResolve := ex.V.Name

		id, ok := res.Resolve(nameToResolve)
		if !ok {
			return fmt.Errorf("undefined variable: %q", nameToResolve)
		}

		ex.ID = id
		ex.V.Ty = tcx.NodeTypes[id]
		return nil

	case *boogie.IntLit:
		// Ensure the literal knows it is an int
		ex.Ty = boogie.IntType{}
		return nil

	case *boogie.BoolLit:
		ex.Ty = boogie.BoolType{}
		return nil

	case *boogie.BinOp:
		if err := resolveAndCheckExpr(ex.Left, res, tcx); err != nil {
			return err
		}

		if err := resolveAndCheckExpr(ex.Right, res, tcx); err != nil {
			return err
		}

		return checkBinOp(ex)

	case *boogie.UnOp:
		if err := resolveAndCheckExpr(ex.X, res, tcx); err != nil {
			return err
		}

		return checkUnOp(ex)
	}
	return nil
}

// ========================
// Statement Checking
// ========================

func resolveAndCheckStmt(s boogie.Stmt, res *Resolver, tcx *TyCtx, procMap map[string]*boogie.Procedure, cproc *boogie.Procedure) error {
	switch st := s.(type) {

	case *boogie.Assign:
		if err := resolveAndCheckExpr(st.Rhs, res, tcx); err != nil {
			return err
		}
		if err := resolveAndCheckExpr(st.Lhs, res, tcx); err != nil {
			return err
		}
		return checkAssign(st)

	case *boogie.If:
		if err := checkExprBool(st.Cond); err != nil {
			return fmt.Errorf("if condition: %w", err)
		}

		if err := resolveAndCheckExpr(st.Cond, res, tcx); err != nil {
			return err
		}

		if _, ok := st.Cond.Type().(boogie.BoolType); !ok {
			return fmt.Errorf("if condition must be bool, got %T", st.Cond.Type())
		}

		if err := resolveAndCheckExpr(st.Cond, res, tcx); err != nil {
			return fmt.Errorf("if condition: %w", err)
		}

		res.EnterScope()
		for _, sThen := range st.Then {
			if err := resolveAndCheckStmt(sThen, res, tcx, procMap, cproc); err != nil {
				return err
			}
		}

		res.ExitScope()

		res.EnterScope()
		for _, sElse := range st.Else {
			if err := resolveAndCheckStmt(sElse, res, tcx, procMap, cproc); err != nil {
				return err
			}
		}

		res.ExitScope()
		return nil

	case *boogie.While:
		if err := resolveAndCheckExpr(st.Cond, res, tcx); err != nil {
			return err
		}

		if _, ok := st.Cond.Type().(boogie.BoolType); !ok {
			return fmt.Errorf("while condition must be bool, got %T", st.Cond.Type())
		}

		if err := resolveAndCheckExpr(st.Cond, res, tcx); err != nil {
			return fmt.Errorf("while condition: %w", err)
		}

		res.EnterScope()
		for _, b := range st.Body {
			if err := resolveAndCheckStmt(b, res, tcx, procMap, cproc); err != nil {
				return err
			}
		}

		res.ExitScope()
		return nil

	case *boogie.Call:
		// RECURSION CHECK:
		if st.Name == cproc.Name {
			return fmt.Errorf("recursive calls are not allowed: %s calls itself", cproc)
		}

		// Resolve arguments
		for _, arg := range st.Args {
			if err := resolveAndCheckExpr(arg, res, tcx); err != nil {
				return err
			}
		}

		return checkCall(st, procMap)

	case *boogie.Return:
		// Check arity (number of return values)
		if len(st.Values) == 0 {
			return nil
		}

		// Check arity if explicit values are provided
		if len(st.Values) != len(cproc.Rets) {
			return fmt.Errorf("return arity mismatch: expected %d values, got %d", len(cproc.Rets), len(st.Values))
		}

		// 2. Check types
		for i, val := range st.Values {
			if err := resolveAndCheckExpr(val, res, tcx); err != nil {
				return err
			}

			// Compare the expression type to the signature's expected type
			if !sameType(val.Type(), cproc.Rets[i].Ty) {
				return fmt.Errorf("return type mismatch at index %d: expected %T, got %T", i, cproc.Rets[i].Ty, val.Type())
			}
		}

		return nil

	case *boogie.LocalDecl:
		id := res.Define(st.V.Name)
		tcx.NodeTypes[id] = st.V.Ty
		tcx.Resolutions[id] = Definition{Name: st.V.Name, Kind: "local"}
		return nil

	default:
		return nil
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

		b.Ty = boogie.IntType{}
		return nil

	case boogie.Eq, boogie.Lt, boogie.Lte, boogie.Gt, boogie.Gte:
		if !sameType(b.Left.Type(), b.Right.Type()) {
			return fmt.Errorf("binary op operands must have same type")
		}

		b.Ty = boogie.BoolType{}
		return nil

	case boogie.And, boogie.Or:
		if err := requireBool(b.Left); err != nil {
			return err
		}

		if err := requireBool(b.Right); err != nil {
			return err
		}

		b.Ty = boogie.BoolType{}
		return nil

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
