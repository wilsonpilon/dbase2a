package dbf

import (
	"path/filepath"
	"testing"
)

func TestDBase2CreateOpenAndRoundTrip(t *testing.T) {
	tempDir := t.TempDir()
	dbfPath := filepath.Join(tempDir, "TESTE.DBF")

	fields := []Field{
		{Name: "CODIGO", Type: TypeNumeric, Length: 5, Decimals: 0},
		{Name: "NOME", Type: TypeCharacter, Length: 20, Decimals: 0},
		{Name: "VALOR", Type: TypeNumeric, Length: 8, Decimals: 2},
		{Name: "ATIVO", Type: TypeLogical, Length: 1, Decimals: 0},
	}

	// 1. Criação da tabela
	tbl, err := CreateTable(dbfPath, fields)
	if err != nil {
		t.Fatalf("Erro ao criar tabela dBASE II: %v", err)
	}

	if tbl.Header.Version != DBase2VersionByte {
		t.Errorf("Versão esperada 0x%02X, obteve 0x%02X", DBase2VersionByte, tbl.Header.Version)
	}

	// 2. Inserção de registros
	r1 := tbl.AppendBlank()
	r1.Set("CODIGO", "101")
	r1.Set("NOME", "PRODUTO ALFA")
	r1.Set("VALOR", "150.50")
	r1.Set("ATIVO", "T")

	r2 := tbl.AppendBlank()
	r2.Set("CODIGO", "102")
	r2.Set("NOME", "PRODUTO BETA")
	r2.Set("VALOR", "89.90")
	r2.Set("ATIVO", "F")

	r3 := tbl.AppendBlank()
	r3.Set("CODIGO", "103")
	r3.Set("NOME", "PRODUTO GAMA")
	r3.Set("VALOR", "12.00")
	r3.Set("ATIVO", "T")
	r3.Deleted = true // Marcado para exclusão

	if err := tbl.Save(); err != nil {
		t.Fatalf("Erro ao salvar tabela DBF: %v", err)
	}

	// 3. Reabertura e validação binária
	loaded, err := OpenTable(dbfPath)
	if err != nil {
		t.Fatalf("Erro ao reabrir tabela DBF: %v", err)
	}

	if len(loaded.Records) != 3 {
		t.Fatalf("Esperava 3 registros, encontrou %d", len(loaded.Records))
	}

	if loaded.Records[0].Get("NOME") != "PRODUTO ALFA" {
		t.Errorf("Nome esperado 'PRODUTO ALFA', obteve '%s'", loaded.Records[0].Get("NOME"))
	}
	if loaded.Records[0].Get("ATIVO") != "T" {
		t.Errorf("Ativo esperado 'T', obteve '%s'", loaded.Records[0].Get("ATIVO"))
	}
	if !loaded.Records[2].Deleted {
		t.Errorf("Registro 3 deveria estar marcado como deletado")
	}

	// 4. Teste de PACK
	removed := loaded.Pack()
	if removed != 1 {
		t.Errorf("PACK deveria ter removido 1 registro, removeu %d", removed)
	}
	if len(loaded.Records) != 2 {
		t.Errorf("Após PACK esperava 2 registros, encontrou %d", len(loaded.Records))
	}

	_ = loaded.Save()

	// Validação final do arquivo após PACK
	reloaded, err := OpenTable(dbfPath)
	if err != nil {
		t.Fatalf("Erro ao reabrir tabela após PACK: %v", err)
	}
	if len(reloaded.Records) != 2 {
		t.Errorf("Esperava 2 registros após persistência do PACK, obteve %d", len(reloaded.Records))
	}
}

func TestExistingMSXDBFFiles(t *testing.T) {
	// Testa se existem arquivos DBF de recurso e tenta lê-los
	files, err := filepath.Glob("resource/files/*.DBF")
	if err != nil || len(files) == 0 {
		return
	}
	for _, f := range files {
		tbl, err := OpenTable(f)
		if err != nil {
			t.Logf("Aviso ao abrir arquivo MSX DBF (%s): %v", f, err)
			continue
		}
		if tbl.Header.Version != DBase2VersionByte {
			t.Errorf("Arquivo %s tem versão diferente de 0x02: 0x%02X", f, tbl.Header.Version)
		}
	}
}
