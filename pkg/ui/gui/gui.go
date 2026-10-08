//go:build !nogui

// Package gui implementa a interface gráfica desktop utilizando o framework Fyne
// para visualização tabular de dados .DBF, editor de consultas e controle do pipeline MSX.
package gui

import (
	"fmt"
	"strings"

	"fyne.io/fyne/v2"
	"fyne.io/fyne/v2/app"
	"fyne.io/fyne/v2/container"
	"fyne.io/fyne/v2/widget"

	"dbase2a/pkg/engine"
	"dbase2a/pkg/msx"
	"dbase2a/pkg/version"
)

// Run inicia a interface gráfica Fyne para o dBase II++.
func Run(eng *engine.Engine) error {
	vInfo := version.Current
	a := app.NewWithID("com.dbase2plus.app")
	w := a.NewWindow(fmt.Sprintf("dBase II++ %s \"%s\" - PC & MSX Edition", vInfo.VersionString(), vInfo.Codename))
	w.Resize(fyne.NewSize(900, 650))

	// Status e console de saída
	consoleOutput := widget.NewMultiLineEntry()
	consoleOutput.SetText(fmt.Sprintf("=== dBase II++ %s \"%s\" Inicializado ===\nDigite comandos na barra inferior ou use a barra de ferramentas.\n\n", vInfo.VersionString(), vInfo.Codename))
	consoleOutput.Wrapping = fyne.TextWrapWord

	statusLabel := widget.NewLabel("Status: Pronto | Nenhuma tabela carregada.")

	// Redireciona saídas do motor para o console da GUI
	eng.Out = &guiWriter{entry: consoleOutput}

	// Campo de entrada de comandos
	cmdEntry := widget.NewEntry()
	cmdEntry.SetPlaceHolder("Digite comando dBASE II (ex: USE CLIENTES; LIST ALL)...")

	var runCmd func()
	runCmd = func() {
		text := strings.TrimSpace(cmdEntry.Text)
		if text == "" {
			return
		}
		consoleOutput.SetText(consoleOutput.Text + fmt.Sprintf(". %s\n", text))
		cmdEntry.SetText("")

		if err := eng.ExecuteLine(text); err != nil {
			consoleOutput.SetText(consoleOutput.Text + fmt.Sprintf("*** ERRO: %s\n", err.Error()))
		}

		if eng.ActiveTable != nil {
			statusLabel.SetText(fmt.Sprintf("Status: Tabela [%s] | Registros: %d | Campos: %d",
				eng.ActiveAlias, len(eng.ActiveTable.Records), len(eng.ActiveTable.Header.Fields)))
		} else {
			statusLabel.SetText("Status: Nenhuma tabela em uso.")
		}

		if eng.QuitRequested {
			w.Close()
		}
	}

	cmdEntry.OnSubmitted = func(s string) {
		runCmd()
	}

	runBtn := widget.NewButton("Executar", runCmd)

	// Aba MSX Generator
	msxProjName := widget.NewEntry()
	msxProjName.SetText("APPMSX")

	msxTargetSelect := widget.NewSelect([]string{"msx2p", "msx2", "msx1", "msxtr"}, func(s string) {})
	msxTargetSelect.SetSelected("msx2p")

	msxSDKSelect := widget.NewSelect([]string{"fusion-c", "msxgl", "sdcc"}, func(s string) {})
	msxSDKSelect.SetSelected("fusion-c")

	msxMapperSelect := widget.NewSelect([]string{"256 KB", "128 KB", "512 KB", "64 KB"}, func(s string) {})
	msxMapperSelect.SetSelected("256 KB")

	exportMSXBtn := widget.NewButton("Exportar Projeto MSX", func() {
		if eng.ActiveTable == nil {
			consoleOutput.SetText(consoleOutput.Text + "*** ERRO: Abra uma tabela .DBF primeiro para exportar para MSX.\n")
			return
		}

		mgr := msx.NewToolchainManager(eng.Config, eng.Storage)
		opts := msx.ProjectOptions{
			ProjectName:    msxProjName.Text,
			Target:         msx.TargetPlatform(msxTargetSelect.Selected),
			SDK:            msx.SDKType(msxSDKSelect.Selected),
			MemoryMapperKB: 256,
			TableName:      eng.ActiveAlias,
			Fields:         eng.ActiveTable.Header.Fields,
		}

		outDir := fmt.Sprintf("msx_build_%s", strings.ToLower(msxProjName.Text))
		if err := mgr.ExportProject(outDir, opts); err != nil {
			consoleOutput.SetText(consoleOutput.Text + fmt.Sprintf("*** ERRO MSX: %s\n", err.Error()))
		} else {
			consoleOutput.SetText(consoleOutput.Text + fmt.Sprintf(">>> Projeto MSX exportado com sucesso na pasta '%s'!\n", outDir))
		}
	})

	msxForm := container.NewVBox(
		widget.NewLabelWithStyle("Gerador e Compilador de Código para MSX", fyne.TextAlignLeading, fyne.TextStyle{Bold: true}),
		widget.NewLabel("Nome do Projeto:"),
		msxProjName,
		widget.NewLabel("Plataforma MSX Alvo:"),
		msxTargetSelect,
		widget.NewLabel("SDK / Framework:"),
		msxSDKSelect,
		widget.NewLabel("Memória RAM Mapeada (Mapper):"),
		msxMapperSelect,
		exportMSXBtn,
	)

	// Abas principais
	tabs := container.NewAppTabs(
		container.NewTabItem("Console & Dados", container.NewBorder(nil, nil, nil, nil, consoleOutput)),
		container.NewTabItem("Pipeline MSX", msxForm),
	)

	bottomBar := container.NewBorder(nil, statusLabel, nil, runBtn, cmdEntry)
	mainContent := container.NewBorder(nil, bottomBar, nil, nil, tabs)

	w.SetContent(mainContent)
	w.ShowAndRun()
	return nil
}

type guiWriter struct {
	entry *widget.Entry
}

func (g *guiWriter) Write(p []byte) (n int, err error) {
	if g.entry != nil {
		g.entry.SetText(g.entry.Text + string(p))
	}
	return len(p), nil
}
