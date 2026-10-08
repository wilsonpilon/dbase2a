* Programa.: gerador  vers±o 1.0
* Nota.....: Copyright 1987
* Notas....: Este programa Ç um menu-dirigido para os
*            utilit†rios gerador
* Data.....: 20/01/87
* Palavras reservadas : selecao, selenum
*
SET TALK OFF
SET BELL OFF
SET COLON OFF

DO WHILE T

ERASE
@  1, 0 SAY "========================================"
@  1,40 SAY "========================================"
@  2, 0 SAY "||"
@  2,15 SAY "G E R A  -  M E N U    P R I N C I P A L"
@  2,78 SAY "||"
@  3, 0 SAY "========================================"
@  3,40 SAY "========================================"
@  4, 0 SAY "||"
@  4,78 SAY "||"
@  5, 0 SAY "||"
@  5,78 SAY "||"
@  6, 0 SAY "||"
@  6,78 SAY "||"
@  7, 0 SAY "||"
@  7,78 SAY "||"
@  8, 0 SAY "||"
@  8,78 SAY "||"
@  9, 0 SAY "||"
@  9,78 SAY "||"
@ 10, 0 SAY "||"
@ 10,78 SAY "||"
@ 11, 0 SAY "||"
@ 11,78 SAY "||"
@ 12, 0 SAY "========================================"
@ 12,40 SAY "========================================"
@  5,27 SAY " 0. fim"
@  6,27 SAY " 1. gerador de MENU"
@  7,27 SAY " 2. gerador de ARQUIVO"
@  8,27 SAY " 3. gerador de RELATäRIO"
@  9,27 SAY " 4. gerador de ETIQUETA"
STORE  5 TO selenum
DO WHILE selenum < 0 .OR. selenum >  4
   STORE " " TO selecao
   @ 12,33 SAY " selecione : : "
   @ 12,45 GET selecao PICTURE "#"
   READ
   STORE VAL(selecao) TO selenum
ENDDO

DO CASE
   CASE selenum = 0
      SET COLON ON
      SET BELL ON
      SET TALK ON
      CLEAR
      RETURN
   CASE selenum = 1
   *  DO gerador de MENU
      DO Geramenu
   CASE selenum = 2
   *  DO gerador de ARQUIVO
      DO Geraarq
   CASE selenum = 3
   *  DO gerador de FORMATO DE RELATäRIO
      DO Gerarel
   CASE selenum = 4
   *  DO gerador de ETIQUçTAS
      DO Geraetiq
ENDCASE

* ---Os sets seguintes devem ser resetados, porque os programas Geramenu,
* ---Geraarq, Gerarel, e Geraetiq os modificam para seus valores
* ---default.
SET TALK OFF
SET BELL OFF
SET COLON OFF
?
?
STORE " " TO selecao
@ 22,0 SAY "Aperte qualquer tecla para continuar... " GET selecao
READ

ENDDO T
* EOF: dGEN.CMD
