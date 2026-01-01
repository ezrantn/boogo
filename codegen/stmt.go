package codegen

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/ezrantn/boogo/boogie"
)

// EmitStmt emits Go code for a single Boogie statement.
func EmitStmt(s boogie.Stmt, indent int) string {
	switch st := s.(type) {

	case *boogie.Assign:
		return emitAssign(st, indent)

	case *boogie.If:
		return emitIf(st, indent)

	case *boogie.While:
		return emitWhile(st, indent)

	case *boogie.Call:
		return emitCall(st, indent)

	case *boogie.Return:
		return emitReturn(st, indent)

	case *boogie.HeapWrite:
		return emitHeapWrite(st, indent)

	default:
		panic(fmt.Sprintf("unsupported statement in codegen: %T", s))
	}
}

// EmitStmts emits a sequence of statements.
func EmitStmts(stmts []boogie.Stmt, indent int) string {
	var b strings.Builder
	for _, s := range stmts {
		b.WriteString(EmitStmt(s, indent))
	}
	return b.String()
}

// ========================
// Individual statements
// ========================

func emitAssign(a *boogie.Assign, indent int) string {
	lhs := EmitExpr(a.Lhs)
	rhs := EmitExpr(a.Rhs)
	return indentStr(indent) + lhs + " = " + rhs + "\n"
}

func emitIf(i *boogie.If, indent int) string {
	var b strings.Builder

	cond := EmitExpr(i.Cond)
	b.WriteString(indentStr(indent) + "if " + cond + " {\n")
	b.WriteString(EmitStmts(i.Then, indent+1))
	b.WriteString(indentStr(indent) + "} else {\n")
	b.WriteString(EmitStmts(i.Else, indent+1))
	b.WriteString(indentStr(indent) + "}\n")

	return b.String()
}

func emitWhile(w *boogie.While, indent int) string {
	var b strings.Builder

	cond := EmitExpr(w.Cond)
	b.WriteString(indentStr(indent) + "for " + cond + " {\n")
	b.WriteString(EmitStmts(w.Body, indent+1))
	b.WriteString(indentStr(indent) + "}\n")

	return b.String()
}

func emitCall(c *boogie.Call, indent int) string {
	var args []string
	for _, a := range c.Args {
		args = append(args, EmitExpr(a))
	}

	return indentStr(indent) +
		c.Name +
		"(" + strings.Join(args, ", ") + ")\n"
}

func emitReturn(r *boogie.Return, indent int) string {
	if len(r.Values) == 0 {
		return indentStr(indent) + "return\n"
	}

	var vals []string
	for _, v := range r.Values {
		vals = append(vals, EmitExpr(v))
	}

	return indentStr(indent) +
		"return " + strings.Join(vals, ", ") + "\n"
}

func emitHeapWrite(h *boogie.HeapWrite, indent int) string {
	obj := EmitExpr(h.Obj)
	field := strconv.Quote(h.Field)
	val := EmitExpr(h.Value)

	// heapWrite(obj, "field", value)
	return indentStr(indent) +
		"heapWrite(" + obj + ", " + field + ", " + val + ")\n"
}

// ========================
// Utilities
// ========================

func indentStr(n int) string {
	return strings.Repeat("\t", n)
}
