package engine

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"dbase2a/pkg/config"
	"dbase2a/pkg/dbf"
)

func TestEngineFullExecutionCycle(t *testing.T) {
	tempDir := t.TempDir()
	dbfPath := filepath.Join(tempDir, "LIVROS.DBF")

	cfg := config.DefaultConfig()
	cfg.Talk = false // Silencia para testar saída limpa
	buf := new(bytes.Buffer)

	eng := NewEngine(cfg, nil, buf)

	// 1. Criação manual da tabela
	fields := []dbf.Field{
		{Name: "ID", Type: dbf.TypeNumeric, Length: 4, Decimals: 0},
		{Name: "TITULO", Type: dbf.TypeCharacter, Length: 25, Decimals: 0},
		{Name: "PRECO", Type: dbf.TypeNumeric, Length: 8, Decimals: 2},
	}
	tbl, err := dbf.CreateTable(dbfPath, fields)
	if err != nil {
		t.Fatalf("Erro ao criar tabela: %v", err)
	}

	r1 := tbl.AppendBlank()
	r1.Set("ID", "1")
	r1.Set("TITULO", "DRACULA")
	r1.Set("PRECO", "45.00")

	r2 := tbl.AppendBlank()
	r2.Set("ID", "2")
	r2.Set("TITULO", "FRANKENSTEIN")
	r2.Set("PRECO", "50.00")

	_ = tbl.Save()

	// 2. Executa USE
	err = eng.ExecuteLine("USE " + dbfPath)
	if err != nil {
		t.Fatalf("Erro no USE: %v", err)
	}

	// 3. Executa LIST
	buf.Reset()
	err = eng.ExecuteLine("LIST ALL")
	if err != nil {
		t.Fatalf("Erro no LIST: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "DRACULA") || !strings.Contains(out, "FRANKENSTEIN") {
		t.Errorf("LIST ALL não retornou os registros esperados: %s", out)
	}

	// 4. Executa COUNT e SUM
	buf.Reset()
	_ = eng.ExecuteLine("COUNT FOR PRECO > 40")
	if !strings.Contains(buf.String(), "2 REGISTROS") {
		t.Errorf("COUNT incorreto: %s", buf.String())
	}

	buf.Reset()
	_ = eng.ExecuteLine("SUM PRECO")
	if !strings.Contains(buf.String(), "95.00") {
		t.Errorf("SUM incorreto: %s", buf.String())
	}

	// 5. Executa REPLACE e LOCATE
	_ = eng.ExecuteLine("REPLACE TITULO WITH NOSFERATU FOR ID = 1")
	if eng.ActiveTable.Records[0].Get("TITULO") != "NOSFERATU" {
		t.Errorf("REPLACE falhou, valor é %s", eng.ActiveTable.Records[0].Get("TITULO"))
	}

	// 6. Executa DELETE e PACK
	_ = eng.ExecuteLine("DELETE ALL FOR ID = 2")
	_ = eng.ExecuteLine("PACK")
	if len(eng.ActiveTable.Records) != 1 {
		t.Errorf("Após PACK esperava 1 registro, obteve %d", len(eng.ActiveTable.Records))
	}

	// 7. QUIT
	_ = eng.ExecuteLine("QUIT")
	if !eng.QuitRequested {
		t.Errorf("QUIT deveria ter setado QuitRequested como true")
	}
}
