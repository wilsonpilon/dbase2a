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
	Raw   string
	Start int
	End   int
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
		return Token{Type: TokenEOF, Value: "", Raw: "", Start: l.pos, End: l.pos}
	}

	start := l.pos
	ch := l.input[l.pos]

	if ch == '"' || ch == '\'' {
		return l.readString(ch)
	}

	if ch == ',' {
		l.pos++
		return Token{Type: TokenComma, Value: ",", Raw: ",", Start: start, End: l.pos}
	}

	if ch == ';' {
		l.pos++
		return Token{Type: TokenSemicolon, Value: ";", Raw: ";", Start: start, End: l.pos}
	}

	// Operadores de comparação com 2 caracteres
	if ch == '>' {
		l.pos++
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			return Token{Type: TokenIdentifier, Value: ">=", Raw: ">=", Start: start, End: l.pos}
		}
		return Token{Type: TokenIdentifier, Value: ">", Raw: ">", Start: start, End: l.pos}
	}

	if ch == '<' {
		l.pos++
		if l.pos < len(l.input) {
			if l.input[l.pos] == '=' {
				l.pos++
				return Token{Type: TokenIdentifier, Value: "<=", Raw: "<=", Start: start, End: l.pos}
			}
			if l.input[l.pos] == '>' {
				l.pos++
				return Token{Type: TokenIdentifier, Value: "<>", Raw: "<>", Start: start, End: l.pos}
			}
		}
		return Token{Type: TokenIdentifier, Value: "<", Raw: "<", Start: start, End: l.pos}
	}

	if ch == '!' {
		l.pos++
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			return Token{Type: TokenIdentifier, Value: "!=", Raw: "!=", Start: start, End: l.pos}
		}
		return Token{Type: TokenIdentifier, Value: "!", Raw: "!", Start: start, End: l.pos}
	}

	if ch == '=' {
		l.pos++
		return Token{Type: TokenEquals, Value: "=", Raw: "=", Start: start, End: l.pos}
	}

	if unicode.IsDigit(ch) {
		return l.readNumber()
	}

	if unicode.IsLetter(ch) || ch == '_' || ch == '.' || ch == '\\' || ch == '/' || ch == ':' {
		return l.readIdentifierOrKeyword()
	}

	l.pos++
	str := string(ch)
	return Token{Type: TokenIdentifier, Value: str, Raw: str, Start: start, End: l.pos}
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		l.pos++
	}
}

func (l *Lexer) readString(quote rune) Token {
	start := l.pos
	l.pos++ // pula abertura da aspa
	contentStart := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != quote {
		l.pos++
	}
	content := string(l.input[contentStart:l.pos])
	if l.pos < len(l.input) && l.input[l.pos] == quote {
		l.pos++ // pula fechamento da aspa
	}
	raw := string(l.input[start:l.pos])
	return Token{Type: TokenString, Value: content, Raw: raw, Start: start, End: l.pos}
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
	raw := string(l.input[start:l.pos])
	return Token{Type: TokenNumber, Value: raw, Raw: raw, Start: start, End: l.pos}
}

func (l *Lexer) readIdentifierOrKeyword() Token {
	start := l.pos
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' || ch == '.' || ch == '\\' || ch == '/' || ch == ':' || ch == '-' || ch == '$' {
			l.pos++
		} else {
			break
		}
	}
	raw := string(l.input[start:l.pos])
	return Token{Type: TokenIdentifier, Value: raw, Raw: raw, Start: start, End: l.pos}
}
