package version

import (
	"path/filepath"
	"testing"
)

func TestVersionBumping(t *testing.T) {
	tempDir := t.TempDir()
	vFile := filepath.Join(tempDir, "version.json")

	info := Info{
		Major:    0,
		Minor:    1,
		Patch:    0,
		Codename: "Nosferatu",
	}

	if err := info.Save(vFile); err != nil {
		t.Fatalf("Erro ao salvar version.json: %v", err)
	}

	loaded, err := Load(vFile)
	if err != nil {
		t.Fatalf("Erro ao carregar version.json: %v", err)
	}

	if loaded.VersionString() != "v0.1.0" {
		t.Errorf("Versão esperada 'v0.1.0', obteve '%s'", loaded.VersionString())
	}

	loaded.BumpPatch()
	if loaded.Patch != 1 {
		t.Errorf("Esperava patch 1 após bump, obteve %d", loaded.Patch)
	}
	if loaded.VersionString() != "v0.1.1" {
		t.Errorf("Versão esperada 'v0.1.1', obteve '%s'", loaded.VersionString())
	}
}
