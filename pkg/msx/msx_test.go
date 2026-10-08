package msx

import (
	"os"
	"path/filepath"
	"strings"
	"testing"

	"dbase2a/pkg/config"
	"dbase2a/pkg/dbf"
)

func TestGenerateCSource_MSXgl(t *testing.T) {
	opts := ProjectOptions{
		ProjectName:    "CLIENTES",
		AppTitle:       "Controle de Clientes MSX",
		Target:         TargetMSX2P,
		SDK:            SDKMSXgl,
		MemoryMapperKB: 256,
		TableName:      "CLIENTES",
		GenerateMapper: true,
		Fields: []dbf.Field{
			{Name: "CODIGO", Type: dbf.TypeNumeric, Length: 5},
			{Name: "NOME", Type: dbf.TypeCharacter, Length: 20},
			{Name: "ATIVO", Type: dbf.TypeLogical, Length: 1},
		},
	}

	src, err := GenerateCSource(opts)
	if err != nil {
		t.Fatalf("Erro ao gerar fonte C com MSXgl: %v", err)
	}

	if !strings.Contains(src, "Record_CLIENTES") {
		t.Errorf("Fonte C gerado não contém a struct esperada")
	}
	if !strings.Contains(src, "msxgl.h") {
		t.Errorf("Fonte C deveria incluir msxgl.h")
	}
	if !strings.Contains(src, "dos_mapper.h") {
		t.Errorf("Fonte C deveria incluir dos_mapper.h")
	}
	if !strings.Contains(src, "256 KB") {
		t.Errorf("Fonte C deveria indicar 256 KB de Mapper")
	}
	if !strings.Contains(src, "VDP_SetMode(VDP_MODE_TEXT1)") {
		t.Errorf("Fonte C deveria configurar modo texto 80 colunas (VDP_MODE_TEXT1)")
	}
}

func TestGenerateMSXglConfigAndProject(t *testing.T) {
	opts := ProjectOptions{
		ProjectName: "PRODUTOS",
		TableName:   "PRODUTOS",
	}

	cfgH := GenerateMSXglConfig(opts)
	if !strings.Contains(cfgH, "TARGET_DOS2") || !strings.Contains(cfgH, "DOS_USE_MAPPER") {
		t.Errorf("msxgl_config.h não contém configurações esperadas de DOS2/Mapper")
	}

	projJs := GenerateMSXglProjectConfig(opts)
	if !strings.Contains(projJs, "PRODUTOS") || !strings.Contains(projJs, "LibModules") {
		t.Errorf("project_config.js não contém os módulos de biblioteca necessários")
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
		SDK:            SDKMSXgl,
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

	filesToCheck := []string{"main.c", "msxgl_config.h", "project_config.js", "Makefile", "build.bat", "build.sh"}
	for _, f := range filesToCheck {
		if _, err := os.Stat(filepath.Join(outDir, f)); os.IsNotExist(err) {
			t.Errorf("Arquivo %s não foi criado na pasta exportada", f)
		}
	}
}

func TestHex2BinConversion(t *testing.T) {
	// Exemplo de Intel HEX simples com endereço base 0x0100
	// :03 0100 00 C3 05 01 34 (3 bytes em 0x0100: JP 0x0105, Checksum 0x34)
	// :00 0000 01 FF (EOF)
	ihxSample := ":03010000C3050134\n:00000001FF\n"

	bin, err := ConvertIHXToBin(ihxSample, 0x0100)
	if err != nil {
		t.Fatalf("Erro na conversão Intel HEX para binário: %v", err)
	}

	if len(bin) != 3 {
		t.Fatalf("Esperado 3 bytes, obtido %d", len(bin))
	}

	if bin[0] != 0xC3 || bin[1] != 0x05 || bin[2] != 0x01 {
		t.Errorf("Bytes convertidos incorretos: %v", bin)
	}
}
