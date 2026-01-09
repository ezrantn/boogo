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

type lowerCtx struct {
	nextID BlockID
	blocks []*Block
}

func newLowerCtx() *lowerCtx {
	return &lowerCtx{nextID: 0}
}

func (c *lowerCtx) newBlock() *Block {
	b := &Block{
		ID: c.nextID,
	}
	c.nextID++
	c.blocks = append(c.blocks, b)
	return b
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

func LowerProcToCFG(p *boogie.Procedure) ([]*Block, BlockID, error) {
	ctx := newLowerCtx()

	entry := ctx.newBlock()
	last, err := lowerStmts(ctx, entry, p.Body)
	if err != nil {
		return nil, 0, err
	}

	if last != nil && last.Term == nil {
		last.Term = &Return{}
	}

	return ctx.blocks, entry.ID, nil
}

func lowerStmts(ctx *lowerCtx, cur *Block, stmts []boogie.Stmt) (*Block, error) {
	for _, s := range stmts {
		switch t := s.(type) {

		case *boogie.Assign, *boogie.Assert, *boogie.Assume:
			cur.Stmts = append(cur.Stmts, s)

		case *boogie.Return:
			cur.Term = &Return{Values: t.Values}
			return nil, nil

		case *boogie.If:
			return lowerIf(ctx, cur, t)

		case *boogie.While:
			return lowerWhile(ctx, cur, t)

		default:
			return nil, fmt.Errorf("unsupported stmt in EBS: %T", s)
		}
	}

	return cur, nil
}

func lowerIf(ctx *lowerCtx, cur *Block, s *boogie.If) (*Block, error) {
	thenBlock := ctx.newBlock()
	elseBlock := ctx.newBlock()
	join := ctx.newBlock()

	cur.Term = &If{
		Cond: s.Cond,
		Then: thenBlock.ID,
		Else: elseBlock.ID,
	}

	thenEnd, err := lowerStmts(ctx, thenBlock, s.Then)
	if err != nil {
		return nil, err
	}
	if thenEnd != nil {
		thenEnd.Term = &Goto{Targets: []BlockID{join.ID}}
	}

	elseEnd, err := lowerStmts(ctx, elseBlock, s.Else)
	if err != nil {
		return nil, err
	}
	if elseEnd != nil {
		elseEnd.Term = &Goto{Targets: []BlockID{join.ID}}
	}

	return join, nil
}

func lowerWhile(ctx *lowerCtx, cur *Block, s *boogie.While) (*Block, error) {
	header := ctx.newBlock()
	body := ctx.newBlock()
	exit := ctx.newBlock()

	// jump to header
	cur.Term = &Goto{Targets: []BlockID{header.ID}}

	// condition
	header.Term = &If{
		Cond: s.Cond,
		Then: body.ID,
		Else: exit.ID,
	}

	bodyEnd, err := lowerStmts(ctx, body, s.Body)
	if err != nil {
		return nil, err
	}
	if bodyEnd != nil {
		bodyEnd.Term = &Goto{Targets: []BlockID{header.ID}}
	}

	return exit, nil
}

func Structure(cfg *CFG) ([]boogie.Stmt, error) {
	visited := make(map[BlockID]bool)
	// TODO:
	// We need a way to find where branches meet (Post-Dominators)
	// For now, let's focus on the recursive fix.
	return structBlock(cfg, cfg.Entry, visited)
}

func structBlock(cfg *CFG, id BlockID, seen map[BlockID]bool) ([]boogie.Stmt, error) {
	if seen[id] {
		return nil, nil // Stop recursion if we've seen this (loop header or join)
	}

	b, ok := cfg.Blocks[id]
	if !ok {
		return nil, nil
	}

	seen[id] = true
	var result []boogie.Stmt
	result = append(result, b.Stmts...)

	switch t := b.Term.(type) {
	case *Return:
		result = append(result, &boogie.Return{Values: t.Values})
		return result, nil

	case *If:
		thenStmts, _ := structBlock(cfg, t.Then, copySeen(seen))
		elseStmts, _ := structBlock(cfg, t.Else, copySeen(seen))

		result = append(result, &boogie.If{
			Cond: t.Cond,
			Then: thenStmts,
			Else: elseStmts,
		})

		// TODO:
		// IMPORTANT: In our lowerIf, we have a 'join' block.
		// We need to find it and continue structuring FROM there
		// after the If statement is closed.
		// joinID := findJoinPoint(t.Then, t.Else)
		// rest, _ := structBlock(cfg, joinID, seen)
		// result = append(result, rest...)

	case *Goto:
		if len(t.Targets) == 1 {
			next, err := structBlock(cfg, t.Targets[0], seen)
			if err != nil {
				return nil, err
			}
			result = append(result, next...)
		}
	}
	return result, nil
}

func copySeen(seen map[BlockID]bool) map[BlockID]bool {
	cp := make(map[BlockID]bool, len(seen))
	for k, v := range seen {
		cp[k] = v
	}

	return cp
}
