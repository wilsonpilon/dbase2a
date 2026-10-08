//go:build nogui

package gui

import (
	"fmt"

	"dbase2a/pkg/engine"
	"dbase2a/pkg/ui/cli"
)

// Run inicia o modo CLI como fallback quando o binário é compilado sem GUI (nogui).
func Run(eng *engine.Engine) error {
	fmt.Println("Nota: Binário compilado em modo headless/console (tag 'nogui'). Iniciando CLI...")
	return cli.Run(eng)
}
