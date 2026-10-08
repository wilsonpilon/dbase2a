// Package cli implementa o editor de linha com histórico, edição interativa,
// autocomplete por TAB cíclico e realce de sintaxe em tempo real.
package cli

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"unicode/utf8"

	"golang.org/x/term"
)

// LineEditor provê a leitura de linhas interativa no terminal.
type LineEditor struct {
	In           io.Reader
	Out          io.Writer
	History      []string
	HistoryIndex int
	AutoState    AutocompleteState
	SavedBuffer  string
	bufReader    *bufio.Reader
}

// NewLineEditor instancia um novo editor de linha interativo.
func NewLineEditor(in io.Reader, out io.Writer) *LineEditor {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	return &LineEditor{
		In:        in,
		Out:       out,
		History:   make([]string, 0),
		bufReader: bufio.NewReader(in),
	}
}

// ReadLine lê uma linha interativa formatada com syntax highlighting e autocomplete.
func (le *LineEditor) ReadLine(prompt string) (string, error) {
	inFile, isInFile := le.In.(*os.File)
	outFile, isOutFile := le.Out.(*os.File)

	// Fallback para leitura não-interativa (pipes, scripts, testes automatizados)
	if !isInFile || !isOutFile || !term.IsTerminal(int(inFile.Fd())) || !term.IsTerminal(int(outFile.Fd())) {
		fmt.Fprintf(le.Out, "%s ", prompt)
		if le.bufReader == nil {
			le.bufReader = bufio.NewReader(le.In)
		}
		line, err := le.bufReader.ReadString('\n')
		if err != nil && len(line) == 0 {
			return "", err
		}
		return strings.TrimRight(line, "\r\n"), nil
	}

	// Ativa modo Raw no terminal para controle total da linha
	oldState, err := term.MakeRaw(int(inFile.Fd()))
	if err != nil {
		// Fallback para leitura padrão
		fmt.Fprintf(le.Out, "%s ", prompt)
		if le.bufReader == nil {
			le.bufReader = bufio.NewReader(le.In)
		}
		line, err := le.bufReader.ReadString('\n')
		if err != nil && len(line) == 0 {
			return "", err
		}
		return strings.TrimRight(line, "\r\n"), nil
	}
	defer func() {
		_ = term.Restore(int(inFile.Fd()), oldState)
	}()

	var buffer []rune
	cursorPos := 0
	le.HistoryIndex = len(le.History)
	le.SavedBuffer = ""
	le.AutoState.Reset()

	promptLen := utf8.RuneCountInString(prompt) + 1

	redraw := func() {
		// Retorna ao início da linha e limpa até o fim
		fmt.Fprint(le.Out, "\r\x1b[K")

		// Imprime prompt colorido
		fmt.Fprintf(le.Out, "%s ", FormatPrompt(prompt))

		// Imprime conteúdo com realce de sintaxe
		highlighted := HighlightLine(string(buffer))
		fmt.Fprint(le.Out, highlighted)

		// Reposiciona o cursor no ponto correto
		col := promptLen + cursorPos + 1
		fmt.Fprintf(le.Out, "\r\x1b[%dC", col-1)
	}

	redraw()

	readBuf := make([]byte, 16)
	for {
		n, err := inFile.Read(readBuf)
		if err != nil || n == 0 {
			return "", err
		}

		input := readBuf[:n]

		// Tratamento de sequências ANSI de escape
		if bytes.HasPrefix(input, []byte("\x1b")) {
			if len(input) == 1 {
				// Tecla ESC isolada: limpa autocomplete
				le.AutoState.Reset()
				continue
			}

			// Seta para CIMA: histórico anterior
			if bytes.Equal(input, []byte("\x1b[A")) || bytes.Equal(input, []byte("\x1bOA")) {
				le.AutoState.Reset()
				if len(le.History) > 0 && le.HistoryIndex > 0 {
					if le.HistoryIndex == len(le.History) {
						le.SavedBuffer = string(buffer)
					}
					le.HistoryIndex--
					buffer = []rune(le.History[le.HistoryIndex])
					cursorPos = len(buffer)
					redraw()
				}
				continue
			}

			// Seta para BAIXO: histórico posterior
			if bytes.Equal(input, []byte("\x1b[B")) || bytes.Equal(input, []byte("\x1bOB")) {
				le.AutoState.Reset()
				if le.HistoryIndex < len(le.History) {
					le.HistoryIndex++
					if le.HistoryIndex == len(le.History) {
						buffer = []rune(le.SavedBuffer)
					} else {
						buffer = []rune(le.History[le.HistoryIndex])
					}
					cursorPos = len(buffer)
					redraw()
				}
				continue
			}

			// Seta para ESQUERDA
			if bytes.Equal(input, []byte("\x1b[D")) || bytes.Equal(input, []byte("\x1bOD")) {
				le.AutoState.Reset()
				if cursorPos > 0 {
					cursorPos--
					redraw()
				}
				continue
			}

			// Seta para DIREITA
			if bytes.Equal(input, []byte("\x1b[C")) || bytes.Equal(input, []byte("\x1bOC")) {
				le.AutoState.Reset()
				if cursorPos < len(buffer) {
					cursorPos++
					redraw()
				}
				continue
			}

			// HOME
			if bytes.Equal(input, []byte("\x1b[H")) || bytes.Equal(input, []byte("\x1b[1~")) {
				le.AutoState.Reset()
				cursorPos = 0
				redraw()
				continue
			}

			// END
			if bytes.Equal(input, []byte("\x1b[F")) || bytes.Equal(input, []byte("\x1b[4~")) {
				le.AutoState.Reset()
				cursorPos = len(buffer)
				redraw()
				continue
			}

			// DELETE
			if bytes.Equal(input, []byte("\x1b[3~")) {
				le.AutoState.Reset()
				if cursorPos < len(buffer) {
					buffer = append(buffer[:cursorPos], buffer[cursorPos+1:]...)
					redraw()
				}
				continue
			}

			continue
		}

		// Tecla única
		b := input[0]

		switch b {
		case '\r', '\n': // Enter: submeter linha
			fmt.Fprint(le.Out, "\r\n")
			res := string(buffer)
			trimmed := strings.TrimSpace(res)
			if trimmed != "" {
				// Adiciona ao histórico se não for repetição imediata
				if len(le.History) == 0 || le.History[len(le.History)-1] != trimmed {
					le.History = append(le.History, trimmed)
				}
			}
			return res, nil

		case '\t': // TAB: Autocomplete cíclico
			newLine, newPos, ok := le.AutoState.NextCompletion(string(buffer), cursorPos)
			if ok {
				buffer = []rune(newLine)
				cursorPos = newPos
				redraw()
			}

		case 127, 8: // Backspace
			le.AutoState.Reset()
			if cursorPos > 0 {
				buffer = append(buffer[:cursorPos-1], buffer[cursorPos:]...)
				cursorPos--
				redraw()
			}

		case 3: // Ctrl+C: cancela linha atual
			le.AutoState.Reset()
			fmt.Fprint(le.Out, "^C\r\n")
			return "", nil

		case 4: // Ctrl+D: EOF / QUIT se buffer estiver vazio
			if len(buffer) == 0 {
				fmt.Fprint(le.Out, "QUIT\r\n")
				return "QUIT", nil
			}
			// Se não estiver vazio, age como Delete
			le.AutoState.Reset()
			if cursorPos < len(buffer) {
				buffer = append(buffer[:cursorPos], buffer[cursorPos+1:]...)
				redraw()
			}

		case 1: // Ctrl+A: início da linha
			le.AutoState.Reset()
			cursorPos = 0
			redraw()

		case 5: // Ctrl+E: fim da linha
			le.AutoState.Reset()
			cursorPos = len(buffer)
			redraw()

		case 11: // Ctrl+K: apaga até o fim da linha
			le.AutoState.Reset()
			buffer = buffer[:cursorPos]
			redraw()

		case 21: // Ctrl+U: limpa a linha inteira
			le.AutoState.Reset()
			buffer = nil
			cursorPos = 0
			redraw()

		case 12: // Ctrl+L: limpa a tela do terminal
			fmt.Fprint(le.Out, "\x1b[H\x1b[2J")
			redraw()

		default:
			// Insere caractere digitado (tratando caracteres UTF-8)
			le.AutoState.Reset()
			r, _ := utf8.DecodeRune(input)
			if r != utf8.RuneError && r >= 32 {
				if cursorPos == len(buffer) {
					buffer = append(buffer, r)
				} else {
					buffer = append(buffer[:cursorPos], append([]rune{r}, buffer[cursorPos:]...)...)
				}
				cursorPos++
				redraw()
			}
		}
	}
}
