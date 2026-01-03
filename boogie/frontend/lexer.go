package frontend

import "unicode"

type TokenKind int

type Lexer struct {
	src []rune
	pos int
}

func NewLexer(input string) *Lexer {
	return &Lexer{src: []rune(input), pos: 0}
}

// peek returns the current character without advancing
func (l *Lexer) peek() rune {
	if l.pos >= len(l.src) {
		return 0
	}
	return l.src[l.pos]
}

// advance consumes the current character and returns it
func (l *Lexer) advance() rune {
	ch := l.peek()
	l.pos++
	return ch
}

// skipWhitespace ignores spaces, tabs, and newlines
func (l *Lexer) skipWhitespace() {
	for unicode.IsSpace(l.peek()) {
		l.advance()
	}
}

const (
	EOF TokenKind = iota

	// identifiers + literals
	IDENT
	INT_LIT
	BOOL_LIT

	// keywords
	PROCEDURE
	RETURNS
	VAR
	IF
	ELSE
	WHILE
	RETURN

	// symbols
	LPAREN
	RPAREN
	LBRACE
	RBRACE
	COLON
	COMMA
	SEMI
	ASSIGN // :=
	PLUS
	MINUS
	MUL
	EQ
	LT
	GT
	GTE
	LTE
	AND
	OR
	NOT
)

type Token struct {
	Kind  TokenKind
	Value string
}

func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	ch := l.peek()
	if ch == 0 {
		return Token{Kind: EOF, Value: ""}
	}

	if ch == '/' {
		if l.peek() == '/' { // You'd need a peekNext helper
			l.skipLineComment()
			return l.NextToken()
		}
	}

	// Handle Identifiers and Keywords
	if unicode.IsLetter(ch) || ch == '_' || ch == '$' || ch == '\'' {
		return l.lexIdentifier()
	}

	// Handle Numbers
	if unicode.IsDigit(ch) {
		return l.lexNumber()
	}

	// Handle Multi-character and Single-character Symbols
	l.advance()
	switch ch {
	case '(':
		return Token{LPAREN, "("}
	case ')':
		return Token{RPAREN, ")"}
	case '{':
		return Token{LBRACE, "{"}
	case '}':
		return Token{RBRACE, "}"}
	case ',':
		return Token{COMMA, ","}
	case ';':
		return Token{SEMI, ";"}
	case '+':
		return Token{PLUS, "+"}
	case '-':
		return Token{MINUS, "-"}
	case '*':
		return Token{MUL, "*"}
	case ':':
		if l.peek() == '=' {
			l.advance()
			return Token{ASSIGN, ":="}
		}
		return Token{COLON, ":"}
	case '<':
		if l.peek() == '=' {
			l.advance()
			return Token{LTE, "<="}
		}
		return Token{LT, "<"}
	case '>':
		if l.peek() == '=' {
			l.advance()
			return Token{GTE, ">="}
		}
		return Token{GT, ">"}
	case '=':
		return Token{EQ, "="}
	case '&':
		if l.peek() == '&' {
			l.advance()
			return Token{AND, "&&"}
		}
	case '|':
		if l.peek() == '|' {
			l.advance()
			return Token{OR, "||"}
		}
	case '!':
		return Token{NOT, "!"}
	}

	return Token{EOF, ""} // Or handle as ILLEGAL token
}

func (l *Lexer) skipLineComment() {
	for l.peek() != '\n' && l.peek() != 0 {
		l.advance()
	}
}

var keywords = map[string]TokenKind{
	"procedure": PROCEDURE,
	"returns":   RETURNS,
	"var":       VAR,
	"if":        IF,
	"else":      ELSE,
	"while":     WHILE,
	"return":    RETURN,
	"true":      BOOL_LIT,
	"false":     BOOL_LIT,
}

func (l *Lexer) lexIdentifier() Token {
	start := l.pos
	for isIdentChar(l.peek()) {
		l.advance()
	}
	val := string(l.src[start:l.pos])
	if kind, ok := keywords[val]; ok {
		return Token{kind, val}
	}
	return Token{IDENT, val}
}

func isIdentChar(ch rune) bool {
	return unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '$' || ch == '\'' || ch == '.'
}

func (l *Lexer) lexNumber() Token {
	start := l.pos
	for unicode.IsDigit(l.peek()) {
		l.advance()
	}
	return Token{INT_LIT, string(l.src[start:l.pos])}
}
