package ebs

import "github.com/ezrantn/boogo/boogie"

type Resolver struct {
	scopes []map[string]boogie.SymbolID
	nextID boogie.SymbolID
}

func NewResolver() *Resolver {
	return &Resolver{
		scopes: []map[string]boogie.SymbolID{make(map[string]boogie.SymbolID)},
		nextID: 1,
	}
}

func (r *Resolver) EnterScope() {
	r.scopes = append(r.scopes, make(map[string]boogie.SymbolID))
}

func (r *Resolver) ExitScope() {
	r.scopes = r.scopes[:len(r.scopes)-1]
}

// Define binds a name to a new unique ID in the current scope
func (r *Resolver) Define(name string) boogie.SymbolID {
	if name == "" {
		panic("EBS: Attempted to define a variable with an empty name")
	}

	id := r.nextID
	r.nextID++
	current := r.scopes[len(r.scopes)-1]
	current[name] = id
	return id
}

// Resolve looks up a name in the scope stack
func (r *Resolver) Resolve(name string) (boogie.SymbolID, bool) {
	for i := len(r.scopes) - 1; i >= 0; i-- {
		if id, ok := r.scopes[i][name]; ok {
			return id, true
		}
	}
	return 0, false
}
