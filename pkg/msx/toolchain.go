package msx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"

	"dbase2a/pkg/config"
	"dbase2a/pkg/storage/sqlite"
)

// ToolchainManager coordena as ferramentas de build cruzado para o MSX.
type ToolchainManager struct {
	Config  *config.AppConfig
	Storage *sqlite.Storage
}

// NewToolchainManager instancia o gerenciador de toolchain MSX.
func NewToolchainManager(cfg *config.AppConfig, store *sqlite.Storage) *ToolchainManager {
	if cfg == nil {
		cfg = config.DefaultConfig()
	}
	return &ToolchainManager{
		Config:  cfg,
		Storage: store,
	}
}

// LocateSDCC procura o executável do SDCC nas configurações, no diretório de recursos embutido ou no PATH do sistema.
func (m *ToolchainManager) LocateSDCC() string {
	if m.Config != nil && m.Config.MSX.SDCCPath != "" {
		if _, err := os.Stat(m.Config.MSX.SDCCPath); err == nil {
			return m.Config.MSX.SDCCPath
		}
	}

	// Verifica pacote bundled no repositório
	bundledWin := filepath.Join("resource", "msx", "MSXgl", "tools", "sdcc", "bin", "sdcc.exe")
	bundledLinux := filepath.Join("resource", "msx", "MSXgl", "tools", "sdcc", "bin", "sdcc")

	if runtime.GOOS == "windows" {
		if _, err := os.Stat(bundledWin); err == nil {
			return bundledWin
		}
	} else {
		if _, err := os.Stat(bundledLinux); err == nil {
			return bundledLinux
		}
	}

	// Verifica PATH do sistema operacional
	if path, err := exec.LookPath("sdcc"); err == nil {
		return path
	}

	return ""
}

// CheckEnvironment verifica a disponibilidade dos executáveis de compilação (SDCC, Make, Emulador).
func (m *ToolchainManager) CheckEnvironment() map[string]bool {
	status := make(map[string]bool)

	sdccPath := m.LocateSDCC()
	status["sdcc"] = (sdccPath != "")

	_, err := exec.LookPath("make")
	status["make"] = (err == nil)

	emuPath := ""
	if m.Config != nil {
		emuPath = m.Config.MSX.EmulatorPath
	}
	if emuPath == "" {
		emuPath = "openmsx"
	}
	_, err = exec.LookPath(emuPath)
	status["emulator"] = (err == nil)

	return status
}

// ExportProject gera os arquivos de código-fonte C, cabeçalhos de configuração, scripts e Makefile no diretório de saída.
func (m *ToolchainManager) ExportProject(outputDir string, opts ProjectOptions) error {
	if err := os.MkdirAll(outputDir, 0755); err != nil {
		return fmt.Errorf("erro ao criar pasta do projeto MSX (%s): %w", outputDir, err)
	}

	// 1. Gera main.c
	cSource, err := GenerateCSource(opts)
	if err != nil {
		return err
	}
	mainPath := filepath.Join(outputDir, "main.c")
	if err := os.WriteFile(mainPath, []byte(cSource), 0644); err != nil {
		return fmt.Errorf("erro ao gravar main.c: %w", err)
	}

	// 2. Gera msxgl_config.h
	configH := GenerateMSXglConfig(opts)
	configPath := filepath.Join(outputDir, "msxgl_config.h")
	if err := os.WriteFile(configPath, []byte(configH), 0644); err != nil {
		return fmt.Errorf("erro ao gravar msxgl_config.h: %w", err)
	}

	// 3. Gera project_config.js
	jsConfig := GenerateMSXglProjectConfig(opts)
	jsPath := filepath.Join(outputDir, "project_config.js")
	if err := os.WriteFile(jsPath, []byte(jsConfig), 0644); err != nil {
		return fmt.Errorf("erro ao gravar project_config.js: %w", err)
	}

	// 4. Gera Makefile
	makeContent := GenerateMakefile(opts)
	makePath := filepath.Join(outputDir, "Makefile")
	if err := os.WriteFile(makePath, []byte(makeContent), 0644); err != nil {
		return fmt.Errorf("erro ao gravar Makefile: %w", err)
	}

	// 5. Gera scripts de compilação (.bat e .sh)
	batContent, shContent := GenerateBuildScripts(opts)
	batPath := filepath.Join(outputDir, "build.bat")
	_ = os.WriteFile(batPath, []byte(batContent), 0755)

	shPath := filepath.Join(outputDir, "build.sh")
	_ = os.WriteFile(shPath, []byte(shContent), 0755)

	// 6. Registra no banco SQLite se disponível
	if m.Storage != nil {
		_ = m.Storage.LogMSXBuild(opts.ProjectName, string(opts.Target), string(opts.SDK), outputDir, "EXPORTED", "Projeto MSX C/MSXgl exportado com sucesso.")
	}

	return nil
}

// CompileProject exporta o projeto e invoca a compilação do SDCC para gerar o executável .COM.
func (m *ToolchainManager) CompileProject(outputDir string, opts ProjectOptions) (string, error) {
	if err := m.ExportProject(outputDir, opts); err != nil {
		return "", err
	}

	sdccExec := m.LocateSDCC()
	if sdccExec == "" {
		return "", fmt.Errorf("compilador SDCC não encontrado no sistema ou em resource/msx/MSXgl")
	}

	mainFile := filepath.Join(outputDir, "main.c")
	ihxFile := filepath.Join(outputDir, opts.ProjectName+".ihx")
	comFile := filepath.Join(outputDir, opts.ProjectName+".COM")

	// Invoca SDCC
	cmd := exec.Command(sdccExec, "-mz80", "--no-std-crt0", "--code-loc", "0x0100", "--data-loc", "0x0000", "-I"+outputDir, mainFile, "-o", ihxFile)
	output, err := cmd.CombinedOutput()
	if err != nil {
		return "", fmt.Errorf("falha na compilação SDCC: %s (%w)", string(output), err)
	}

	// Se o arquivo .ihx foi gerado, converte diretamente para executável .COM
	if _, err := os.Stat(ihxFile); err == nil {
		if convErr := ConvertIHXFileToCOM(ihxFile, comFile); convErr != nil {
			return "", fmt.Errorf("erro na conversão de IHX para COM: %w", convErr)
		}
	}

	if m.Storage != nil {
		_ = m.Storage.LogMSXBuild(opts.ProjectName, string(opts.Target), string(opts.SDK), comFile, "SUCCESS", "Binário .COM gerado com sucesso via SDCC.")
	}

	return comFile, nil
}

// GenerateBootDisk cria a imagem de disco FAT12 (.DSK de 720KB) com MSX-DOS 2, dBase II++ e arquivos de sistema.
func (m *ToolchainManager) GenerateBootDisk(dskPath string, customFiles map[string][]byte) error {
	if dskPath == "" {
		dskPath = filepath.Join("release", "dbase2plus.dsk")
	}

	if err := BuildMSXDOS2BootDisk(dskPath, customFiles); err != nil {
		return fmt.Errorf("erro ao gerar disquete de boot MSX-DOS 2: %w", err)
	}

	if m.Storage != nil {
		_ = m.Storage.LogMSXBuild("DBASE2PLUS", "msx2p", "msxgl", dskPath, "DSK_SUCCESS", "Disquete bootável MSX-DOS 2 de 720KB gerado com sucesso.")
	}

	return nil
}
