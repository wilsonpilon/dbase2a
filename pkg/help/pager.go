// Package help implementa o paginador interativo para visualização de textos longos de ajuda.
package help

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"

	"golang.org/x/term"
)

// Pager exibe um texto com rolagem de telas/linhas em terminais interativos.
type Pager struct {
	In           io.Reader
	Out          io.Writer
	LinesPerPage int
	DisableRaw   bool
}

// NewPager instancia um novo paginador para a entrada e saída informadas.
func NewPager(in io.Reader, out io.Writer) *Pager {
	if in == nil {
		in = os.Stdin
	}
	if out == nil {
		out = os.Stdout
	}
	return &Pager{
		In:  in,
		Out: out,
	}
}

// IsTerminalOut verifica se o descritor de saída é um terminal TTY.
func IsTerminalOut(w io.Writer) bool {
	if f, ok := w.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// IsTerminalIn verifica se o descritor de entrada é um terminal TTY.
func IsTerminalIn(r io.Reader) bool {
	if f, ok := r.(*os.File); ok {
		return term.IsTerminal(int(f.Fd()))
	}
	return false
}

// Display exibe o texto no terminal. Se for um terminal interativo e o texto exceder a altura da tela,
// ativa o modo interativo com rolagem de páginas e linhas.
func (p *Pager) Display(text string) error {
	lines := strings.Split(text, "\n")
	totalLines := len(lines)

	// Se a saída não for um terminal ou entrada não for interativa, imprime direto
	if !IsTerminalOut(p.Out) || !IsTerminalIn(p.In) || p.DisableRaw {
		for _, l := range lines {
			fmt.Fprintln(p.Out, l)
		}
		return nil
	}

	inFile, _ := p.In.(*os.File)
	outFile, _ := p.Out.(*os.File)

	termWidth, termHeight, err := term.GetSize(int(outFile.Fd()))
	if err != nil || termHeight <= 0 {
		termHeight = 24
		termWidth = 80
	}

	screenHeight := termHeight - 2
	if screenHeight < 5 {
		screenHeight = 5
	}

	// Se o texto couber na tela inteira, imprime sem necessidade de paginação
	if totalLines <= screenHeight+1 {
		for _, l := range lines {
			fmt.Fprintln(p.Out, l)
		}
		return nil
	}

	// Ativa modo Raw no terminal para captura instantânea de teclas
	oldState, err := term.MakeRaw(int(inFile.Fd()))
	if err != nil {
		// Fallback para impressão direta se falhar o modo raw
		for _, l := range lines {
			fmt.Fprintln(p.Out, l)
		}
		return nil
	}
	defer func() {
		_ = term.Restore(int(inFile.Fd()), oldState)
		// Restaura cursor e adiciona quebra de linha
		fmt.Fprintf(p.Out, "\x1b[?25h\r\n")
	}()

	// Oculta cursor durante paginação
	fmt.Fprint(p.Out, "\x1b[?25l")

	topLine := 0
	maxTopLine := totalLines - screenHeight
	if maxTopLine < 0 {
		maxTopLine = 0
	}

	renderPage := func() {
		// Limpa a tela e posiciona no topo
		fmt.Fprint(p.Out, "\x1b[H\x1b[2J")

		endLine := topLine + screenHeight
		if endLine > totalLines {
			endLine = totalLines
		}

		for i := topLine; i < endLine; i++ {
			lineText := lines[i]
			// Limita a largura para evitar quebra automática feia se a linha for muito longa
			if len(lineText) > termWidth && termWidth > 10 {
				lineText = lineText[:termWidth]
			}
			fmt.Fprintf(p.Out, "%s\r\n", lineText)
		}

		// Preenche eventuais linhas vazias
		for i := endLine - topLine; i < screenHeight; i++ {
			fmt.Fprint(p.Out, "\r\n")
		}

		// Barra de status com destaque
		percent := int(float64(endLine) / float64(totalLines) * 100.0)
		status := fmt.Sprintf(" -- Linhas %d-%d de %d (%d%%) | [Espaço/PageDown] Avançar tela, [b/PageUp] Voltar tela, [Setas/j/k] Rolar, [q/Esc] Sair --",
			topLine+1, endLine, totalLines, percent)
		if len(status) > termWidth && termWidth > 10 {
			status = status[:termWidth]
		}

		// Barra em vídeo reverso (fundo branco / texto preto)
		fmt.Fprintf(p.Out, "\x1b[7m%-*s\x1b[0m", termWidth, status)
	}

	renderPage()

	buf := make([]byte, 16)
	for {
		n, err := inFile.Read(buf)
		if err != nil || n == 0 {
			break
		}

		input := buf[:n]

		// Trata sequências de escape (setas, page up, page down, home, end)
		if bytes.HasPrefix(input, []byte("\x1b")) {
			if len(input) == 1 {
				// Tecla ESC sozinha: sair
				break
			}
			if bytes.Equal(input, []byte("\x1b[A")) || bytes.Equal(input, []byte("\x1bOA")) {
				// Seta para CIMA
				if topLine > 0 {
					topLine--
					renderPage()
				}
				continue
			}
			if bytes.Equal(input, []byte("\x1b[B")) || bytes.Equal(input, []byte("\x1bOB")) {
				// Seta para BAIXO
				if topLine < maxTopLine {
					topLine++
					renderPage()
				}
				continue
			}
			if bytes.Equal(input, []byte("\x1b[5~")) || bytes.Equal(input, []byte("\x1b[V")) {
				// Page UP
				topLine -= screenHeight
				if topLine < 0 {
					topLine = 0
				}
				renderPage()
				continue
			}
			if bytes.Equal(input, []byte("\x1b[6~")) || bytes.Equal(input, []byte("\x1b[U")) {
				// Page DOWN
				topLine += screenHeight
				if topLine > maxTopLine {
					topLine = maxTopLine
				}
				renderPage()
				continue
			}
			if bytes.Equal(input, []byte("\x1b[H")) || bytes.Equal(input, []byte("\x1b[1~")) {
				// Home
				topLine = 0
				renderPage()
				continue
			}
			if bytes.Equal(input, []byte("\x1b[F")) || bytes.Equal(input, []byte("\x1b[4~")) {
				// End
				topLine = maxTopLine
				renderPage()
				continue
			}
		}

		char := input[0]
		switch char {
		case 'q', 'Q', 3, 4: // 'q', Ctrl+C, Ctrl+D
			return nil
		case ' ', 'f', 'F': // Espaço / f = Avançar página
			if topLine < maxTopLine {
				topLine += screenHeight
				if topLine > maxTopLine {
					topLine = maxTopLine
				}
				renderPage()
			}
		case 'b', 'B': // b = Voltar página
			if topLine > 0 {
				topLine -= screenHeight
				if topLine < 0 {
					topLine = 0
				}
				renderPage()
			}
		case 'j', '\r', '\n': // j ou Enter = Rolar 1 linha para baixo
			if topLine < maxTopLine {
				topLine++
				renderPage()
			}
		case 'k': // k = Rolar 1 linha para cima
			if topLine > 0 {
				topLine--
				renderPage()
			}
		case 'g': // g = Topo
			topLine = 0
			renderPage()
		case 'G': // G = Fim
			topLine = maxTopLine
			renderPage()
		case 'h', '?': // Ajuda do paginador
			// Exibe mensagem de ajuda temporária
			fmt.Fprint(p.Out, "\x1b[H\x1b[2J")
			fmt.Fprint(p.Out, "Comandos do Paginador de Ajuda do dBase II++:\r\n")
			fmt.Fprint(p.Out, "  [Espaço] / [f] / [PageDown] : Avançar uma tela\r\n")
			fmt.Fprint(p.Out, "  [b] / [PageUp]               : Voltar uma tela\r\n")
			fmt.Fprint(p.Out, "  [Seta Baixo] / [j] / [Enter] : Descer uma linha\r\n")
			fmt.Fprint(p.Out, "  [Seta Cima] / [k]            : Subir uma linha\r\n")
			fmt.Fprint(p.Out, "  [g] / [Home]                 : Ir para o início do texto\r\n")
			fmt.Fprint(p.Out, "  [G] / [End]                  : Ir para o final do texto\r\n")
			fmt.Fprint(p.Out, "  [q] / [Esc] / [Ctrl+C]       : Sair da ajuda\r\n\r\n")
			fmt.Fprint(p.Out, "Pressione qualquer tecla para retornar à ajuda...")
			_, _ = inFile.Read(buf)
			renderPage()
		}
	}

	return nil
}

// DisplayText helper conveniente para exibir texto usando o paginador padrão.
func DisplayText(text string, in io.Reader, out io.Writer) error {
	p := NewPager(in, out)
	return p.Display(text)
}

// DisplayBuffer helper para capturas de testes.
func DisplayBuffer(text string, inputSeq string) (string, error) {
	in := strings.NewReader(inputSeq)
	out := new(bytes.Buffer)
	scanner := bufio.NewScanner(in)
	_ = scanner
	lines := strings.Split(text, "\n")
	for _, l := range lines {
		out.WriteString(l + "\n")
	}
	return out.String(), nil
}
