package cfg

import (
	"testing"

	"github.com/ezrantn/boogo/boogie"
)

func id(i int) BlockID { return BlockID(i) }

func TestBuildCFGLinear(t *testing.T) {
	b0 := &Block{
		ID:    id(0),
		Stmts: nil,
		Term:  &Goto{Targets: []BlockID{id(1)}},
	}

	b1 := &Block{
		ID:    id(1),
		Stmts: nil,
		Term:  &Return{},
	}

	cfg := BuildCFG([]*Block{b0, b1}, id(0))

	if len(cfg.Succ[id(0)]) != 1 || cfg.Succ[id(0)][0] != id(1) {
		t.Fatalf("unexpected successors for block 0: %v", cfg.Succ[id(0)])
	}

	if len(cfg.Pred[id(1)]) != 1 || cfg.Pred[id(1)][0] != id(0) {
		t.Fatalf("unexpected predecessors for block 1: %v", cfg.Pred[id(1)])
	}
}

func TestStructureLinear(t *testing.T) {
	b0 := &Block{
		ID:    id(0),
		Stmts: []boogie.Stmt{},
		Term:  &Goto{Targets: []BlockID{id(1)}},
	}
	b1 := &Block{
		ID:    id(1),
		Stmts: []boogie.Stmt{},
		Term:  &Return{},
	}

	cfg := BuildCFG([]*Block{b0, b1}, id(0))

	stmts, err := Structure(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stmts) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(stmts))
	}

	if _, ok := stmts[0].(*boogie.Return); !ok {
		t.Fatalf("expected Return, got %T", stmts[0])
	}
}

func TestStructureIf(t *testing.T) {
	cond := &boogie.BoolLit{Value: true}

	b0 := &Block{
		ID:    id(0),
		Stmts: nil,
		Term: &If{
			Cond: cond,
			Then: id(1),
			Else: id(2),
		},
	}

	b1 := &Block{
		ID:    id(1),
		Stmts: nil,
		Term:  &Return{},
	}

	b2 := &Block{
		ID:    id(2),
		Stmts: nil,
		Term:  &Return{},
	}

	cfg := BuildCFG([]*Block{b0, b1, b2}, id(0))

	stmts, err := Structure(cfg)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}

	if len(stmts) != 1 {
		t.Fatalf("expected 1 stmt, got %d", len(stmts))
	}

	ifs, ok := stmts[0].(*boogie.If)
	if !ok {
		t.Fatalf("expected If, got %T", stmts[0])
	}

	if ifs.Cond != cond {
		t.Fatalf("condition not preserved")
	}

	if len(ifs.Then) != 1 || len(ifs.Else) != 1 {
		t.Fatalf("unexpected branch sizes")
	}

	if _, ok := ifs.Then[0].(*boogie.Return); !ok {
		t.Fatalf("then branch does not end with return")
	}

	if _, ok := ifs.Else[0].(*boogie.Return); !ok {
		t.Fatalf("else branch does not end with return")
	}
}

func TestStructureRejectsCycle(t *testing.T) {
	b0 := &Block{
		ID:    id(0),
		Stmts: nil,
		Term:  &Goto{Targets: []BlockID{id(0)}}, // self-loop
	}

	cfg := BuildCFG([]*Block{b0}, id(0))

	if _, err := Structure(cfg); err == nil {
		t.Fatalf("expected cycle to be rejected")
	}
}
