package msx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dbase2a/pkg/config"
	"dbase2a/pkg/dbf"
)

func TestGenerateCSource(t *testing.T) {
	opts := ProjectOptions{
		ProjectName:    "CLIENTES",
		Target:         TargetMSX2P,
		SDK:            SDKFusionC,
		MemoryMapperKB: 256,
		TableName:      "CLIENTES",
		Fields: []dbf.Field{
			{Name: "CODIGO", Type: dbf.TypeNumeric, Length: 5},
			{Name: "NOME", Type: dbf.TypeCharacter, Length: 20},
		},
	}

	src, err := GenerateCSource(opts)
	if err != nil {
		t.Fatalf("Erro ao gerar fonte C: %v", err)
	}

	if !strings.Contains(src, "Record_CLIENTES") {
		t.Errorf("Fonte C gerado não contém a struct esperada")
	}
	if !strings.Contains(src, "msx_fusion.h") {
		t.Errorf("Fonte C deveria incluir cabeçalho Fusion-C")
	}
	if !strings.Contains(src, "256 KB") {
		t.Errorf("Fonte C deveria indicar 256 KB de Mapper")
	}
}

func TestExportProject(t *testing.T) {
	tempDir := t.TempDir()
	outDir := filepath.Join(tempDir, "msx_proj")

	cfg := config.DefaultConfig()
	mgr := NewToolchainManager(cfg, nil)

	opts := ProjectOptions{
		ProjectName:    "ESTOQUE",
		Target:         TargetMSX2P,
		SDK:            SDKFusionC,
		MemoryMapperKB: 256,
		TableName:      "ESTOQUE",
		Fields: []dbf.Field{
			{Name: "ITEM", Type: dbf.TypeCharacter, Length: 15},
		},
	}

	err := mgr.ExportProject(outDir, opts)
	if err != nil {
		t.Fatalf("Erro ao exportar projeto MSX: %v", err)
	}

	if _, err := os.Stat(filepath.Join(outDir, "main.c")); os.IsNotExist(err) {
		t.Errorf("main.c não foi criado na pasta exportada")
	}
	if _, err := os.Stat(filepath.Join(outDir, "Makefile")); os.IsNotExist(err) {
		t.Errorf("Makefile não foi criado na pasta exportada")
	}
}
