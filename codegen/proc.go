package codegen

import (
	"strings"

	"github.com/ezrantn/boogo/boogie"
)

// EmitProc emits Go code for a Boogie procedure.
func EmitProc(p *boogie.Procedure) string {
	var b strings.Builder

	// Function signature
	b.WriteString("func ")
	b.WriteString(p.Name)
	b.WriteString("(")
	b.WriteString(emitParams(p.Params))
	b.WriteString(")")

	if len(p.Rets) > 0 {
		b.WriteString(" ")
		b.WriteString(emitReturns(p.Rets))
	}

	b.WriteString(" {\n")

	// Local variable declarations
	if len(p.Locals) > 0 {
		for _, v := range p.Locals {
			b.WriteString("\tvar ")
			b.WriteString(v.Name)
			b.WriteString(" ")
			b.WriteString(goType(v.Ty))
			b.WriteString("\n")
		}
		b.WriteString("\n")
	}

	// Body
	b.WriteString(EmitStmts(p.Body, 1))

	b.WriteString("}\n\n")
	return b.String()
}

// ========================
// Helpers
// ========================

func emitParams(vars []boogie.Var) string {
	var ps []string
	for _, v := range vars {
		ps = append(ps, v.Name+" "+goType(v.Ty))
	}
	return strings.Join(ps, ", ")
}

func emitReturns(vars []boogie.Var) string {
	if len(vars) == 1 {
		return goType(vars[0].Ty)
	}

	var rs []string
	for _, v := range vars {
		rs = append(rs, goType(v.Ty))
	}
	return "(" + strings.Join(rs, ", ") + ")"
}

// goType maps Boogie types to Go types (EBS v1).
func goType(t boogie.Type) string {
	switch t {

	case boogie.IntType{}:
		return "int"

	case boogie.BoolType{}:
		return "bool"

	case boogie.RefType{}:
		return "int" // object references

	default:
		panic("unsupported type in codegen")
	}
}
