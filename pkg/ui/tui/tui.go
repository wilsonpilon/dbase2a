// Package tui implementa uma interface em modo texto (Terminal User Interface)
// formatada com tabelas e molduras para terminais modernos compatíveis com ANSI.
package tui

import (
	"bufio"
	"fmt"
	"os"
	"strings"

	"dbase2a/pkg/engine"
	"dbase2a/pkg/version"
)

// Run inicia o modo TUI (Terminal User Interface).
func Run(eng *engine.Engine) error {
	vInfo := version.Current
	scanner := bufio.NewScanner(os.Stdin)

	for {
		renderScreen(eng, vInfo)
		fmt.Print("dBase II++ TUI > ")
		if !scanner.Scan() {
			break
		}

		cmd := strings.TrimSpace(scanner.Text())
		if cmd == "" {
			continue
		}

		if err := eng.ExecuteLine(cmd); err != nil {
			fmt.Printf("\n[ERRO]: %s\n", err.Error())
			fmt.Println("Pressione ENTER para continuar...")
			scanner.Scan()
		}

		if eng.QuitRequested {
			break
		}
	}

	return scanner.Err()
}

func renderScreen(eng *engine.Engine, vInfo version.Info) {
	fmt.Print("\033[H\033[2J") // Limpa tela ANSI
	fmt.Println("┌──────────────────────────────────────────────────────────────────────────────┐")
	fmt.Printf("│ dBase II++ %-14s [ %-12s ] - TERMINAL UI MODE          │\n", vInfo.VersionString(), vInfo.Codename)
	fmt.Println("├──────────────────────────────────────────────────────────────────────────────┤")

	if eng.ActiveTable != nil {
		fmt.Printf("│ Tabela Ativa: %-25s | Registros: %-6d | Campos: %-4d   │\n",
			eng.ActiveAlias, len(eng.ActiveTable.Records), len(eng.ActiveTable.Header.Fields))
	} else {
		fmt.Println("│ Nenhuma tabela aberta. Digite 'USE <arquivo>' para carregar dados.          │")
	}

	fmt.Println("├──────────────────────────────────────────────────────────────────────────────┤")
	fmt.Println("│ Comandos rápidos: LIST, LIST STRUCTURE, APPEND, PACK, QUIT, HELP             │")
	fmt.Println("└──────────────────────────────────────────────────────────────────────────────┘")
	fmt.Println()
}
