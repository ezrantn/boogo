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

		case *boogie.Assign, *boogie.Assert, *boogie.Assume, *boogie.Call:
			cur.Stmts = append(cur.Stmts, s)

		case *boogie.Return:
			cur.Term = &Return{Values: t.Values}
			return nil, nil

		case *boogie.If:
			return lowerIf(ctx, cur, t)

		case *boogie.While:
			return lowerWhile(ctx, cur, t)

		case *boogie.LocalDecl:
			// Append it to Stmts so the generator knows to declare it in Go,
			// or just continue if declarations are handled separately.
			cur.Stmts = append(cur.Stmts, s)

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
	path := make(map[BlockID]bool)
	return structBlock(cfg, cfg.Entry, visited, path)
}

func structBlock(cfg *CFG, id BlockID, seen map[BlockID]bool, path map[BlockID]bool) ([]boogie.Stmt, error) {
	// Detect infinite recursion/unstructured cycles
	if path[id] {
		return nil, fmt.Errorf("unstructured cycle detected at block %d", id)
	}

	// Stop if we've already processed this block in this branch (join point)
	if seen[id] {
		return nil, nil
	}

	b, ok := cfg.Blocks[id]
	if !ok {
		return nil, nil
	}

	seen[id] = true
	path[id] = true                     // Push to recursion stack
	defer func() { path[id] = false }() // Pop when done

	var result []boogie.Stmt
	result = append(result, b.Stmts...)

	switch t := b.Term.(type) {
	case *Return:
		result = append(result, &boogie.Return{Values: t.Values})
		return result, nil

	case *If:
		joinID := findJoinPoint(cfg, id)

		// Prepare 'seen' maps for the branches.
		// We mark joinID as 'seen' so the recursive calls stop before entering it.
		branchSeen := copySeen(seen)
		if joinID != 0 {
			branchSeen[joinID] = true
		}

		// Process the Then and Else branches.
		// They will stop once they reach the join point.
		thenStmts, err := structBlock(cfg, t.Then, branchSeen, path)
		if err != nil {
			return nil, err
		}

		elseStmts, err := structBlock(cfg, t.Else, branchSeen, path)
		if err != nil {
			return nil, err
		}

		// Append the structured If to our results.
		result = append(result, &boogie.If{
			Cond: t.Cond,
			Then: thenStmts,
			Else: elseStmts,
		})

		// 4. Continue structuring from the join point.
		// This ensures the 'return' (or whatever code is at the join)
		// appears exactly once after the if-block.
		if joinID != 0 {
			rest, err := structBlock(cfg, joinID, seen, path)
			if err != nil {
				return nil, err
			}
			result = append(result, rest...)
		}
		return result, nil

	case *Goto:
		if len(t.Targets) == 1 {
			tgt := t.Targets[0]
			if seen[tgt] && !path[tgt] {
				return result, nil // Reached a join point
			}

			next, err := structBlock(cfg, tgt, seen, path)
			if err != nil {
				return nil, err
			}
			result = append(result, next...)
		}
	}
	return result, nil
}

func checkAndPullReturn(cfg *CFG, joinID BlockID) []boogie.Stmt {
	if joinID == 0 {
		return nil
	}
	jb := cfg.Blocks[joinID]
	if ret, ok := jb.Term.(*Return); ok {
		var stmts []boogie.Stmt
		stmts = append(stmts, jb.Stmts...)
		stmts = append(stmts, &boogie.Return{Values: ret.Values})
		return stmts
	}
	return nil
}

func isSimpleReturn(term Terminator) bool {
	ret, ok := term.(*Return)
	if !ok {
		return false
	}
	// It's a "simple" auto-generated return if it has no values
	return len(ret.Values) == 0
}

func copySeen(seen map[BlockID]bool) map[BlockID]bool {
	cp := make(map[BlockID]bool, len(seen))
	for k, v := range seen {
		cp[k] = v
	}

	return cp
}

func findJoinPoint(cfg *CFG, id BlockID) BlockID {
	// We look for a block reachable from 'id' that has multiple
	// predecessors, implying it is a merge point for an If or a Loop.
	visited := make(map[BlockID]bool)
	queue := []BlockID{id}

	for len(queue) > 0 {
		curr := queue[0]
		queue = queue[1:]

		// If this block has multiple incoming edges, it's our join point.
		if len(cfg.Pred[curr]) > 1 {
			return curr
		}

		for _, next := range cfg.Succ[curr] {
			if !visited[next] {
				visited[next] = true
				queue = append(queue, next)
			}
		}
	}
	return 0
}
