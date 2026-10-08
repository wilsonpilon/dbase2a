// dBase II++ (PC & MSX Edition)
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

	// Flags de exportação e compilação MSX
	exportMSX := flag.String("export-msx", "", "Exporta projeto em C/SDCC para MSX a partir de um arquivo .DBF")
	compileMSX := flag.String("compile-msx", "", "Exporta e compila executável .COM para MSX a partir de um arquivo .DBF")
	hex2comFile := flag.String("hex2com", "", "Converte um arquivo Intel HEX (.ihx) gerado pelo SDCC em binário MSX-DOS .COM")
	outPath := flag.String("out", "", "Caminho de saída para arquivo gerado (.COM, .DBF, etc.)")
	msxTarget := flag.String("target", "msx2p", "Plataforma MSX de destino: msx1, msx2, msx2p, msxtr")
	msxSDK := flag.String("sdk", "msxgl", "SDK para MSX: msxgl, fusion-c, sdcc")
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

	// 2. Utilitário Hex to COM
	if *hex2comFile != "" {
		dest := *outPath
		if dest == "" {
			dest = strings.TrimSuffix(*hex2comFile, filepath.Ext(*hex2comFile)) + ".COM"
		}
		if err := msx.ConvertIHXFileToCOM(*hex2comFile, dest); err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao converter IHX para COM: %v\n", err)
			os.Exit(1)
		}
		fmt.Printf("Arquivo COM gerado com sucesso: %s\n", dest)
		os.Exit(0)
	}

	// 3. Carrega configurações e banco de dados SQLite
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

	// 4. Inicializa o motor de execução dBASE II
	eng := engine.NewEngine(cfg, store, os.Stdout)

	// 5. Exportação e Compilação para MSX
	if *exportMSX != "" || *compileMSX != "" {
		targetDBF := *exportMSX
		doCompile := false
		if *compileMSX != "" {
			targetDBF = *compileMSX
			doCompile = true
		}

		tbl, err := dbf.OpenTable(targetDBF)
		if err != nil {
			fmt.Fprintf(os.Stderr, "Erro ao ler arquivo .DBF (%s): %v\n", targetDBF, err)
			os.Exit(1)
		}
		defer tbl.Close()

		baseName := strings.TrimSuffix(filepath.Base(targetDBF), filepath.Ext(targetDBF))
		outDir := filepath.Join("msx_export_" + strings.ToLower(baseName))
		if *outPath != "" {
			outDir = *outPath
		}

		opts := msx.ProjectOptions{
			ProjectName:    strings.ToUpper(baseName),
			Target:         msx.TargetPlatform(*msxTarget),
			SDK:            msx.SDKType(*msxSDK),
			MemoryMapperKB: *msxMapper,
			TableName:      strings.ToUpper(baseName),
			Fields:         tbl.Header.Fields,
			GenerateMapper: *msxMapper > 0,
			InteractiveCLI: true,
		}

		toolchain := msx.NewToolchainManager(cfg, store)
		if doCompile {
			comFile, err := toolchain.CompileProject(outDir, opts)
			if err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao compilar projeto MSX: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Projeto MSX compilado com sucesso! Executavel gerado em: %s\n", comFile)
		} else {
			if err := toolchain.ExportProject(outDir, opts); err != nil {
				fmt.Fprintf(os.Stderr, "Erro ao exportar projeto MSX: %v\n", err)
				os.Exit(1)
			}
			fmt.Printf("Projeto MSX gerado com sucesso em '%s' para %s (%d KB RAM, SDK: %s)\n",
				outDir, *msxTarget, *msxMapper, *msxSDK)
		}
		os.Exit(0)
	}

	// 6. Execução em lote via --exec
	if *execCmd != "" {
		if err := eng.ExecuteLine(*execCmd); err != nil {
			fmt.Fprintf(os.Stderr, "Erro na execução do comando: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// 7. Execução de script via --file
	if *fileCmd != "" {
		if err := eng.ExecuteScript(*fileCmd); err != nil {
			fmt.Fprintf(os.Stderr, "Erro na execução do script: %v\n", err)
			os.Exit(1)
		}
		os.Exit(0)
	}

	// 8. Roteamento de Interface: TUI, CLI ou GUI
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

	// Se nenhum modo foi explicitamente passado, iniciamos CLI como fallback robusto.
	if err := cli.Run(eng); err != nil {
		fmt.Fprintf(os.Stderr, "Erro: %v\n", err)
		os.Exit(1)
	}
}
