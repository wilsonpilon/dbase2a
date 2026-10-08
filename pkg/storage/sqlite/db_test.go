package sqlite

import (
	"path/filepath"
	"testing"
	"time"
)

func TestSQLiteStorage(t *testing.T) {
	tempDir := t.TempDir()
	dbPath := filepath.Join(tempDir, "test_dbase2plus.db")

	store, err := Open(dbPath)
	if err != nil {
		t.Fatalf("Erro ao abrir SQLite: %v", err)
	}
	defer store.Close()

	// 1. Config store
	if err := store.SetConfig("PROMPT", "."); err != nil {
		t.Fatalf("Erro ao salvar config: %v", err)
	}
	val, err := store.GetConfig("PROMPT")
	if err != nil || val != "." {
		t.Errorf("Config esperada '.', obteve '%s' (err: %v)", val, err)
	}

	// 2. Command history
	if err := store.LogCommand("USE TESTE", 10*time.Millisecond, true, ""); err != nil {
		t.Fatalf("Erro ao registrar comando: %v", err)
	}
	history, err := store.GetRecentCommands(10)
	if err != nil || len(history) != 1 {
		t.Fatalf("Esperava 1 registro no histórico, obteve %d (err: %v)", len(history), err)
	}
	if history[0].Command != "USE TESTE" || !history[0].Success {
		t.Errorf("Comando registrado incorretamente: %+v", history[0])
	}

	// 3. Table catalog
	if err := store.CatalogTable("CLIENTES", "CLIENTES.DBF", 100, 5, 120, time.Now()); err != nil {
		t.Fatalf("Erro ao catalogar tabela: %v", err)
	}

	// 4. MSX Build log
	if err := store.LogMSXBuild("APPMSX", "msx2p", "fusion-c", "msx_build_appmsx", "SUCCESS", "Build ok"); err != nil {
		t.Fatalf("Erro ao registrar build MSX: %v", err)
	}
}
