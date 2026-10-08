package parser

import (
	"unicode"
)

// TokenType representa a classe léxica do token.
type TokenType int

const (
	TokenEOF TokenType = iota
	TokenIdentifier
	TokenKeyword
	TokenString
	TokenNumber
	TokenComma
	TokenSemicolon
	TokenEquals
	TokenDot
)

// Token representa uma unidade léxica com sua posição e valor textual.
type Token struct {
	Type  TokenType
	Value string
}

// Lexer divide a linha de comando em uma sequência de tokens.
type Lexer struct {
	input []rune
	pos   int
}

// NewLexer cria uma nova instância de analisador léxico.
func NewLexer(input string) *Lexer {
	return &Lexer{
		input: []rune(input),
		pos:   0,
	}
}

// NextToken consome e retorna o próximo token.
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return Token{Type: TokenEOF, Value: ""}
	}

	ch := l.input[l.pos]

	if ch == '"' || ch == '\'' {
		return l.readString(ch)
	}

	if ch == ',' {
		l.pos++
		return Token{Type: TokenComma, Value: ","}
	}

	if ch == ';' {
		l.pos++
		return Token{Type: TokenSemicolon, Value: ";"}
	}

	if ch == '=' {
		l.pos++
		return Token{Type: TokenEquals, Value: "="}
	}

	if unicode.IsDigit(ch) {
		return l.readNumber()
	}

	if unicode.IsLetter(ch) || ch == '_' || ch == '.' || ch == '\\' || ch == '/' || ch == ':' {
		return l.readIdentifierOrKeyword()
	}

	l.pos++
	return Token{Type: TokenIdentifier, Value: string(ch)}
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		l.pos++
	}
}

func (l *Lexer) readString(quote rune) Token {
	l.pos++ // pula abertura da aspa
	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != quote {
		l.pos++
	}
	str := string(l.input[start:l.pos])
	if l.pos < len(l.input) && l.input[l.pos] == quote {
		l.pos++ // pula fechamento da aspa
	}
	return Token{Type: TokenString, Value: str}
}

func (l *Lexer) readNumber() Token {
	start := l.pos
	hasDot := false
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if unicode.IsDigit(ch) {
			l.pos++
		} else if ch == '.' && !hasDot {
			hasDot = true
			l.pos++
		} else {
			break
		}
	}
	return Token{Type: TokenNumber, Value: string(l.input[start:l.pos])}
}

func (l *Lexer) readIdentifierOrKeyword() Token {
	start := l.pos
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '.' || ch == '\\' || ch == '/' || ch == ':' || ch == '-' {
			l.pos++
		} else {
			break
		}
	}
	val := string(l.input[start:l.pos])
	return Token{Type: TokenIdentifier, Value: val}
}
