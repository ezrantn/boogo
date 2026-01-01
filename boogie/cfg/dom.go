package cfg

func ComputeDominators(cfg *CFG) map[BlockID]map[BlockID]bool {
	dom := make(map[BlockID]map[BlockID]bool)

	// init
	for id := range cfg.Blocks {
		dom[id] = make(map[BlockID]bool)
		for j := range cfg.Blocks {
			dom[id][j] = true
		}
	}

	// entry dominates itself
	dom[cfg.Entry] = map[BlockID]bool{cfg.Entry: true}

	changed := true
	for changed {
		changed = false

		for b := range cfg.Blocks {
			if b == cfg.Entry {
				continue
			}

			newDom := make(map[BlockID]bool)
			first := true

			for _, p := range cfg.Pred[b] {
				if first {
					for x := range dom[p] {
						newDom[x] = true
					}
					first = false
				} else {
					for x := range newDom {
						if !dom[p][x] {
							delete(newDom, x)
						}
					}
				}
			}

			newDom[b] = true

			if !equalSet(dom[b], newDom) {
				dom[b] = newDom
				changed = true
			}
		}
	}

	return dom
}

func equalSet(a, b map[BlockID]bool) bool {
	if len(a) != len(b) {
		return false
	}
	for k := range a {
		if !b[k] {
			return false
		}
	}
	return true
}
