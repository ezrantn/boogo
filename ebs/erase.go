package ebs

import "github.com/ezrantn/boogo/boogie"

// Erase removes verification-only constructs from a program,
// yielding an executable EBS program.
func Erase(p *boogie.Program) *boogie.Program {
	out := &boogie.Program{}

	for _, proc := range p.Procs {
		out.Procs = append(out.Procs, eraseProc(proc))
	}

	return out
}

func eraseProc(p *boogie.Procedure) *boogie.Procedure {
	np := &boogie.Procedure{
		Name:   p.Name,
		Params: p.Params,
		Rets:   p.Rets,
		Locals: p.Locals,
		Body:   eraseStmts(p.Body),
	}
	return np
}

func eraseStmts(stmts []boogie.Stmt) []boogie.Stmt {
	var out []boogie.Stmt
	for _, s := range stmts {
		switch st := s.(type) {

		case *boogie.Assert:
			// erase verification-only assertion
			continue

		case *boogie.If:
			out = append(out, &boogie.If{
				Cond: st.Cond,
				Then: eraseStmts(st.Then),
				Else: eraseStmts(st.Else),
			})

		case *boogie.While:
			out = append(out, &boogie.While{
				Cond: st.Cond,
				Body: eraseStmts(st.Body),
			})

		default:
			// executable statement
			out = append(out, s)
		}
	}
	return out
}
