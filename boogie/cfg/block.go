package cfg

import "github.com/ezrantn/boogo/boogie"

type BlockID int

type Block struct {
	ID    BlockID
	Label string
	Stmts []boogie.Stmt
	Term  Terminator
}


