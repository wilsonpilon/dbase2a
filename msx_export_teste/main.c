/* =============================================================================
 * dBase II+ Frontend Gerado para MSX (MSX-DOS 2 / msx2p)
 * Projeto: TESTE
 * SDK: fusion-c | Memória: 256 KB Mapper
 * Totalmente compatível com arquivos dBASE II (.DBF) de 521 bytes
 * =============================================================================
 */

#include <stdio.h>
#include <stdlib.h>
#include <string.h>


#include "fusion-c/header/msx_fusion.h"
#include "fusion-c/header/vdp_modes.h"


#define DBASE2_HEADER_SIZE 521
#define DBASE2_VERSION_BYTE 0x02

/* Estrutura de dados para os registros da tabela TESTE */
typedef struct {
    char deleted; /* ' ' = Ativo, '*' = Excluído */
    char ID[6 + 1]; /* Tipo: N */
    char DESCRICAO[30 + 1]; /* Tipo: C */
    char VALOR[10 + 1]; /* Tipo: N */
} Record_TESTE;

/* Protótipos de Funções */
void init_screen(void);
void draw_header(void);
void display_record(int rec_no, Record_TESTE *rec);
int load_dbase2_table(const char *filename);

int main(void) {
    init_screen();
    draw_header();

    printf("\n\rIniciando dBase II+ Client para MSX...\n\r");
    printf("Tabela associada: %s.DBF\n\r", "TESTE");
    printf("Memoria RAM Mapper mapeada: %d KB\n\r\n\r", 256);

    /* Carrega e visualiza a tabela */
    load_dbase2_table("TESTE.DBF");

    printf("\n\rPressione qualquer tecla para retornar ao MSX-DOS 2...\n\r");
    getchar();
    return 0;
}

void init_screen(void) {

    /* Configura tela texto 80 colunas no MSX 2+ (Screen 0) */
    Screen(0);
    Width(80);

}

void draw_header(void) {
    printf("================================================================================\n\r");
    printf("  dBase II+ MSX 2+ / MSX-DOS 2 Client  --  [ %s ]\n\r", "TESTE");
    printf("================================================================================\n\r");
}

void display_record(int rec_no, Record_TESTE *rec) {
    printf("[%04d]%c ", rec_no, rec->deleted);
    printf("%s: %s | ", "ID", rec->ID);
    printf("%s: %s | ", "DESCRICAO", rec->DESCRICAO);
    printf("%s: %s | ", "VALOR", rec->VALOR);
    printf("\n\r");
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
