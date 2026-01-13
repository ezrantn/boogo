package ebs

import (
	"fmt"

	"github.com/ezrantn/boogo/boogie"
)

func Check(p *boogie.Program) error {
	// Map procedures for global lookup
	procMap := make(map[string]*boogie.Procedure)
	for _, proc := range p.Procs {
		if _, ok := procMap[proc.Name]; ok {
			return fmt.Errorf("duplicate procedure: %s", proc.Name)
		}
		procMap[proc.Name] = proc
	}

	// Run the checks
	for _, proc := range p.Procs {
		tcx := NewTyCtx()
		// Pass BOTH the procedure-specific tcx and the global procMap
		// We assume procedures are checked in isolation
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
		tcx.Resolutions[id] = Definition{
			Name: p.Name,
			Kind: "param",
		}
	}

	// Register Return Values (In Boogie, these are accessible like locals)
	for i := range proc.Rets {
		r := &proc.Rets[i]
		if r.Name == "" {
			return fmt.Errorf("return var %d has no name", i)
		}
		id := res.Define(r.Name)
		tcx.NodeTypes[id] = r.Ty
		tcx.Resolutions[id] = Definition{
			Name: r.Name,
			Kind: "local",
		}
	}

	// Register Locals
	for i := range proc.Locals {
		l := &proc.Locals[i]
		if l.Name == "" {
			return fmt.Errorf("local %d has no name", i)
		}

		id := res.Define(l.Name)
		tcx.NodeTypes[id] = l.Ty
		tcx.Resolutions[id] = Definition{
			Name: l.Name,
			Kind: "local",
		}
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
	case *boogie.HeapRead:
		if err := resolveAndCheckExpr(ex.Obj, res, tcx); err != nil {
			return err
		}

		return checkHeapRead(ex)
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
		if err := resolveAndCheckExpr(st.Cond, res, tcx); err != nil {
			return fmt.Errorf("if condition resolution: %w", err)
		}

		if err := requireBoolExpr(st.Cond); err != nil {
			return fmt.Errorf("if condition type error: %w", err)
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

		if err := requireBoolExpr(st.Cond); err != nil {
			return err
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
			return fmt.Errorf("recursive calls are not allowed: %s calls itself", cproc.Name)
		}
		
		// Resolve Arguments
		for _, arg := range st.Args {
			if err := resolveAndCheckExpr(arg, res, tcx); err != nil {
				return err
			}
		}

		// Resolve Return Variables (LHS of the call)
		// Without this, st.Rets[i].Type() is nil, triggering the panic.
		for _, ret := range st.Rets {
			if err := resolveAndCheckExpr(ret, res, tcx); err != nil {
				return fmt.Errorf("resolving call return: %w", err)
			}
		}

		// Perform the signature/arity check
		return checkCall(st, res, tcx, procMap)

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
	case *boogie.HeapRead:
		// Resolve variables inside the expression (like the 'Obj')
		if err := resolveAndCheckExpr(st.Obj, res, tcx); err != nil {
			return err
		}

		// Perform the semantic check (ensuring Obj is a RefType)
		return checkHeapRead(st)
	case *boogie.HeapWrite:
		if err := resolveAndCheckExpr(st.Obj, res, tcx); err != nil {
			return err
		}

		if err := resolveAndCheckExpr(st.Value, res, tcx); err != nil {
			return err
		}

		return checkHeapWrite(st)

	case *boogie.Assert:
		// Resolve the expression inside the assert (e.g., variables, ops)
		if err := resolveAndCheckExpr(st.Cond, res, tcx); err != nil {
			return fmt.Errorf("assert expression: %w", err)
		}

		// Ensure the expression evaluates to a boolean
		if err := requireBoolExpr(st.Cond); err != nil {
			return fmt.Errorf("assert condition must be bool: %w", err)
		}

		return nil

	default:
		return fmt.Errorf("unsupported statement type: %T", s)
	}
}

// ========================
// Expression Checking
// ========================

func requireBoolExpr(e boogie.Expr) error {
	if e.Type() == nil {
		panic("internal error: expression has no type")
	}

	if e.Type().Kind() != boogie.BoolKind {
		return fmt.Errorf("expected Bool expression, got %v", e.Type().Kind())
	}
	return nil
}

// ========================
// Specific Checks
// ========================

func checkAssign(a *boogie.Assign) error {
	if a.Lhs.Type() == nil || a.Rhs.Type() == nil {
		panic("internal error: assignment with untyped expression")
	}

	if !sameType(a.Lhs.Type(), a.Rhs.Type()) {
		return fmt.Errorf("type mismatch in assignment")
	}
	return nil
}

func checkCall(
	c *boogie.Call,
	res *Resolver,
	tcx *TyCtx,
	procMap map[string]*boogie.Procedure,
) error {

	target, ok := procMap[c.Name]
	if !ok {
		return fmt.Errorf("call to unknown procedure: %s", c.Name)
	}

	// arguments
	if len(c.Args) != len(target.Params) {
		return fmt.Errorf(
			"arity mismatch in call to %s: expected %d args, got %d",
			c.Name, len(target.Params), len(c.Args),
		)
	}

	for i, arg := range c.Args {
		if arg.Type() == nil {
			panic("internal error: untyped call argument")
		}

		if !sameType(arg.Type(), target.Params[i].Ty) {
			return fmt.Errorf(
				"argument %d type mismatch in call to %s: expected %v, got %v",
				i, c.Name,
				target.Params[i].Ty.Kind(),
				arg.Type().Kind(),
			)
		}
	}

	// return values
	if len(c.Rets) != len(target.Rets) {
		return fmt.Errorf(
			"return arity mismatch in call to %s: expected %d returns, got %d",
			c.Name, len(target.Rets), len(c.Rets),
		)
	}

	for i, ret := range c.Rets {
		// ret must already be resolved
		if ret.Type() == nil {
			panic("internal error: unresolved call return variable")
		}

		if !sameType(ret.Type(), target.Rets[i].Ty) {
			return fmt.Errorf(
				"return value %d type mismatch in call to %s: expected %v, got %v",
				i, c.Name,
				target.Rets[i].Ty.Kind(),
				ret.Type().Kind(),
			)
		}
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
		if err := requireBool(u.X); err != nil {
			return err
		}

		u.Ty = boogie.BoolType{}
		return nil
	case boogie.Neg:
		if err := requireInt(u.X); err != nil {
			return err
		}

		u.Ty = boogie.IntType{}
		return nil
	default:
		return fmt.Errorf("unsupported unary operator")
	}
}

func checkHeapRead(h *boogie.HeapRead) error {
	if h.Obj.Type().Kind() != boogie.RefKind {
		return fmt.Errorf("heap object must be ref type")
	}

	return nil
}

func checkHeapWrite(h *boogie.HeapWrite) error {
	if h.Obj.Type().Kind() != boogie.RefKind {
		return fmt.Errorf("heap object must be ref type")
	}

	return nil
}

// ========================
// Utilities
// ========================

// For checking type equality
func sameType(a, b boogie.Type) bool {
	return a.Kind() == b.Kind()
}

func requireInt(e boogie.Expr) error {
	if e.Type().Kind() != boogie.IntKind {
		return fmt.Errorf("type error: expected Int, got %s",
			e.Type())
	}

	return nil
}

func requireBool(e boogie.Expr) error {
	if e.Type().Kind() != boogie.BoolKind {
		return fmt.Errorf("type error: expected Bool, got %s",
			e.Type())
	}

	return nil
}
