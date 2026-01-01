package boogie

func Erase(stmts []Stmt) []Stmt {
	var out []Stmt
	for _, s := range stmts {
		switch s.(type) {
		case *Assert:
			continue
		default:
			out = append(out, s)
		}
	}
	return out
}
