package frontend

import (
	"fmt"
	"strconv"

	"github.com/ezrantn/boogo/boogie"
)

// Parse is a *minimal* Boogie frontend.
// For now, it only supports a very small executable subset
// sufficient for E2E testing.
//
// procedure p(x: int) returns (y: int)
// {
//  var z: int;
//
//  z := x + 1;
//  if (z > 0) {
//    y := z;
//  } else {
//    y := 0;
//  }
//
//  return y;
// }
//
// Explicitly rejected
//
// - axiom, invariant, requires, ensures
// - forall, exists
// - havoc
// - goto (frontend only; allowed internally via CFG)
// - call with multiple returns (for now)
// - maps other than heap encoding

type Parser struct {
	lexer *Lexer
	curr  Token
	peek  Token
}

func Parse(src []byte) (*boogie.Program, error) {
	l := NewLexer(string(src))
	p := NewParser(l)

	return p.ParseProgram(), nil
}

func NewParser(l *Lexer) *Parser {
	p := &Parser{lexer: l}
	// Read two tokens so curr and peek are both set
	p.nextToken()
	p.nextToken()
	return p
}

const (
	PREC_LOWEST      = iota
	PREC_EQUALS      // ==
	PREC_LESSGREATER // > or <
	PREC_SUM         // + or -
	PREC_PRODUCT     // *
)

var precedences = map[TokenKind]int{
	EQ:    PREC_EQUALS,
	LT:    PREC_LESSGREATER,
	LTE:   PREC_LESSGREATER,
	GT:    PREC_LESSGREATER,
	GTE:   PREC_LESSGREATER,
	PLUS:  PREC_SUM,
	MINUS: PREC_SUM,
	MUL:   PREC_PRODUCT,
}

func (p *Parser) peekPrecedence() int {
	if p, ok := precedences[p.peek.Kind]; ok {
		return p
	}
	return PREC_LOWEST
}

func (p *Parser) currPrecedence() int {
	if p, ok := precedences[p.curr.Kind]; ok {
		return p
	}
	return PREC_LOWEST
}

func (p *Parser) nextToken() {
	p.curr = p.peek
	p.peek = p.lexer.NextToken()
}

func (p *Parser) expect(kind TokenKind) {
	if p.curr.Kind == kind {
		p.nextToken()
	} else {
		panic(fmt.Sprintf("expected %v, got %v", kind, p.curr.Kind))
	}
}

func (p *Parser) ParseProgram() *boogie.Program {
	prog := &boogie.Program{}
	for p.curr.Kind != EOF {
		if p.curr.Kind == PROCEDURE {
			prog.Procs = append(prog.Procs, p.parseProcedure())
		} else {
			p.nextToken() // Skip unknown top-level decls for now
		}
	}
	return prog
}

func (p *Parser) parseProcedure() *boogie.Procedure {
	p.expect(PROCEDURE)
	name := p.curr.Value
	p.expect(IDENT)

	params := p.parseVarList()

	var rets []boogie.Var
	if p.curr.Kind == RETURNS {
		p.nextToken()
		rets = p.parseVarList()
	}

	p.expect(LBRACE)
	body := p.parseStatements()
	p.expect(RBRACE)

	return &boogie.Procedure{
		Name:   name,
		Params: params,
		Rets:   rets,
		Body:   body,
	}
}

func (p *Parser) parseVarList() []boogie.Var {
	var vars []boogie.Var
	p.expect(LPAREN)

	for p.curr.Kind != RPAREN && p.curr.Kind != EOF {
		name := p.curr.Value
		p.expect(IDENT)
		p.expect(COLON)

		ty := p.parseType()
		vars = append(vars, boogie.Var{Name: name, Ty: ty})

		if p.curr.Kind == COMMA {
			p.nextToken()
		} else {
			break
		}
	}

	p.expect(RPAREN)
	return vars
}

func (p *Parser) parseType() boogie.Type {
	typeName := p.curr.Value
	p.expect(IDENT)

	switch typeName {
	case "int":
		return boogie.IntType{}
	case "bool":
		return boogie.BoolType{}
	default:
		// You can expand this for bitvectors: case "bv32": ...
		return boogie.IntType{}
	}
}

func (p *Parser) parseExpression(precedence int) boogie.Expr {
	// Parse the "Prefix" part (identifiers, numbers, or grouping)
	left := p.parsePrimary()

	// while the next token isn't a semicolon/brace
	// and the next operator binds tighter than our current level
	for p.curr.Kind != SEMI && p.curr.Kind != RPAREN && precedence < p.currPrecedence() {
		left = p.parseInfix(left)
	}

	return left
}

func (p *Parser) parseInfix(left boogie.Expr) boogie.Expr {
	kind := p.curr.Kind
	prec := p.currPrecedence()
	p.nextToken() // consume operator

	return &boogie.BinOp{
		Op:    tokenToOp(kind),
		Left:  left,
		Right: p.parseExpression(prec),
	}
}

// Helper to map TokenKind to boogie.BinOpKind
func tokenToOp(kind TokenKind) boogie.BinOpKind {
	switch kind {
	case PLUS:
		return boogie.Add
	case MINUS:
		return boogie.Sub
	case MUL:
		return boogie.Mul
	case LT:
		return boogie.Lt
	case LTE:
		return boogie.Lte
	case EQ:
		return boogie.Eq
	case GT:
		return boogie.Gt
	case GTE:
		return boogie.Gte
	default:
		panic("unsupported operator")
	}
}

func (p *Parser) parsePrimary() boogie.Expr {
	switch p.curr.Kind {
	case IDENT:
		name := p.curr.Value
		p.nextToken()
		return &boogie.VarExpr{V: boogie.Var{Name: name}}
	case INT_LIT:
		val, _ := strconv.Atoi(p.curr.Value)
		p.nextToken()
		return &boogie.IntLit{Value: val}
	case LPAREN:
		p.nextToken() // consume (
		expr := p.parseExpression(PREC_LOWEST)
		p.expect(RPAREN) // consume )
		return expr
	default:
		panic(fmt.Sprintf("unexpected token: %v", p.curr.Kind))
	}
}

func (p *Parser) parseStatements() []boogie.Stmt {
	var stmts []boogie.Stmt
	for p.curr.Kind != RBRACE && p.curr.Kind != EOF {
		switch p.curr.Kind {
		case VAR:
			stmts = append(stmts, p.parseVarDecl())
		case IDENT: // Likely an assignment: y := ...
			stmts = append(stmts, p.parseAssignment())
		case IF:
			stmts = append(stmts, p.parseIf())
		case RETURN:
			p.nextToken()
			expr := p.parseExpression(PREC_LOWEST)
			p.expect(SEMI)
			stmts = append(stmts, &boogie.Return{Values: []boogie.Expr{expr}})
		default:
			p.nextToken()
		}
	}
	return stmts
}

func (p *Parser) parseVarDecl() boogie.Stmt {
	p.expect(VAR)
	name := p.curr.Value
	p.expect(IDENT)
	p.expect(COLON)
	ty := p.parseType()
	p.expect(SEMI)

	return &boogie.LocalDecl{
		V: boogie.Var{Name: name, Ty: ty},
	}
}

func (p *Parser) parseIf() boogie.Stmt {
	p.expect(IF)

	// Parse condition, e.g., (z > 0)
	p.expect(LPAREN)
	cond := p.parseExpression(PREC_LOWEST)
	p.expect(RPAREN)

	// Parse 'then' block
	p.expect(LBRACE)
	thenBody := p.parseStatements()
	p.expect(RBRACE)

	var elseBody []boogie.Stmt
	// Check for optional 'else'
	if p.curr.Kind == ELSE {
		p.nextToken()
		// Handle 'else if' vs 'else { ... }'
		if p.curr.Kind == IF {
			elseBody = append(elseBody, p.parseIf())
		} else {
			p.expect(LBRACE)
			elseBody = p.parseStatements()
			p.expect(RBRACE)
		}
	}

	return &boogie.If{
		Cond: cond,
		Then: thenBody,
		Else: elseBody,
	}
}

func (p *Parser) parseAssignment() boogie.Stmt {
	lhsName := p.curr.Value
	p.expect(IDENT)
	p.expect(ASSIGN)
	rhs := p.parseExpression(PREC_LOWEST)
	p.expect(SEMI)

	return &boogie.Assign{
		Lhs: &boogie.VarExpr{V: boogie.Var{Name: lhsName}},
		Rhs: rhs,
	}
}
