package cfg

import "github.com/ezrantn/boogo/boogie"

type Terminator interface{ isTerm() }

type Goto struct {
	Targets []BlockID
}

func (*Goto) isTerm() {}

type If struct {
	Cond boogie.Expr
	Then BlockID
	Else BlockID
}

func (*If) isTerm() {}

type Return struct {
	Values []boogie.Expr
}

func (*Return) isTerm() {}
