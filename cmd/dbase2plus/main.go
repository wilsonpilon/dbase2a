// dBase II+ (PC & MSX Edition)
// Ponto de entrada unificado da aplicação para Windows, Linux e compilação cruzada MSX.
package main

import (
	"flag"
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"dbase2a/pkg/config"
	"dbase2a/pkg/dbf"
	"dbase2a/pkg/engine"
	"dbase2a/pkg/msx"
	"dbase2a/pkg/storage/sqlite"
	"dbase2a/pkg/ui/cli"
	"dbase2a/pkg/ui/gui"
	"dbase2a/pkg/ui/tui"
	"dbase2a/pkg/version"
)

func main() {
	// Flags de modo de interface e execução
	cliMode := flag.Bool("cli", false, "Inicia o modo interativo CLI com prompt '.' do dBASE II")
	tuiMode := flag.Bool("tui", false, "Inicia o modo Terminal User Interface (TUI)")
	guiMode := flag.Bool("gui", false, "Inicia a interface gráfica desktop (Fyne)")
	execCmd := flag.String("exec", "", "Executa uma sequência de comandos dBASE II em lote e encerra")
	fileCmd := flag.String("file", "", "Executa um arquivo de script de comandos (.CMD ou .PRG)")
	
	// Flags de exportação MSX
	exportMSX := flag.String("export-msx", "", "Exporta projeto em C/SDCC para MSX a partir de um arquivo .DBF")
	msxTarget := flag.String("target", "msx2p", "Plataforma MSX de destino: msx1, msx2, msx2p, msxtr")
	msxSDK := flag.String("sdk", "fusion-c", "SDK para MSX: fusion-c, msxgl, sdcc")
	msxMapper := flag.Int("mapper", 256, "Tamanho da memória RAM mapeada (Memory Mapper) em KB")
	
	// Flags de configuração e sistema
	configPath := flag.String("config", "config.json", "Caminho do arquivo de configuração JSON")
	dbPath := flag.String("db", "dbase2plus.db", "Caminho do banco de dados SQLite")
	showVersion := flag.Bool("version", false, "Exibe informações de versão, codinome de terror e compilação")

	flag.Parse()

	// 1. Carrega informações de versão
	vInfo, _ := version.Load("version.json")

	if *showVersion {
		fmt.Println(vInfo.FullString())
		os.Exit(0)
	}

	// 2. Carrega configurações e banco de dados SQLite
	cfg, err := config.Load(*configPath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Aviso ao carregar config (%s): %v. Usando padrão.\n", *configPath, err)
		cfg = config.DefaultConfig()
	}

	if *dbPath != "" {
		cfg.DatabasePath = *dbPath
	}

	store, err := sqlite.Open(cfg.DatabasePath)
	if err != nil {
		fmt.Fprintf(os.Stderr, "Aviso ao inicializar SQLite (%s): %v. Prosseguindo sem persistência relacional.\n", cfg.DatabasePath, err)
	} else {
		defer store.Close()
	}

	// 3. Inicializa o motor de execução (Engine)
	eng := engine.NewEngine(cfg, store, os.Stdout)

	// 4. Tratamento do modo de exportação MSX via CLI
	if *exportMSX != "" {
		tbl, err := dbf.OpenTable(*exportMSX)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao abrir DBF para exportação MSX: %v\n", err)
			os.Exit(1)
		}

		projName := strings.TrimSuffix(filepath.Base(*exportMSX), filepath.Ext(*exportMSX))
		mgr := msx.NewToolchainManager(cfg, store)
		opts := msx.ProjectOptions{
			ProjectName:    strings.ToUpper(projName),
			Target:         msx.TargetPlatform(*msxTarget),
			SDK:            msx.SDKType(*msxSDK),
			MemoryMapperKB: *msxMapper,
			TableName:      strings.ToUpper(projName),
			Fields:         tbl.Header.Fields,
		}

		outDir := fmt.Sprintf("msx_export_%s", strings.ToLower(projName))
		if err := mgr.ExportProject(outDir, opts); err != nil {
			fmt.Fprintf(os.Stderr, "Erro na exportação MSX: %v\n", err)
			os.Exit(1)
		}

		fmt.Printf("Projeto MSX gerado com sucesso em '%s' para %s (%d KB RAM, SDK: %s)\n",
			outDir, *msxTarget, *msxMapper, *msxSDK)
		os.Exit(0)
	}

	// 5. Execução em lote via --exec
	if *execCmd != "" {
		if err := eng.ExecuteLine(*execCmd); err != nil {
			fmt.Fprintf(os.Stderr, "Erro na execução do comando: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// 6. Execução de script via --file
	if *fileCmd != "" {
		if err := eng.ExecuteScript(*fileCmd); err != nil {
			fmt.Fprintf(os.Stderr, "Erro na execução do script: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// 7. Roteamento de Interface: TUI, CLI ou GUI
	if *tuiMode {
		if err := tui.Run(eng); err != nil {
			fmt.Fprintf(os.Stderr, "Erro no TUI: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *cliMode {
		if err := cli.Run(eng); err != nil {
			fmt.Fprintf(os.Stderr, "Erro no CLI: %v\n", err)
			os.Exit(1)
		}
		return
	}

	if *guiMode {
		if err := gui.Run(eng); err != nil {
			fmt.Fprintf(os.Stderr, "Erro na GUI: %v\n", err)
			os.Exit(1)
		}
		return
	}

	// Se nenhum modo foi explicitamente passado:
	// Se stdin for um terminal interativo e não houver GUI requisitada, inicia CLI por padrão se headless, ou GUI.
	// No ambiente de linha de comando ou headless, iniciamos CLI como fallback robusto.
	if err := cli.Run(eng); err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}
}
