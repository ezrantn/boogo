package cfg

import (
	"fmt"

	"github.com/ezrantn/boogo/boogie"
)

type CFG struct {
	Entry  BlockID
	Blocks map[BlockID]*Block

	Pred map[BlockID][]BlockID
	Succ map[BlockID][]BlockID
}

func BuildCFG(blocks []*Block, entry BlockID) *CFG {
	cfg := &CFG{
		Entry:  entry,
		Blocks: make(map[BlockID]*Block),
		Pred:   make(map[BlockID][]BlockID),
		Succ:   make(map[BlockID][]BlockID),
	}

	for _, b := range blocks {
		cfg.Blocks[b.ID] = b
	}

	for _, b := range blocks {
		switch t := b.Term.(type) {
		case *Goto:
			for _, tgt := range t.Targets {
				cfg.Succ[b.ID] = append(cfg.Succ[b.ID], tgt)
				cfg.Pred[tgt] = append(cfg.Pred[tgt], b.ID)
			}

		case *If:
			cfg.Succ[b.ID] = append(cfg.Succ[b.ID], t.Then, t.Else)
			cfg.Pred[t.Then] = append(cfg.Pred[t.Then], b.ID)
			cfg.Pred[t.Else] = append(cfg.Pred[t.Else], b.ID)

		case *Return:
			// no successors
		}
	}

	return cfg
}

func Structure(cfg *CFG) ([]boogie.Stmt, error) {
	visited := make(map[BlockID]bool)
	return structBlock(cfg, cfg.Entry, visited)
}

func structBlock(cfg *CFG, id BlockID, seen map[BlockID]bool) ([]boogie.Stmt, error) {
	if seen[id] {
		return nil, fmt.Errorf("cycle not structured")
	}
	seen[id] = true

	b := cfg.Blocks[id]
	stmts := append([]boogie.Stmt{}, b.Stmts...)

	switch t := b.Term.(type) {

	case *Return:
		return append(stmts, &boogie.Return{Values: t.Values}), nil

	case *If:
		thenStmts, err := structBlock(cfg, t.Then, copySeen(seen))
		if err != nil {
			return nil, err
		}

		elseStmts, err := structBlock(cfg, t.Else, copySeen(seen))
		if err != nil {
			return nil, err
		}

		return append(stmts, &boogie.If{
			Cond: t.Cond,
			Then: thenStmts,
			Else: elseStmts,
		}), nil

	case *Goto:
		if len(t.Targets) == 1 {
			return structBlock(cfg, t.Targets[0], seen)
		}
		return nil, fmt.Errorf("unsupported goto")
	}

	return nil, fmt.Errorf("unsupported terminator: %T", b.Term)
}

func copySeen(seen map[BlockID]bool) map[BlockID]bool {
	cp := make(map[BlockID]bool, len(seen))
	for k, v := range seen {
		cp[k] = v
	}
	return cp
}
