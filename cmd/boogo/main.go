package boogo

import (
	"strings"

	"github.com/ezrantn/boogo/boogie"
	"github.com/ezrantn/boogo/codegen"
)

// EmitProgram emits a complete Go source file from a Boogie program.
func EmitProgram(p *boogie.Program) string {
	var b strings.Builder

	// Package header
	b.WriteString("package main\n\n")

	// Imports (minimal, explicit)
	b.WriteString("import (\n")
	b.WriteString("\t\"fmt\"\n")
	b.WriteString(")\n\n")

	// Runtime heap
	b.WriteString(emitHeapRuntime())

	// Procedures
	for _, proc := range p.Procs {
		b.WriteString(codegen.EmitProc(proc))
	}

	// Optional: bootstrap main()
	if hasMain(p) {
		b.WriteString(emitMainWrapper())
	}

	return b.String()
}

// ========================
// Helpers
// ========================

func hasMain(p *boogie.Program) bool {
	for _, proc := range p.Procs {
		if proc.Name == "main" {
			return true
		}
	}

	return false
}

// emitMainWrapper emits a Go main() that calls Boogie main().
func emitMainWrapper() string {
	var b strings.Builder

	b.WriteString("func main() {\n")
	b.WriteString("\tmain()\n")
	b.WriteString("}\n")

	return b.String()
}

// emitHeapRuntime embeds the heap runtime.
// This assumes heap.go is part of the same package;
// in a real system this could be factored out.
func emitHeapRuntime() string {
	var b strings.Builder

	b.WriteString("// ========================\n")
	b.WriteString("// Runtime Heap\n")
	b.WriteString("// ========================\n\n")

	b.WriteString("var heap = make(map[int]map[string]interface{})\n\n")

	b.WriteString("func heapRead(obj interface{}, field string) interface{} {\n")
	b.WriteString("\tid, ok := obj.(int)\n")
	b.WriteString("\tif !ok {\n")
	b.WriteString("\t\tpanic(\"heapRead: object is not an int reference\")\n")
	b.WriteString("\t}\n\n")
	b.WriteString("\tfields, ok := heap[id]\n")
	b.WriteString("\tif !ok {\n")
	b.WriteString("\t\tpanic(\"heapRead: unknown object\")\n")
	b.WriteString("\t}\n\n")
	b.WriteString("\tval, ok := fields[field]\n")
	b.WriteString("\tif !ok {\n")
	b.WriteString("\t\tpanic(\"heapRead: unknown field\")\n")
	b.WriteString("\t}\n\n")
	b.WriteString("\treturn val\n")
	b.WriteString("}\n\n")

	b.WriteString("func heapWrite(obj interface{}, field string, value interface{}) {\n")
	b.WriteString("\tid, ok := obj.(int)\n")
	b.WriteString("\tif !ok {\n")
	b.WriteString("\t\tpanic(\"heapWrite: object is not an int reference\")\n")
	b.WriteString("\t}\n\n")
	b.WriteString("\tfields, ok := heap[id]\n")
	b.WriteString("\tif !ok {\n")
	b.WriteString("\t\tfields = make(map[string]interface{})\n")
	b.WriteString("\t\theap[id] = fields\n")
	b.WriteString("\t}\n\n")
	b.WriteString("\tfields[field] = value\n")
	b.WriteString("}\n\n")

	return b.String()
}
