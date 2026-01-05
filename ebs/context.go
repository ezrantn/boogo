package ebs

import "github.com/ezrantn/boogo/boogie"

type Definition struct {
	Name string
	Kind string // "local", "param", "proc"
}

type TyCtx struct {
	Resolutions map[boogie.SymbolID]Definition
	NodeTypes   map[boogie.SymbolID]boogie.Type
}

func NewTyCtx() *TyCtx {
	return &TyCtx{
		Resolutions: make(map[boogie.SymbolID]Definition),
		NodeTypes:   make(map[boogie.SymbolID]boogie.Type),
	}
}
