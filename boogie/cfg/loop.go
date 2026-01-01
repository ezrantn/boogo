package cfg

func FindLoops(cfg *CFG, dom map[BlockID]map[BlockID]bool) map[BlockID]BlockID {
	// map: header -> backedge source
	loops := make(map[BlockID]BlockID)

	for b, succs := range cfg.Succ {
		for _, s := range succs {
			if dom[b][s] {
				// back-edge b -> s
				loops[s] = b
			}
		}
	}

	return loops
}
