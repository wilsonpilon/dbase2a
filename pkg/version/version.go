// Package version gerencia o versionamento semântico vX.Y.Z,
// metadados de compilação e codinomes temáticos de filmes de terror/horror para o dBase II++.
package version

import (
	"bytes"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"time"
)

// Info armazena os metadados de versão e compilação do dBase II++.
type Info struct {
	Major     int    `json:"major"`      // X: Pacote de features ou mudança de fase
	Minor     int    `json:"minor"`      // Y: Novo recurso adicionado ou terminado
	Patch     int    `json:"patch"`      // Z: Incrementado a cada compilação/build
	Codename  string `json:"codename"`   // Apelido do release (Filme de Terror/Horror)
	BuildDate string `json:"build_date"` // Data da compilação (YYYY-MM-DD)
	Commit    string `json:"commit"`     // Hash do commit Git ou 'dev'
}

// Variáveis injetadas em tempo de compilação via ldflags (-X)
var (
	BuildMajor    = "0"
	BuildMinor    = "6"
	BuildPatch    = "4"
	BuildCodename = "The Exorcist"
	BuildDateVal  = "2026-10-08"
	BuildCommit   = "dev"
)

// DefaultVersion define a versão padrão com base nas variáveis de build embutidas.
func getDefaultVersion() Info {
	var major, minor, patch int
	fmt.Sscanf(BuildMajor, "%d", &major)
	fmt.Sscanf(BuildMinor, "%d", &minor)
	fmt.Sscanf(BuildPatch, "%d", &patch)

	return Info{
		Major:     major,
		Minor:     minor,
		Patch:     patch,
		Codename:  BuildCodename,
		BuildDate: BuildDateVal,
		Commit:    BuildCommit,
	}
}

// Current armazena a versão ativa carregada em tempo de execução.
var Current = getDefaultVersion()

// VersionString retorna a string formatada no padrão "vX.Y.Z (Codename)".
func (v Info) VersionString() string {
	return fmt.Sprintf("v%d.%d.%d", v.Major, v.Minor, v.Patch)
}

// FullString retorna a descrição detalhada com versão, codinome e data de build.
func (v Info) FullString() string {
	return fmt.Sprintf("dBase II++ %s \"%s\" (Build: %s, Commit: %s)",
		v.VersionString(), v.Codename, v.BuildDate, v.Commit)
}

// Load lê o arquivo version.json a partir do caminho informado ou caminhos padrão.
func Load(path string) (Info, error) {
	def := getDefaultVersion()

	pathsToTry := []string{path}
	if path == "" || path == "version.json" {
		pathsToTry = []string{"version.json"}
		if exePath, err := os.Executable(); err == nil {
			pathsToTry = append(pathsToTry, filepath.Join(filepath.Dir(exePath), "version.json"))
		}
	}

	for _, p := range pathsToTry {
		data, err := os.ReadFile(p)
		if err == nil && len(data) > 0 {
			data = bytes.TrimPrefix(data, []byte("\xef\xbb\xbf"))
			var info Info
			if err := json.Unmarshal(data, &info); err == nil {
				Current = info
				return info, nil
			}
		}
	}

	Current = def
	return def, nil
}

// Save grava os metadados de versão no arquivo JSON especificado.
func (v Info) Save(path string) error {
	if path == "" {
		path = "version.json"
	}
	if err := os.MkdirAll(filepath.Dir(path), 0755); err != nil && filepath.Dir(path) != "." {
		return err
	}
	data, err := json.MarshalIndent(v, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, append(data, '\n'), 0644)
}

// BumpPatch incrementa o contador Z de compilação e atualiza a data de build.
func (v *Info) BumpPatch() {
	v.Patch++
	v.BuildDate = time.Now().Format("2006-01-02")
}
