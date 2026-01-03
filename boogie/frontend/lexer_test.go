package frontend

import "testing"

func TestLexer(t *testing.T) {
	input := `procedure inc(x: int) returns (y: int) {
        var z: int;
        z := x + 1;
        if (z > 0) {
            y := z;
        } else {
            y := 0;
        }
        return y;
    }`

	input = `procedure inc(x: int) { z := x + 1; if (z <= 0) { return true; } }`

	expected := []struct {
		expectedKind  TokenKind
		expectedValue string
	}{
		{PROCEDURE, "procedure"},
		{IDENT, "inc"},
		{LPAREN, "("},
		{IDENT, "x"},
		{COLON, ":"},
		{IDENT, "int"},
		{RPAREN, ")"},
		{LBRACE, "{"},
		{IDENT, "z"},
		{ASSIGN, ":="},
		{IDENT, "x"},
		{PLUS, "+"},
		{INT_LIT, "1"},
		{SEMI, ";"},
		{IF, "if"},
		{LPAREN, "("},
		{IDENT, "z"},
		{LTE, "<="},
		{INT_LIT, "0"},
		{RPAREN, ")"},
		{LBRACE, "{"},
		{RETURN, "return"},
		{BOOL_LIT, "true"},
		{SEMI, ";"},
		{RBRACE, "}"},
		{RBRACE, "}"},
		{EOF, ""},
	}

	lexer := NewLexer(input)

	for i, tt := range expected {
		tok := lexer.NextToken()

		if tok.Kind != tt.expectedKind {
			t.Fatalf("tests[%d] - tokenkind wrong. expected=%d, got=%d",
				i, tt.expectedKind, tok.Kind)
		}

		if tok.Value != tt.expectedValue {
			t.Fatalf("tests[%d] - value wrong. expected=%q, got=%q",
				i, tt.expectedValue, tok.Value)
		}
	}
}
