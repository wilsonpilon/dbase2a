package eval

import (
	"strings"
	"unicode"
)

// TokenType define o tipo de token da expressão.
type TokenType int

const (
	TokEOF TokenType = iota
	TokIdentifier
	TokString
	TokNumber
	TokLogical
	TokPlus
	TokMinus
	TokStar
	TokSlash
	TokEqual
	TokNotEqual
	TokLess
	TokGreater
	TokLessEqual
	TokGreaterEqual
	TokSubstring // $
	TokAnd       // .AND.
	TokOr        // .OR.
	TokNot       // .NOT. ou !
	TokLParen    // (
	TokRParen    // )
	TokComma     // ,
)

// Token representa uma unidade léxica na expressão.
type Token struct {
	Type  TokenType
	Value string
}

// Lexer divide a expressão matemática/lógica em tokens.
type Lexer struct {
	input []rune
	pos   int
}

// NewLexer inicializa o analisador léxico de expressões.
func NewLexer(input string) *Lexer {
	return &Lexer{
		input: []rune(input),
		pos:   0,
	}
}

// NextToken retorna o próximo token da expressão.
func (l *Lexer) NextToken() Token {
	l.skipWhitespace()

	if l.pos >= len(l.input) {
		return Token{Type: TokEOF, Value: ""}
	}

	ch := l.input[l.pos]

	// Strings com aspas simples, duplas ou colchetes [string]
	if ch == '"' || ch == '\'' {
		return l.readString(ch)
	}

	// Parênteses e vírgula
	if ch == '(' {
		l.pos++
		return Token{Type: TokLParen, Value: "("}
	}
	if ch == ')' {
		l.pos++
		return Token{Type: TokRParen, Value: ")"}
	}
	if ch == ',' {
		l.pos++
		return Token{Type: TokComma, Value: ","}
	}

	// Operadores aritméticos
	if ch == '+' {
		l.pos++
		return Token{Type: TokPlus, Value: "+"}
	}
	if ch == '-' {
		l.pos++
		return Token{Type: TokMinus, Value: "-"}
	}
	if ch == '*' {
		l.pos++
		return Token{Type: TokStar, Value: "*"}
	}
	if ch == '/' {
		l.pos++
		return Token{Type: TokSlash, Value: "/"}
	}
	if ch == '$' {
		l.pos++
		return Token{Type: TokSubstring, Value: "$"}
	}
	if ch == '!' {
		l.pos++
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			return Token{Type: TokNotEqual, Value: "!="}
		}
		return Token{Type: TokNot, Value: "!"}
	}
	if ch == '#' {
		l.pos++
		return Token{Type: TokNotEqual, Value: "#"}
	}

	// Operadores relacionais (=, <, >, <=, >=, <>)
	if ch == '=' {
		l.pos++
		return Token{Type: TokEqual, Value: "="}
	}
	if ch == '<' {
		l.pos++
		if l.pos < len(l.input) {
			if l.input[l.pos] == '>' {
				l.pos++
				return Token{Type: TokNotEqual, Value: "<>"}
			}
			if l.input[l.pos] == '=' {
				l.pos++
				return Token{Type: TokLessEqual, Value: "<="}
			}
		}
		return Token{Type: TokLess, Value: "<"}
	}
	if ch == '>' {
		l.pos++
		if l.pos < len(l.input) && l.input[l.pos] == '=' {
			l.pos++
			return Token{Type: TokGreaterEqual, Value: ">="}
		}
		return Token{Type: TokGreater, Value: ">"}
	}

	// Literais lógicos ou operadores delimitados por ponto (.AND., .OR., .NOT., .T., .F.)
	if ch == '.' {
		// Checa se é número decimal (.5) ou token com ponto (.AND.)
		if l.pos+1 < len(l.input) && unicode.IsDigit(l.input[l.pos+1]) {
			return l.readNumber()
		}
		return l.readDottedToken()
	}

	// Números
	if unicode.IsDigit(ch) {
		return l.readNumber()
	}

	// Identificadores, palavras-chave e nomes de funções
	if unicode.IsLetter(ch) || ch == '_' {
		return l.readIdentifier()
	}

	l.pos++
	return Token{Type: TokIdentifier, Value: string(ch)}
}

func (l *Lexer) skipWhitespace() {
	for l.pos < len(l.input) && unicode.IsSpace(l.input[l.pos]) {
		l.pos++
	}
}

func (l *Lexer) readString(quote rune) Token {
	l.pos++ // pula abertura
	start := l.pos
	for l.pos < len(l.input) && l.input[l.pos] != quote {
		l.pos++
	}
	val := string(l.input[start:l.pos])
	if l.pos < len(l.input) && l.input[l.pos] == quote {
		l.pos++ // pula fechamento
	}
	return Token{Type: TokString, Value: val}
}

func (l *Lexer) readNumber() Token {
	start := l.pos
	hasDot := false
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if unicode.IsDigit(ch) {
			l.pos++
		} else if ch == '.' && !hasDot {
			// Só considera ponto como decimal se próximo caractere for dígito ou fim
			if l.pos+1 < len(l.input) && unicode.IsLetter(l.input[l.pos+1]) {
				break
			}
			hasDot = true
			l.pos++
		} else {
			break
		}
	}
	return Token{Type: TokNumber, Value: string(l.input[start:l.pos])}
}

func (l *Lexer) readDottedToken() Token {
	start := l.pos
	l.pos++ // consome o primeiro '.'
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if ch == '.' {
			l.pos++ // consome o fechamento '.'
			break
		}
		if !unicode.IsLetter(ch) {
			break
		}
		l.pos++
	}
	raw := string(l.input[start:l.pos])
	upper := strings.ToUpper(raw)

	switch upper {
	case ".AND.":
		return Token{Type: TokAnd, Value: ".AND."}
	case ".OR.":
		return Token{Type: TokOr, Value: ".OR."}
	case ".NOT.":
		return Token{Type: TokNot, Value: ".NOT."}
	case ".T.", ".Y.", ".TRUE.":
		return Token{Type: TokLogical, Value: ".T."}
	case ".F.", ".N.", ".FALSE.":
		return Token{Type: TokLogical, Value: ".F."}
	default:
		return Token{Type: TokIdentifier, Value: raw}
	}
}

func (l *Lexer) readIdentifier() Token {
	start := l.pos
	for l.pos < len(l.input) {
		ch := l.input[l.pos]
		if unicode.IsLetter(ch) || unicode.IsDigit(ch) || ch == '_' {
			l.pos++
		} else {
			break
		}
	}
	val := string(l.input[start:l.pos])
	upper := strings.ToUpper(val)

	// Palavras-chave lógicas sem ponto (opcional no dBASE II/III estendido)
	switch upper {
	case "AND":
		return Token{Type: TokAnd, Value: "AND"}
	case "OR":
		return Token{Type: TokOr, Value: "OR"}
	case "NOT":
		return Token{Type: TokNot, Value: "NOT"}
	default:
		return Token{Type: TokIdentifier, Value: val}
	}
}
