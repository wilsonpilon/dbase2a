// Package cli implementa a interface de linha de comando interativa (CLI)
// com o clássico prompt de ponto '.' do dBase II original.
package cli

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"dbase2a/pkg/engine"
	"dbase2a/pkg/version"
)

// Run inicia o loop interativo do prompt '.' do dBase II.
func Run(eng *engine.Engine) error {
	vInfo := version.Current
	fmt.Printf("dBase II+ %s \"%s\" (PC & MSX Edition)\n", vInfo.VersionString(), vInfo.Codename)
	fmt.Println("Digite HELP ou ? para ajuda, ou QUIT para encerrar.")
	fmt.Println()

	scanner := bufio.NewScanner(os.Stdin)
	prompt := eng.Config.InteractivePrompt
	if prompt == "" {
		prompt = "."
	}

	for {
		fmt.Printf("%s ", prompt)
		if !scanner.Scan() {
			break
		}

		line := strings.TrimSpace(scanner.Text())
		if line == "" {
			continue
		}

		if err := eng.ExecuteLine(line); err != nil {
			fmt.Printf("*** ERRO: %s\n", err.Error())
			if eng.Config.Bell {
				fmt.Print("\a") // Beep
			}
		}

		if eng.QuitRequested {
			break
		}
	}

	return scanner.Err()
}
