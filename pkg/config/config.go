// Package config gerencia as opções de configuração do dBase II+,
// diretórios de dados, preferências de usuário e localização das toolchains MSX.
package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// AppConfig armazena as configurações globais da aplicação dBase II+.
type AppConfig struct {
	// DatabasePath indica o caminho do banco de dados SQLite para metadados e histórico.
	DatabasePath string `json:"database_path"`

	// DefaultDataDir diretório padrão onde arquivos .DBF e .CMD residem.
	DefaultDataDir string `json:"default_data_dir"`

	// InteractivePrompt símbolo do prompt no modo CLI (padrão é ponto '.').
	InteractivePrompt string `json:"interactive_prompt"`

	// Talk habilita mensagens detalhadas de feedback após operações (equivalente ao SET TALK ON).
	Talk bool `json:"talk"`

	// Echo ecoa comandos executados a partir de arquivos de script (SET ECHO ON).
	Echo bool `json:"echo"`

	// Bell emite sinal sonoro em erros ou avisos (SET BELL ON).
	Bell bool `json:"bell"`

	// ScreenColumns largura da tela do terminal ou TUI (80 padrão, ou 40 para modo MSX 1).
	ScreenColumns int `json:"screen_columns"`

	// ScreenLines altura da tela (24 ou 25 padrão).
	ScreenLines int `json:"screen_lines"`

	// MSXConfig configurações de compilação cruzada para MSX.
	MSX MSXToolchainConfig `json:"msx"`
}

// MSXToolchainConfig detalha caminhos e preferências da toolchain de compilação MSX.
type MSXToolchainConfig struct {
	TargetPlatform string `json:"target_platform"` // "msx1", "msx2", "msx2p", "msxtr"
	MemoryMapperKB int    `json:"memory_mapper_kb"` // 64, 128, 256, 512
	SDK            string `json:"sdk"`              // "fusion-c", "msxgl", "sdcc-raw"
	SDCCPath       string `json:"sdcc_path"`        // Caminho do executável do SDCC
	FusionCPath    string `json:"fusion_c_path"`    // Caminho da biblioteca Fusion-C
	MSXglPath      string `json:"msxgl_path"`       // Caminho do framework MSXgl
	EmulatorPath   string `json:"emulator_path"`    // Caminho do emulador (openMSX, BlueMSX, etc.)
}

// DefaultConfig retorna a configuração padrão de fábrica do dBase II+.
func DefaultConfig() *AppConfig {
	return &AppConfig{
		DatabasePath:      "dbase2plus.db",
		DefaultDataDir:    ".",
		InteractivePrompt: ".",
		Talk:              true,
		Echo:              false,
		Bell:              true,
		ScreenColumns:     80,
		ScreenLines:       24,
		MSX: MSXToolchainConfig{
			TargetPlatform: "msx2p",
			MemoryMapperKB: 256,
			SDK:            "fusion-c",
			SDCCPath:       "sdcc",
			FusionCPath:    "resources/msx/fusion-c",
			MSXglPath:      "resources/msx/msxgl",
			EmulatorPath:   "openmsx",
		},
	}
}

// Load lê o arquivo de configuração JSON ou retorna valores padrão caso não exista.
func Load(path string) (*AppConfig, error) {
	if path == "" {
		path = "config.json"
	}
	data, err := os.ReadFile(path)
	if err != nil {
		cfg := DefaultConfig()
		_ = cfg.Save(path)
		return cfg, nil
	}
	cfg := DefaultConfig()
	if err := json.Unmarshal(data, cfg); err != nil {
		return DefaultConfig(), err
	}
	return cfg, nil
}

// Save persiste a configuração atual no caminho especificado em formato JSON indentado.
func (c *AppConfig) Save(path string) error {
	if path == "" {
		path = "config.json"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil && filepath.Dir(path) != "." {
		return err
	}
	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}
