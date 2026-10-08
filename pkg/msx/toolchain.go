package msx

import (
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

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
	return &ToolchainManager{
		Config:  cfg,
		Storage: store,
	}
}

// CheckEnvironment verifica a disponibilidade dos executáveis de compilação (SDCC, Make, etc.).
func (m *ToolchainManager) CheckEnvironment() map[string]bool {
	status := make(map[string]bool)

	sdccPath := m.Config.MSX.SDCCPath
	if sdccPath == "" {
		sdccPath = "sdcc"
	}
	_, err := exec.LookPath(sdccPath)
	status["sdcc"] = (err == nil)

	_, err = exec.LookPath("make")
	status["make"] = (err == nil)

	emuPath := m.Config.MSX.EmulatorPath
	if emuPath == "" {
		emuPath = "openmsx"
	}
	_, err = exec.LookPath(emuPath)
	status["emulator"] = (err == nil)

	return status
}

// ExportProject gera os arquivos de código-fonte C, cabeçalhos e Makefile em um diretório de saída.
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

	// 2. Gera Makefile
	makeContent := GenerateMakefile(opts)
	makePath := filepath.Join(outputDir, "Makefile")
	if err := os.WriteFile(makePath, []byte(makeContent), 0644); err != nil {
		return fmt.Errorf("erro ao gravar Makefile: %w", err)
	}

	// 3. Registra no banco SQLite se disponível
	if m.Storage != nil {
		_ = m.Storage.LogMSXBuild(opts.ProjectName, string(opts.Target), string(opts.SDK), outputDir, "EXPORTED", "Projeto MSX C/SDCC exportado.")
	}

	return nil
}
