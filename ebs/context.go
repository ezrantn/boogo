package ebs

import "github.com/ezrantn/boogo/boogie"

type Definition struct {
	ID   boogie.SymbolID
	Name string
	Kind string // "local", "param", "proc"
	Ty   boogie.Type
}

type TyCtx struct {
	Symbols     map[boogie.SymbolID]*Definition
	Resolutions map[boogie.SymbolID]Definition
	NodeTypes   map[boogie.SymbolID]boogie.Type
}

func NewTyCtx() *TyCtx {
	return &TyCtx{
		Resolutions: make(map[boogie.SymbolID]Definition),
		NodeTypes:   make(map[boogie.SymbolID]boogie.Type),
	}
}
