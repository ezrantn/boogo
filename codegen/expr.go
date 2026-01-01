package codegen

import (
	"fmt"
	"strconv"

	"github.com/ezrantn/boogo/boogie"
)

// EmitExpr emits Go code for a Boogie expression (EBS v1 only).
func EmitExpr(e boogie.Expr) string {
	switch ex := e.(type) {

	case *boogie.VarExpr:
		return ex.V.Name

	case *boogie.IntLit:
		return strconv.Itoa(ex.Value)

	case *boogie.BoolLit:
		if ex.Value {
			return "true"
		}
		return "false"

	case *boogie.BinOp:
		return emitBinOp(ex)

	case *boogie.UnOp:
		return emitUnOp(ex)

	case *boogie.HeapRead:
		return emitHeapRead(ex)

	default:
		panic(fmt.Sprintf("unsupported expression in codegen: %T", e))
	}
}

// ========================
// Binary Operators
// ========================

func emitBinOp(b *boogie.BinOp) string {
	l := EmitExpr(b.Left)
	r := EmitExpr(b.Right)

	switch b.Op {

	case boogie.Add:
		return "(" + l + " + " + r + ")"

	case boogie.Sub:
		return "(" + l + " - " + r + ")"

	case boogie.Mul:
		return "(" + l + " * " + r + ")"

	case boogie.Eq:
		return "(" + l + " == " + r + ")"

	case boogie.Lt:
		return "(" + l + " < " + r + ")"

	case boogie.Le:
		return "(" + l + " <= " + r + ")"

	case boogie.And:
		return "(" + l + " && " + r + ")"

	case boogie.Or:
		return "(" + l + " || " + r + ")"

	default:
		panic("unsupported binary operator")
	}
}

// ========================
// Unary Operators
// ========================

func emitUnOp(u *boogie.UnOp) string {
	x := EmitExpr(u.X)

	switch u.Op {

	case boogie.Not:
		return "(!" + x + ")"

	case boogie.Neg:
		return "(-" + x + ")"

	default:
		panic("unsupported unary operator")
	}
}

// ========================
// Heap Access
// ========================

func emitHeapRead(h *boogie.HeapRead) string {
	obj := EmitExpr(h.Obj)
	field := strconv.Quote(h.Field)

	// heapRead(obj, "field")
	return "heapRead(" + obj + ", " + field + ")"
}
