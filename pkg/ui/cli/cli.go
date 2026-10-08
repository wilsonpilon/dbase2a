// Package cli implementa a interface de linha de comando interativa (CLI)
// com o clássico prompt de ponto '.' do dBase II original, autocomplete com TAB
// cíclico, realce de sintaxe em tempo real e relatórios de erro estilizados em vermelho.
package cli

import (
	"fmt"
	"io"
	"os"
	"strings"

	"dbase2a/pkg/engine"
	"dbase2a/pkg/version"
)

// Run inicia o loop interativo do prompt '.' do dBase II com LineEditor integrado.
func Run(eng *engine.Engine) error {
	vInfo := version.Current
	fmt.Printf("dBase II++ %s \"%s\" (PC & MSX Edition)\n", vInfo.VersionString(), vInfo.Codename)
	fmt.Println("Digite HELP ou ? para ajuda, ou QUIT para encerrar.")
	fmt.Println()

	editor := NewLineEditor(os.Stdin, os.Stdout)
	prompt := eng.Config.InteractivePrompt
	if prompt == "" {
		prompt = "."
	}

	for {
		line, err := editor.ReadLine(prompt)
		if err != nil {
			if err == io.EOF {
				break
			}
			return err
		}

		trimmed := strings.TrimSpace(line)
		if trimmed == "" {
			continue
		}

		if err := eng.ExecuteLine(trimmed); err != nil {
			fmt.Println(FormatError(err.Error()))
			if eng.Config.Bell {
				fmt.Print("\a") // Beep
			}
		}

		if eng.QuitRequested {
			break
		}
	}

	return nil
}
