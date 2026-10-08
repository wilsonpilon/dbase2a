// Package msx implementa a geração de código, templates de frontends modernos
// e a ponte de compilação cruzada para o computador MSX (especialmente MSX 2+ com 256KB e MSX-DOS 2).
package msx

import (
	"bytes"
	"fmt"
	"strings"
	"text/template"

	"dbase2a/pkg/dbf"
)

// TargetPlatform especifica o modelo da máquina MSX de destino.
type TargetPlatform string

const (
	TargetMSX1  TargetPlatform = "msx1"
	TargetMSX2  TargetPlatform = "msx2"
	TargetMSX2P TargetPlatform = "msx2p" // MSX 2+ (Padrão recomendado com 256KB)
	TargetMSXTR TargetPlatform = "msxtr" // MSX Turbo R
)

// SDKType define a biblioteca base para geração de código C.
type SDKType string

const (
	SDKFusionC SDKType = "fusion-c" // Fusion-C (MSX Fusion)
	SDKMSXgl   SDKType = "msxgl"    // MSXgl Game & App Library
	SDKSDCCRaw SDKType = "sdcc"     // SDCC com bibliotecas padrão
)

// ProjectOptions parametriza a geração do projeto e frontend MSX.
type ProjectOptions struct {
	ProjectName    string         `json:"project_name"`
	Target         TargetPlatform `json:"target"`
	SDK            SDKType        `json:"sdk"`
	MemoryMapperKB int            `json:"memory_mapper_kb"` // ex: 256
	ScreenMode     int            `json:"screen_mode"`      // 0 = 80 colunas (MSX2+), 1, 2
	TableName      string         `json:"table_name"`
	Fields         []dbf.Field    `json:"fields"`
	GenerateMapper bool           `json:"generate_mapper"`
}

// GenerateCSource gera o arquivo de código-fonte C compatível com SDCC e Fusion-C/MSXgl.
func GenerateCSource(opts ProjectOptions) (string, error) {
	tmplStr := `/* =============================================================================
 * dBase II+ Frontend Gerado para MSX (MSX-DOS 2 / {{.Target}})
 * Projeto: {{.ProjectName}}
 * SDK: {{.SDK}} | Memória: {{.MemoryMapperKB}} KB Mapper
 * Totalmente compatível com arquivos dBASE II (.DBF) de 521 bytes
 * =============================================================================
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

{{if eq .SDK "fusion-c"}}
#include "fusion-c/header/msx_fusion.h"
#include "fusion-c/header/vdp_modes.h"
{{else if eq .SDK "msxgl"}}
#include "msxgl.h"
{{end}}

#define DBASE2_HEADER_SIZE 521
#define DBASE2_VERSION_BYTE 0x02

/* Estrutura de dados para os registros da tabela {{.TableName}} */
typedef struct {
    char deleted; /* ' ' = Ativo, '*' = Excluído */
{{range .Fields}}    char {{.Name}}[{{.Length}} + 1]; /* Tipo: {{printf "%c" .Type}} */
{{end}}} Record_{{.TableName}};

/* Protótipos de Funções */
void init_screen(void);
void draw_header(void);
void display_record(int rec_no, Record_{{.TableName}} *rec);
int load_dbase2_table(const char *filename);

int main(void) {
    init_screen();
    draw_header();

    printf("\n\rIniciando dBase II+ Client para MSX...\n\r");
    printf("Tabela associada: %s.DBF\n\r", "{{.TableName}}");
    printf("Memoria RAM Mapper mapeada: %d KB\n\r\n\r", {{.MemoryMapperKB}});

    /* Carrega e visualiza a tabela */
    load_dbase2_table("{{.TableName}}.DBF");

    printf("\n\rPressione qualquer tecla para retornar ao MSX-DOS 2...\n\r");
    getchar();
    return 0;
}

void init_screen(void) {
{{if eq .SDK "fusion-c"}}
    /* Configura tela texto 80 colunas no MSX 2+ (Screen 0) */
    Screen(0);
    Width(80);
{{else}}
    /* Inicialização padrão para MSX-DOS */
    printf("\033[2J"); /* Limpa tela ANSI */
{{end}}
}

void draw_header(void) {
    printf("================================================================================\n\r");
    printf("  dBase II+ MSX 2+ / MSX-DOS 2 Client  --  [ %s ]\n\r", "{{.ProjectName}}");
    printf("================================================================================\n\r");
}

void display_record(int rec_no, Record_{{.TableName}} *rec) {
    printf("[%04d]%c ", rec_no, rec->deleted);
{{range .Fields}}    printf("%s: %s | ", "{{.Name}}", rec->{{.Name}});
{{end}}    printf("\n\r");
}

int load_dbase2_table(const char *filename) {
    FILE *fp = fopen(filename, "rb");
    if (!fp) {
        printf("ERRO: Nao foi possivel abrir o arquivo %s\n\r", filename);
        return 0;
    }

    unsigned char header[DBASE2_HEADER_SIZE];
    if (fread(header, 1, DBASE2_HEADER_SIZE, fp) != DBASE2_HEADER_SIZE) {
        printf("ERRO: Cabecalho do dBase II invalido.\n\r");
        fclose(fp);
        return 0;
    }

    if (header[0] != DBASE2_VERSION_BYTE) {
        printf("ERRO: Formato nao e dBASE II (Versao 0x%02X)\n\r", header[0]);
        fclose(fp);
        return 0;
    }

    unsigned short rec_count = header[1] | (header[2] << 8);
    unsigned short rec_size  = header[6] | (header[7] << 8);

    printf("Total de registros na tabela: %d (Tam. Registro: %d bytes)\n\r\n\r", rec_count, rec_size);

    fclose(fp);
    return 1;
}
`

	tmpl, err := template.New("msx_c").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("erro ao compilar template C do MSX: %w", err)
	}

	if opts.TableName == "" {
		opts.TableName = "TABELA"
	}
	if opts.MemoryMapperKB == 0 {
		opts.MemoryMapperKB = 256
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, opts); err != nil {
		return "", fmt.Errorf("erro ao gerar código C: %w", err)
	}

	return buf.String(), nil
}

// GenerateMakefile gera o Makefile / script de compilação SDCC para criar o executável .COM para MSX-DOS.
func GenerateMakefile(opts ProjectOptions) string {
	return fmt.Sprintf(`# Makefile gerado para compilação SDCC do projeto MSX %s
CC = sdcc
CFLAGS = -mz80 --no-std-crt0 --code-loc 0x0100 --data-loc 0x0000
TARGET = %s.COM

all: $(TARGET)

$(TARGET): main.c
	$(CC) $(CFLAGS) main.c -o %s.ihx
	hex2bin -e com %s.ihx
	cp %s.com $(TARGET)
	@echo "Executável MSX-DOS 2 $(TARGET) gerado com sucesso!"

clean:
	rm -f *.ihx *.hex *.bin *.com *.lk *.map *.noi *.rel *.sym *.asm
`, opts.ProjectName, strings.ToUpper(opts.ProjectName), opts.ProjectName, opts.ProjectName, opts.ProjectName)
}
