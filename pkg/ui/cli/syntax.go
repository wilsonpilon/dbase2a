// Package cli implementa o realce de sintaxe e estilização de terminal para o dBase II++.
package cli

import (
	"fmt"
	"regexp"
	"strings"
	"unicode"
)

// Códigos ANSI para colorização de sintaxe no terminal
const (
	AnsiReset        = "\x1b[0m"
	AnsiBold         = "\x1b[1m"
	AnsiDim          = "\x1b[2m"
	AnsiRed          = "\x1b[1;31m" // Erros
	AnsiGreen        = "\x1b[1;32m" // Strings e Nomes de Arquivos
	AnsiYellow       = "\x1b[1;33m" // Operadores lógicos e aritméticos
	AnsiBlue         = "\x1b[1;34m" // Números
	AnsiMagenta      = "\x1b[1;35m" // Funções nativas (@, TRIM, SUBSTR, RECNO...)
	AnsiCyan         = "\x1b[1;36m" // Comandos e Verbos dBase (USE, CREATE, LIST...)
	AnsiWhite        = "\x1b[1;37m" // Palavras-chave de cláusulas (FOR, WHILE, WITH, TO...)
	AnsiGray         = "\x1b[90m"   // Comentários (*, NOTE)
	AnsiPromptYellow = "\x1b[1;33m" // Prompt '.'
)

var (
	// Verbos / Comandos principais do dBASE II
	dBaseCommands = map[string]bool{
		"APPEND":   true,
		"BROWSE":   true,
		"CLEAR":    true,
		"CLS":      true,
		"CONTINUE": true,
		"COPY":     true,
		"COUNT":    true,
		"CREATE":   true,
		"DELETE":   true,
		"DISPLAY":  true,
		"DO":       true,
		"EDIT":     true,
		"FIND":     true,
		"GO":       true,
		"GOTO":     true,
		"HELP":     true,
		"INDEX":    true,
		"LIST":     true,
		"LOCATE":   true,
		"MODIFY":   true,
		"PACK":     true,
		"QUIT":     true,
		"EXIT":     true,
		"RECALL":   true,
		"REPLACE":  true,
		"RESTORE":  true,
		"SAVE":     true,
		"SELECT":   true,
		"SET":      true,
		"SKIP":     true,
		"SORT":     true,
		"STORE":    true,
		"SUM":      true,
		"TOTAL":    true,
		"USE":      true,
		"ZAP":      true,
		"?":        true,
		"??":       true,
		"@":        true,
	}

	// Palavras-chave de cláusulas e argumentos
	dBaseKeywords = map[string]bool{
		"FOR":       true,
		"WHILE":     true,
		"WITH":      true,
		"TO":        true,
		"ALL":       true,
		"STRUCTURE": true,
		"BLANK":     true,
		"ON":        true,
		"OFF":       true,
		"FIELDS":    true,
		"STEP":      true,
		"TOP":       true,
		"BOTTOM":    true,
		"COMMAND":   true,
		"COMMANDS":  true,
		"COMANDOS":  true,
		"RECORD":    true,
		"TALK":      true,
		"BELL":      true,
		"ECHO":      true,
		"EXACT":     true,
		"HEADING":   true,
		"INTENSITY": true,
		"SAFETY":    true,
		"STATUS":    true,
		"DEFAULT":   true,
		"DEVICE":    true,
		"SCREEN":    true,
		"PRINT":     true,
		"SAY":       true,
		"GET":       true,
		"READ":      true,
	}

	// Funções intrínsecas do dBASE II++
	dBaseFunctions = map[string]bool{
		"TRIM":    true,
		"RTRIM":   true,
		"LTRIM":   true,
		"ALLTRIM": true,
		"UPPER":   true,
		"LOWER":   true,
		"LEN":     true,
		"SUBSTR":  true,
		"VAL":     true,
		"STR":     true,
		"INT":     true,
		"ABS":     true,
		"DELETED": true,
		"RECNO":   true,
		"EOF":     true,
		"DATE":    true,
		"TIME":    true,
		"IIF":     true,
		"CHR":     true,
		"RANK":    true,
		"TYPE":    true,
		"TEST":    true,
		"FILE":    true,
	}

	// Operadores lógicos do dBase (.AND., .OR., .NOT., .T., .F.)
	dBaseLogicalOps = map[string]bool{
		".AND.": true,
		".OR.":  true,
		".NOT.": true,
		".T.":   true,
		".F.":   true,
		".Y.":   true,
		".N.":   true,
	}

	// Regex para identificar arquivos comuns (.DBF, .CMD, .PRG, .NDX, .TXT, .MEM, .COM, .DAT)
	fileRegex = regexp.MustCompile(`(?i)^[A-Z0-9_\-\\]+\.(DBF|CMD|PRG|NDX|TXT|MEM|COM|DAT)$`)
)

// HighlightLine aplica syntax highlighting sobre uma linha de comando dBase II++.
func HighlightLine(line string) string {
	trimmed := strings.TrimSpace(line)
	if trimmed == "" {
		return line
	}

	// Comentários: linha inteira em cinza
	if strings.HasPrefix(trimmed, "*") || strings.HasPrefix(strings.ToUpper(trimmed), "NOTE ") {
		return AnsiGray + line + AnsiReset
	}

	var out strings.Builder
	runes := []rune(line)
	n := len(runes)
	i := 0
	wordIndex := 0

	for i < n {
		r := runes[i]

		// 1. Espaços em branco
		if unicode.IsSpace(r) {
			out.WriteRune(r)
			i++
			continue
		}

		// 2. Strings delimitadas por aspas simples ou duplas
		if r == '"' || r == '\'' {
			quote := r
			start := i
			i++
			for i < n && runes[i] != quote {
				i++
			}
			if i < n && runes[i] == quote {
				i++
			}
			out.WriteString(AnsiGreen)
			out.WriteString(string(runes[start:i]))
			out.WriteString(AnsiReset)
			wordIndex++
			continue
		}

		// 3. Operadores lógicos com pontos como .AND., .OR., .NOT., .T., .F.
		if r == '.' && i+1 < n && unicode.IsLetter(runes[i+1]) {
			start := i
			i++
			for i < n && (unicode.IsLetter(runes[i]) || runes[i] == '.') {
				if runes[i] == '.' {
					i++
					break
				}
				i++
			}
			token := string(runes[start:i])
			if dBaseLogicalOps[strings.ToUpper(token)] {
				out.WriteString(AnsiYellow)
				out.WriteString(token)
				out.WriteString(AnsiReset)
			} else {
				out.WriteString(token)
			}
			wordIndex++
			continue
		}

		// 4. Operadores e pontuações especiais
		if r == '=' || r == '<' || r == '>' || r == '+' || r == '-' || r == '*' || r == '/' || r == '$' || r == '#' || r == '!' {
			// Verifica operadores compostos como <=, >=, <>, !=
			start := i
			i++
			if i < n && (runes[i] == '=' || runes[i] == '>') {
				i++
			}
			out.WriteString(AnsiYellow)
			out.WriteString(string(runes[start:i]))
			out.WriteString(AnsiReset)
			continue
		}

		if r == ',' || r == ';' || r == '(' || r == ')' {
			out.WriteRune(r)
			i++
			continue
		}

		if r == '?' {
			start := i
			i++
			if i < n && runes[i] == '?' {
				i++
			}
			out.WriteString(AnsiCyan)
			out.WriteString(string(runes[start:i]))
			out.WriteString(AnsiReset)
			wordIndex++
			continue
		}

		if r == '@' {
			out.WriteString(AnsiCyan + "@" + AnsiReset)
			i++
			wordIndex++
			continue
		}

		// 5. Números literais
		if unicode.IsDigit(r) {
			start := i
			hasDot := false
			for i < n && (unicode.IsDigit(runes[i]) || (runes[i] == '.' && !hasDot)) {
				if runes[i] == '.' {
					hasDot = true
				}
				i++
			}
			out.WriteString(AnsiBlue)
			out.WriteString(string(runes[start:i]))
			out.WriteString(AnsiReset)
			wordIndex++
			continue
		}

		// 6. Palavras (identificadores, comandos, funções, arquivos, campos)
		start := i
		for i < n && (unicode.IsLetter(runes[i]) || unicode.IsDigit(runes[i]) || runes[i] == '_' || runes[i] == '.' || runes[i] == '\\' || runes[i] == '/' || runes[i] == ':') {
			i++
		}
		word := string(runes[start:i])
		upperWord := strings.ToUpper(word)

		// Verifica tipo da palavra
		if wordIndex == 0 && dBaseCommands[upperWord] {
			// Primeiro termo sendo um comando dBase
			out.WriteString(AnsiCyan)
			out.WriteString(word)
			out.WriteString(AnsiReset)
		} else if dBaseCommands[upperWord] {
			// Comando dBase em outra posição (ex: após ponto e vírgula)
			out.WriteString(AnsiCyan)
			out.WriteString(word)
			out.WriteString(AnsiReset)
		} else if dBaseKeywords[upperWord] {
			// Cláusulas e argumentos (FOR, WHILE, WITH, TO...)
			out.WriteString(AnsiWhite)
			out.WriteString(word)
			out.WriteString(AnsiReset)
		} else if dBaseFunctions[upperWord] {
			// Funções nativas (TRIM, SUBSTR, RECNO...)
			out.WriteString(AnsiMagenta)
			out.WriteString(word)
			out.WriteString(AnsiReset)
		} else if fileRegex.MatchString(word) {
			// Nome ou caminho de arquivo com extensão
			out.WriteString(AnsiGreen)
			out.WriteString(word)
			out.WriteString(AnsiReset)
		} else {
			// Identificador comum / nome de campo / variável
			out.WriteString(word)
		}

		wordIndex++
	}

	return out.String()
}

// FormatError colore mensagens de erro em vermelho brilhante.
func FormatError(errMsg string) string {
	return fmt.Sprintf("%s*** ERRO: %s%s", AnsiRed, errMsg, AnsiReset)
}

// FormatPrompt colore o prompt interativo.
func FormatPrompt(prompt string) string {
	return fmt.Sprintf("%s%s%s", AnsiPromptYellow, prompt, AnsiReset)
}
