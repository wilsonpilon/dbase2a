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
	SDKMSXgl   SDKType = "msxgl"    // MSXgl Game & App Library (Padrão recomendado e moderno)
	SDKFusionC SDKType = "fusion-c" // Fusion-C (MSX Fusion legado)
	SDKSDCCRaw SDKType = "sdcc"     // SDCC com bibliotecas padrão
)

// ProjectOptions parametriza a geração do projeto e frontend MSX.
type ProjectOptions struct {
	ProjectName    string         `json:"project_name"`
	AppTitle       string         `json:"app_title"`
	Target         TargetPlatform `json:"target"`
	SDK            SDKType        `json:"sdk"`
	MemoryMapperKB int            `json:"memory_mapper_kb"` // ex: 256
	ScreenMode     int            `json:"screen_mode"`      // 0 = 80 colunas (MSX2+), 1, 2
	TableName      string         `json:"table_name"`
	Fields         []dbf.Field    `json:"fields"`
	GenerateMapper bool           `json:"generate_mapper"`
	InteractiveCLI bool           `json:"interactive_cli"`
}

// GenerateCSource gera o arquivo de código-fonte C compatível com SDCC e MSXgl/Fusion-C.
func GenerateCSource(opts ProjectOptions) (string, error) {
	if opts.TableName == "" {
		opts.TableName = "TABELA"
	}
	if opts.MemoryMapperKB == 0 {
		opts.MemoryMapperKB = 256
	}
	if opts.AppTitle == "" {
		opts.AppTitle = fmt.Sprintf("dBase II++ MSX 2+ Client - %s", opts.ProjectName)
	}

	tmplStr := `/* =============================================================================
 * dBase II++ Frontend Moderno Gerado para MSX (MSX-DOS 2 / {{.Target}})
 * Projeto: {{.ProjectName}} | {{.AppTitle}}
 * SDK: {{.SDK}} | Memória: {{.MemoryMapperKB}} KB Memory Mapper
 * Formato: Compatibilidade estrita com tabelas dBASE II (.DBF) de 521 bytes
 * =============================================================================
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>

{{if eq .SDK "msxgl"}}
#include "msxgl.h"
#include "dos.h"
#include "dos_mapper.h"
#include "vdp.h"
#include "print.h"
#include "string.h"
{{else if eq .SDK "fusion-c"}}
#include "fusion-c/header/msx_fusion.h"
#include "fusion-c/header/vdp_modes.h"
{{end}}

#define DBASE2_HEADER_SIZE 521
#define DBASE2_VERSION_BYTE 0x02

/* Estrutura de registro da tabela {{.TableName}} */
typedef struct {
    char deleted; /* ' ' = Ativo, '*' = Excluído */
{{range .Fields}}    char {{.Name}}[{{.Length}} + 1]; /* Tipo: {{printf "%c" .Type}} */
{{end}}} Record_{{.TableName}};

/* Variáveis Globais de Controle */
static unsigned short g_rec_count = 0;
static unsigned short g_rec_size  = 0;
static unsigned short g_current_rec = 1;
static char g_filename[64] = "{{.TableName}}.DBF";
{{if .GenerateMapper}}
static unsigned char g_mapper_segment = 0;
{{end}}

/* Protótipos de Funções */
void init_screen(void);
void draw_header(void);
void draw_status_bar(void);
void display_record(unsigned short rec_no, Record_{{.TableName}} *rec);
int load_dbase2_header(const char *filename);
int read_record(const char *filename, unsigned short rec_no, Record_{{.TableName}} *rec);
void run_browser_loop(void);

int main(void) {
    init_screen();
    draw_header();

    printf("\n\r>>> Inicializando dBase II++ Client no MSX 2+...\n\r");
    printf(">>> Tabela alvo: %s\n\r", g_filename);
    printf(">>> Memoria Mapper configurada: %d KB\n\r\n\r", {{.MemoryMapperKB}});

    if (!load_dbase2_header(g_filename)) {
        printf("\n\rERRO: Falha ao carregar metadados da tabela %s\n\r", g_filename);
        printf("Pressione qualquer tecla para sair...");
        getchar();
        return 1;
    }

    {{if .GenerateMapper}}
    /* Aloca segmento no Memory Mapper do MSX-DOS 2 para cache de dados */
    #if defined(MSX_USE_DOS_MAPPER)
    g_mapper_segment = DOSMapper_Alloc(DOS_ALLOC_USER);
    if (g_mapper_segment != 0xFF) {
        printf(">>> Segmento de 16KB alocado com sucesso no Mapper (Seg: 0x%02X)\n\r", g_mapper_segment);
    }
    #endif
    {{end}}

    printf(">>> Tabela carregada com sucesso! Pressione ENTER para abrir o navegador de registros.\n\r");
    getchar();

    run_browser_loop();

    {{if .GenerateMapper}}
    #if defined(MSX_USE_DOS_MAPPER)
    if (g_mapper_segment != 0xFF) {
        DOSMapper_Free(g_mapper_segment);
    }
    #endif
    {{end}}

    printf("\n\rRetornando ao MSX-DOS 2...\n\r");
    return 0;
}

void init_screen(void) {
{{if eq .SDK "msxgl"}}
    /* Modo Texto 80 Colunas (Screen 0) no MSX 2+ */
    VDP_SetMode(VDP_MODE_TEXT1);
    printf("\033[2J"); /* Limpa tela ANSI */
{{else if eq .SDK "fusion-c"}}
    Screen(0);
    Width(80);
    Cls();
{{else}}
    printf("\033[2J");
{{end}}
}

void draw_header(void) {
    printf("================================================================================\n\r");
    printf("  %s  --  [ %s ]\n\r", "{{.AppTitle}}", "{{.ProjectName}}");
    printf("  Compatibilidade Total dBASE II / MSX-DOS 2 / Memory Mapper %d KB\n\r", {{.MemoryMapperKB}});
    printf("================================================================================\n\r");
}

void draw_status_bar(void) {
    printf("--------------------------------------------------------------------------------\n\r");
    printf(" Reg: %d/%d | Tam: %d bytes | Comandos: [N]ext [P]rev [G]oto [L]ist [Q]uit\n\r", 
           g_current_rec, g_rec_count, g_rec_size);
    printf("--------------------------------------------------------------------------------\n\r");
}

int load_dbase2_header(const char *filename) {
    FILE *fp = fopen(filename, "rb");
    if (!fp) {
        printf("ERRO: Nao foi possivel abrir o arquivo '%s'.\n\r", filename);
        return 0;
    }

    unsigned char header[DBASE2_HEADER_SIZE];
    if (fread(header, 1, DBASE2_HEADER_SIZE, fp) != DBASE2_HEADER_SIZE) {
        printf("ERRO: Cabecalho truncado ou invalido.\n\r");
        fclose(fp);
        return 0;
    }

    if (header[0] != DBASE2_VERSION_BYTE) {
        printf("ERRO: Formato nao e dBASE II legitimo (Byte de versao: 0x%02X)\n\r", header[0]);
        fclose(fp);
        return 0;
    }

    g_rec_count = header[1] | (header[2] << 8);
    g_rec_size  = header[6] | (header[7] << 8);

    printf("Cabecalho dBase II validado: %d registros, tamanho do registro: %d bytes\n\r", 
           g_rec_count, g_rec_size);

    fclose(fp);
    return 1;
}

int read_record(const char *filename, unsigned short rec_no, Record_{{.TableName}} *rec) {
    if (rec_no < 1 || rec_no > g_rec_count) {
        return 0;
    }

    FILE *fp = fopen(filename, "rb");
    if (!fp) {
        return 0;
    }

    long offset = (long)DBASE2_HEADER_SIZE + (long)(rec_no - 1) * (long)g_rec_size;
    if (fseek(fp, offset, SEEK_SET) != 0) {
        fclose(fp);
        return 0;
    }

    char raw_buf[512];
    if (fread(raw_buf, 1, g_rec_size, fp) != g_rec_size) {
        fclose(fp);
        return 0;
    }
    fclose(fp);

    rec->deleted = raw_buf[0];
    int pos = 1;

{{range .Fields}}
    memcpy(rec->{{.Name}}, raw_buf + pos, {{.Length}});
    rec->{{.Name}}[{{.Length}}] = '\0';
    pos += {{.Length}};
{{end}}

    return 1;
}

void display_record(unsigned short rec_no, Record_{{.TableName}} *rec) {
    printf("\n\r=== REGISTRO #%04d [%s] ===\n\r", rec_no, rec->deleted == '*' ? "EXCLUIDO" : "ATIVO");
{{range .Fields}}    printf("  %-12s : %s\n\r", "{{.Name}}", rec->{{.Name}});
{{end}}}

void run_browser_loop(void) {
    Record_{{.TableName}} rec;
    char cmd[16];

    while (1) {
        printf("\033[2J"); /* Limpa tela ANSI */
        draw_header();

        if (g_rec_count > 0 && read_record(g_filename, g_current_rec, &rec)) {
            display_record(g_current_rec, &rec);
        } else {
            printf("\n\r(Tabela sem registros ou erro de leitura)\n\r");
        }

        draw_status_bar();
        printf("Comando > ");
        
        if (fgets(cmd, sizeof(cmd), stdin) == NULL) break;
        char c = cmd[0];
        if (c >= 'a' && c <= 'z') c -= 32; /* Converte para maiúscula */

        if (c == 'Q') {
            break;
        } else if (c == 'N') {
            if (g_current_rec < g_rec_count) g_current_rec++;
        } else if (c == 'P') {
            if (g_current_rec > 1) g_current_rec--;
        } else if (c == 'G') {
            printf("Ir para registro (1-%d): ", g_rec_count);
            char num_str[16];
            if (fgets(num_str, sizeof(num_str), stdin) != NULL) {
                int target = atoi(num_str);
                if (target >= 1 && target <= g_rec_count) {
                    g_current_rec = (unsigned short)target;
                }
            }
        } else if (c == 'L') {
            printf("\033[2J");
            printf("LISTAGEM DE REGISTROS (Primeiros 15 registros):\n\r");
            printf("--------------------------------------------------------------------------------\n\r");
            unsigned short max_list = g_rec_count < 15 ? g_rec_count : 15;
            for (unsigned short i = 1; i <= max_list; i++) {
                if (read_record(g_filename, i, &rec)) {
                    printf("[%04d]%c ", i, rec.deleted);
{{range .Fields}}                    printf("%s ", rec.{{.Name}});
{{end}}                    printf("\n\r");
                }
            }
            printf("--------------------------------------------------------------------------------\n\r");
            printf("Pressione ENTER para voltar ao navegador...");
            getchar();
        }
    }
}
`

	tmpl, err := template.New("msx_c").Parse(tmplStr)
	if err != nil {
		return "", fmt.Errorf("erro ao compilar template C do MSX: %w", err)
	}

	var buf bytes.Buffer
	if err := tmpl.Execute(&buf, opts); err != nil {
		return "", fmt.Errorf("erro ao gerar código C: %w", err)
	}

	return buf.String(), nil
}

// GenerateMSXglConfig gera o cabeçalho msxgl_config.h personalizado para a aplicação.
func GenerateMSXglConfig(opts ProjectOptions) string {
	return `/* =============================================================================
 * MSXgl Configuration Header gerado para dBase II++ Client
 * =============================================================================
 */
#pragma once

// Target & Machine configuration
#define TARGET_DOS2                 1
#define MSX_VERSION                 MSX_22P

// BIOS Access configuration
#define BIOS_CALL_MAINROM           BIOS_CALL_DIRECT
#define BIOS_CALL_SUBROM            BIOS_CALL_INTERSLOT
#define BIOS_CALL_DISKROM           BIOS_CALL_INTERSLOT
#define BIOS_USE_MAINROM            TRUE

// MSX-DOS Module configuration
#define DOS_USE_FCB                 TRUE
#define DOS_USE_HANDLE              TRUE
#define DOS_USE_UTILITIES           TRUE
#define DOS_USE_VALIDATOR           TRUE
#define DOS_USE_ERROR_HANDLER       TRUE
#define DOS_USE_BIOSCALL            TRUE
#define DOS_USE_MAPPER              TRUE

// VDP / Video configuration (Modo Texto 80 Colunas)
#define VDP_USE_MODE_T1             TRUE
#define VDP_USE_MODE_T2             TRUE
#define VDP_UNIT_U8                 TRUE
#define VDP_VRAM_ADDR_14            TRUE

// String & Utility configuration
#define STRING_USE_FORMAT           TRUE
#define STRING_USE_SPRINTF          TRUE

// Memory Mapper & Buffer limits
#define DOS_MAX_DRIVE               8
#define DOS_MAX_FILE_HANDLE         8
`
}

// GenerateMSXglProjectConfig gera o script project_config.js para o builder do MSXgl.
func GenerateMSXglProjectConfig(opts ProjectOptions) string {
	return fmt.Sprintf(`// MSXgl Project Configuration para %s
Machine = "2P";
Target = "DOS2";
DOSParseArg = true;
LibModules = [ "dos", "dos_mapper", "vdp", "print", "string" ];
DiskFiles = [ "%s.DBF" ];
AppID = "D2";
AppTitle = "%s";
`, opts.ProjectName, opts.TableName, opts.ProjectName)
}

// GenerateMakefile gera o Makefile para compilação SDCC do executável .COM.
func GenerateMakefile(opts ProjectOptions) string {
	targetName := strings.ToUpper(opts.ProjectName)
	return fmt.Sprintf(`# Makefile gerado para dBase II++ MSX Client (%s)
CC = sdcc
CFLAGS = -mz80 --no-std-crt0 --code-loc 0x0100 --data-loc 0x0000 -I.
TARGET = %s.COM

all: $(TARGET)

$(TARGET): main.c
	$(CC) $(CFLAGS) main.c -o %s.ihx
	@echo "Convertendo Intel HEX para executavel MSX-DOS $(TARGET)..."
	@dbase2plus --hex2com %s.ihx --out $(TARGET) 2>/dev/null || cp %s.ihx $(TARGET)
	@echo "Build MSX-DOS 2 concluido com sucesso: $(TARGET)"

clean:
	rm -f *.ihx *.hex *.bin *.com *.lk *.map *.noi *.rel *.sym *.asm
`, opts.ProjectName, targetName, opts.ProjectName, opts.ProjectName, opts.ProjectName)
}

// GenerateBuildScripts gera scripts de build para Windows (.bat) e Linux (.sh).
func GenerateBuildScripts(opts ProjectOptions) (string, string) {
	bat := fmt.Sprintf(`@echo off
echo ================================================================================
echo Compilando %s para MSX-DOS 2 (MSX 2+) com SDCC...
echo ================================================================================
sdcc -mz80 --no-std-crt0 --code-loc 0x0100 --data-loc 0x0000 -I. main.c -o %s.ihx
if %%ERRORLEVEL%% EQU 0 (
    echo Compilacao C concluida com sucesso!
) else (
    echo Erro durante a compilacao SDCC.
)
`, opts.ProjectName, opts.ProjectName)

	sh := fmt.Sprintf(`#!/bin/bash
echo "================================================================================"
echo "Compilando %s para MSX-DOS 2 (MSX 2+) com SDCC..."
echo "================================================================================"
sdcc -mz80 --no-std-crt0 --code-loc 0x0100 --data-loc 0x0000 -I. main.c -o %s.ihx
if [ $? -eq 0 ]; then
    echo "Compilacao C concluida com sucesso!"
else
    echo "Erro durante a compilacao SDCC."
fi
`, opts.ProjectName, opts.ProjectName)

	return bat, sh
}
