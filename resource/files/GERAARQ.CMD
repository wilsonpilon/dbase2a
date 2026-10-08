clea                 
SET  ESCAPE OFF
SET  TALK OFF
STOR "Seu Nome" TO seunome
STOR ".CMD" TO extensao
ERAS
@    2, 0 SAY "G E R A D O R  (CONSULTA, ADI€°O, EDI€°O E COMPACTA€°O DE DADOS)"
@    2,72 SAY DATE()
@    3, 0 SAY "========================================"
@    3,40 SAY "========================================"
ACCE "Digite o nome do arquivo de BANCO DE DADOS" TO arqent 
STOR !( TRIM(arqent) ) +"." TO arqent
STOR $(arqent,1,@(".",arqent)-1) TO arqent
DO CASE
   CASE arqent = " "
        ERAS 
   	SET TALK ON
   	SET ESCAPE ON
   	CLEA 
   	RETU 
   CASE .NOT. FILE(arqent+".DBF")
	?    "*** ARQUIVO N°O EXISTE" 
	SET  TALK ON
	SET  ESCAPE ON
	CLEA 
	RETU 
ENDCASE
IF   ":" $ arqent
   STOR $(arqent,3,2) TO arqsist  
ELSE 
   STOR $(arqent,1,2) TO arqsist
ENDI 
ACCE "Digite o nome do arquivo de INDICE"  TO arqindice
STOR !( TRIM(arqindice) ) +"." TO arqindice
STOR $(arqindice,1,@(".",arqindice)-1) TO arqindice
IF   arqindice = " "
   ?    "*** DEVE SER DIGITADO UM NOME DE ARQUIVO DE INDICE"
   CLEA 
   SET TALK ON
   SET ESCAPE ON
   RETU 
ENDI 
ACCE "Digite o campo chave de indice " TO cpochave
STOR !( TRIM(cpochave) ) TO cpochave
USE &arqent
IF   0 = TEST(&cpochave)
   ?    "*** CAMPO CHAVE INV„LIDO"
   CLEA 
   SET TALK ON
   SET ESCAPE ON
   RETU 
ENDI 
STOR &cpochave TO valchave
@    8,0 SAY " "
SET BELL OFF
IF   .NOT. FILE(arqindice+".NDX")
   STOR " " TO selecao
   @ $+1,0 SAY "ARQUIVO DE INDICE N°O EXISTE.  Cria?  (S/N) " GET selecao PICTURE "!"
   READ NOUPDATE
   @    $+1,0 SAY " "
   IF  selecao <> "S"
      CLEA 
      SET TALK ON
      SET ESCAPE ON
      RETU 
   ENDI 
   @    $+1,0 SAY "Criando arquivo de indice..."
   INDE ON &cpochave TO &arqindice
ENDI 
IF   ":" $ arqent
   STOR $(arqent,3,1) + " " TO expan
   STOR 4 TO pos
ELSE 
   STOR $(arqent,1,1) + " " TO expan
   STOR 2 TO pos
ENDI 
DO WHILE pos <= LEN(arqent)
   STOR expan + $(arqent,pos,1) + " " TO expan
   STOR pos + 1 TO pos
ENDD 
STOR TRIM(expan) TO expan
@    $+1,0 SAY "Buscando os nomes dos campos..."
COPY STRUCTURE EXTENDED TO &arqent..$$$
USE &arqent..$$$
STOR 0 TO ultlinha
DO WHILE .NOT. EOF .AND. (# <= 14)
   STOR 2 TO cont   
   STOR $(Field:name,1,1) TO field:save
   DO WHILE cont <= 10
      STOR $(Field:name,cont,1) TO caract
      IF caract >= "A" .AND. caract <= "Z"
          STOR CHR( RANK(caract)+32 ) TO caract
      ENDI 
      STOR field:save + caract TO field:save
      STOR cont + 1 TO cont
   ENDD 
   REPL Field:name WITH field:save
   STOR ultlinha + 1 TO ultlinha
   SKIP 
ENDD 
STOR ultlinha + 6 TO ultlinha
GO   TOP
@    $+2,0 SAY "Checando a existˆncia dos nomes de arquivos..."
STOR " " TO selecao
STOR 1 TO cont
DO WHILE cont <= 8
   STOR TRIM( $("PRIN TELASGETS QUALQEDIT PACK PESQ LOCAT",         (cont-1)*5+1,5) ) TO palavra
   STOR arqsist + "-" + palavra + extensao TO nomearq 
   IF   FILE( nomearq  )
      @    $+1,0 SAY ["]+nomearq+["]
      @    $,15 SAY "J„ EXISTE."
      @    $,34 SAY "Apaga? (S/N) "             GET selecao PICTURE "!"
      READ 
      IF  selecao <> "S"
           @    $+1,0 SAY " "
	   SET BELL ON
	   SET TALK ON
	   SET ESCAPE ON
	   CLEA 
	   RETU 
      ENDI 
   ENDI 
   STOR cont + 1 TO cont
ENDD 
SET  BELL ON
SET  RAW ON
ERAS 
STOR arqsist + "-PRIN" + extensao TO arqsai 
SET  ALTERNATE TO &arqsai 
SET  ALTERNATE ON
?    [* Programa.: ]+arqsai 
?    [* Autor....: ]+seunome 
?    [* Data.....: ]+DATE()
?    [* Notas....: Copyright 19]+$(DATE(),7,2)+                [, Direitos Reservados]
?    [* Pal. res.: selecao, selenum, editopc, posopc,]
?    [*            erro, chavepesq, expressao, string, e:eof]
?    [*            limpalin, adiopc, qualq, ultreg]
?    [*]
?    [SET TALK OFF]
?    [SET BELL OFF]
?    [SET COLON OFF]
?    [* ---Use brancos para limpar ate o fim-da-linha.]
?    [STORE $(STR(0,81),1,80) TO limpalin]
?    
?    [DO WHILE T]
?    
?    [ERASE]
?    [@  1, 0 SAY "========================================"]
?    [@  1,40 SAY "========================================"]
?    [@  2, 0 SAY "||"]
STOR expan + "   M E N U   P R I N C I P A L" TO cabec  
STOR (80-LEN(cabec))/2 TO col
?    [@  2,] + STR(col,2) + [ SAY "] + cabec + ["]
?    [@  2,78 SAY "||"]
?    [@  3, 0 SAY "========================================"]
?    [@  3,40 SAY "========================================"]
STOR 4 TO cont
DO WHILE cont <= 11
   ?    [@ ] + STR(cont,2) + [, 0 SAY "||"]
   ?    [@ ] + STR(cont,2) + [,78 SAY "||"]
   STOR cont + 1 TO cont
ENDD 
?    [@ 12, 0 SAY "========================================"]
?    [@ 12,40 SAY "========================================"]
?    [@  5,35 SAY " 0. fim"]
?    [@  6,35 SAY " 1. consultar"]
?    [@  7,35 SAY " 2. adicionar"]
?    [@  8,35 SAY " 3. editar"]
?    [@  9,35 SAY " 4. compactar"]
?    [STORE 5 TO selenum]
?    [DO WHILE selenum < 0 .OR. selenum > 4]
?    [   STORE " " TO selecao]
?    [   @ 12,33 SAY " selecione : : "]
?    [   @ 12,45 GET selecao PICTURE "#"]
?    [   READ]
?    [   STORE VAL(selecao) TO selenum]
?    [ENDDO]
?    
?    [DO CASE]
?    [   CASE selenum= 0]
?    [      CLEAR]
?    [      SET COLON ON]
?    [      SET BELL ON]
?    [      SET TALK ON]
?    [      RETURN]
?    [   CASE selenum= 1]
?    [   *  DO consulta]
?    [      USE ]+arqent+[ INDEX ]+arqindice
?    [      ERASE]
?    [      @ 1, 0 SAY "C O N S U L T A     ]+expan+["]
?    [      @ 1,62 SAY " "]
?    [      @ 1,72 SAY DATE()]
?    [      DO ] + arqsist + [-telas]
?    [      IF # = 0]
?    [         * ---O arquivo de dados esta vazio.]
?    [         STORE " " TO selecao]
?    [         @ ]+STR(ultlinha+1,2)+[,0 SAY "ARQUIVO DE DADOS VAZIO"]
?    [         @ ]+STR(ultlinha+2,2)+[,0 ]+                  [SAY "Pressione qualquer tecla para continuar...";]
?    [                GET selecao]
?    [         READ NOUPDATE]
?    [      ELSE]
?    [         * ---O arquivo de dados contem registros.]
?    [         DO ] + arqsist + [-gets]
?    [         CLEAR GETS]
?    [         STORE "X" TO posopc]
?    [         DO WHILE posopc <> " "]
?    [            DO ] + arqsist + [-pesq]
?    [         ENDDO]
?    [      ENDIF]
?    [      USE]
?    [   CASE selenum= 2]
?    [   *  DO adicao]
?    [      USE ]+arqent+[ INDEX ]+arqindice
?    [      COPY STRUCTURE TO ]+arqent+[.adi]
?    [      SELECT SECONDARY]
?    [      USE ]+arqent+[.adi]
?    [      ERASE]
?    [      @ 1, 0 SAY "A D I C A O     ]+expan+["]
?    [      @ 1,72 SAY DATE()]
?    [      DO ] + arqsist + [-telas]
?    [      @ ]+STR(ultlinha+1,2)+[,0 ]+            [SAY "Pressione <control-W> para sair"]
IF   TYPE(valchave) = "C"
?    [      STORE "X" TO adiopc]
?    [      DO WHILE adiopc <> " "]
?    [         APPEND BLANK]
?    [         DO ] + arqsist + [-gets]
?    [         READ]
?    [         * ---] + cpochave + [ n±o pode ser branco.]
?    [         STORE TRIM( ] + cpochave + [ ) TO adiopc]
?    [      ENDDO]
ELSE 
?    [      STORE 9 TO adiopc]
?    [      DO WHILE adiopc <> 0]
?    [         APPEND BLANK]
?    [         DO ] + arqsist + [-gets]
?    [         READ]
?    [         * ---] + cpochave + [ n±o pode ser zero.]
?    [         STORE ] + cpochave + [ TO adiopc]
?    [      ENDDO]
ENDI 
?    [      DELETE]
?    [      USE]
?    [      SELECT PRIMARY]
?    [      APPEND FROM ]+arqent+[.adi]
?    [      USE]
?    [   CASE selenum= 3]
?    [   *  DO edicao]
?    [      USE ]+arqent+[ INDEX ]+arqindice
?    [      ERASE]
?    [      @ 1, 0 SAY "E D I € ° O   ]+expan+["]
?    [      @ 1,62 SAY " "]
?    [      @ 1,72 SAY DATE()]
?    [      DO ] + arqsist + [-telas]
?    [      IF # = 0]
?    [         * ---O arquivo de dados esta vazio.]
?    [         STORE " " TO selecao]
?    [         @ ]+STR(ultlinha+1,2)+[,0 SAY "ARQUIVO DE DADOS VAZIO"]
?    [         @ ]+STR(ultlinha+2,2)+[,0 ]+                  [SAY "Pressione qualquer tecla para continuar...";]
?    [                GET selecao]
?    [         READ NOUPDATE]
?    [      ELSE]
?    [         * ---O arquivo de dados contem registros.]
?    [         DO ] + arqsist + [-gets]
?    [         CLEAR GETS]
?    [         DO ] + arqsist + [-edit]
?    [      ENDIF]
?    [      USE]
?    [   CASE selenum= 4]
?    [   *  DO compactacao]
?    [      DO ] + arqsist + [-pack]
?    [ENDCASE]
?    
?    [ENDDO T]
?    [* EOF: ] + arqsai 
?    
SET  ALTERNATE OFF
ERAS 
STOR arqsist + "-TELAS" + extensao TO arqsai 
SET  ALTERNATE TO &arqsai 
SET  ALTERNATE ON
?    [* Programa.: ]+arqsai 
?    [* Autor....: ]+seunome  
?    [* Data.....: ]+DATE()
?    [* Nota.....: Copyright 19]+$(DATE(),7,2)+                [, Direitos Reservados]
?    [*]
?    [@  2, 0 SAY "----------------------------------------"]
?    [@  2,40 SAY "----------------------------------------"]
GO   TOP
STOR 4 TO cont   
DO WHILE .NOT. EOF .AND. cont-3 <= 14
   ?    [@ ]+STR(cont,2)+[, 0 SAY "]+     $(TRIM(Field:name)+".........",1,10)+[:"]
   STOR cont + 1 TO cont
   SKIP 
ENDD 
?    [@ ]+STR(ultlinha,2)+[, 0 ]+  [SAY "----------------------------------------"]
?    [@ ]+STR(ultlinha,2)+[,40 ]+  [SAY "----------------------------------------"]
?    [RETURN]
?    [* EOF: ] + arqsai 
?    
SET  ALTERNATE OFF
ERAS 
STOR arqsist + "-GETS" + extensao TO arqsai 
SET  ALTERNATE TO &arqsai 
SET  ALTERNATE ON
?    [* Programa.: ]+arqsai 
?    [* Autor....: ]+seunome 
?    [* Data.....: ]+DATE()
?    [* Nota.....: Copyright 19]+$(DATE(),7,2)+                [, Direitos Reservados]
?    [*]
?    [IF *]
?    [   @ 1,55 SAY "DELETADO"]
?    [ELSE]
?    [   @ 1,55 SAY "        "]
?    [ENDIF]
GO   TOP
STOR 4 TO cont
DO WHILE .NOT. EOF .AND. cont-3 <= 14
   ?    [@ ]+STR(cont,2)+[,11 GET ]+Field:name
   STOR cont + 1 TO cont
   SKIP 
ENDD 
?    [RETURN]
?    [* EOF: ] + arqsai 
?    
SET  ALTERNATE OFF
ERAS 
STOR arqsist + "-QUALQ"+ extensao TO arqsai 
SET  ALTERNATE TO &arqsai 
SET  ALTERNATE ON
?    [* Programa.: ]+arqsai 
?    [* Autor....: ]+seunome  
?    [* Data.....: ]+DATE()
?    [* Nota.....: Copyright 19]+$(DATE(),7,2)+                [, Direitos Reservados]
?    [*]
?    [IF *]
?    [   @ 1,55 SAY "DELETADO"]
?    [ELSE]
?    [   @ 1,55 SAY "        "]
?    [ENDIF]
GO   TOP
STOR 4 TO cont
DO WHILE .NOT. EOF .AND. cont-3 <= 14
   IF   !( TRIM(Field:name) ) $ cpochave 
      ?    [@ ]+STR(cont,2)+[,11 SAY ]+Field:name
   ELSE 
      ?    [@ ]+STR(cont,2)+[,11 GET ]+Field:name
   ENDI 
   STOR cont + 1 TO cont
   SKIP 
ENDD 
?    [RETURN]
?    [* EOF: ] + arqsai   
?    
SET  ALTERNATE OFF
ERAS 
STOR arqsist + "-EDIT" + extensao TO arqsai 
SET  ALTERNATE TO &arqsai 
SET  ALTERNATE ON
?    [* Programa.: ]+arqsai 
?    [* Autor....: ]+seunome 
?    [* Data.....: ]+DATE()
?    [* Nota.....: Copyright 19]+$(DATE(),7,2)+                [, Direitos Reservados]
?    [*]
?    [DO WHILE T]
?    [   STORE " " TO editopc]
?    [   @ ]+STR(ultlinha+1,2)+[,0 ]+            [SAY "COMANDO: (E)ditar, (D)eletar, (R)ecuperar, "+;]
?    [              "(C)ontinuar, (P)osicionar ";]
?    [          GET editopc PICTURE "!"]
?    [   READ NOUPDATE]
?    [   CLEAR GETS]
?    [   @ ]+STR(ultlinha+1,2)+[,0 SAY limpalin]
?    [   DO CASE]
?    [      CASE editopc = " "]
?    [      * ---Sair.]
?    [         RETURN]
?    [      CASE editopc = "P"]
?    [      * ---(P)osicionar]
?    [         STORE "X" TO posopc]
?    [         DO WHILE posopc <> " "]
?    [            DO ] + arqsist + [-pesq]
?    [         ENDDO]
?    [      CASE editopc = "D"]
?    [      * ---(D)eletar]
?    [         DELETE]
?    [         @ 1,55 SAY "DELETADO"]
?    [      CASE editopc = "R"]
?    [      * ---(R)ecuperar]
?    [         RECALL]
?    [         @ 1,55 SAY "        "]
?    [      CASE editopc = "E"]
?    [      * ---(E)ditar]
?    [         @ ]+STR(ultlinha+1,2)+[,0 ]+               [SAY "Pressione <control-W> para sair"]
?    [         IF # <> 0]
?    [            DO ] + arqsist + [-qualq]
?    [            READ]
?    [         ENDIF]
?    [      CASE editopc = "C"]
?    [      * ---(C)ontinua para o proximo registro.]
?    [         STORE # TO ultreg] 
?    [         CONTINUE]
?    [         * ---Checa pelo END-OF-FILE.]
?    [         IF .NOT. EOF]
?    [            DO ] + arqsist + [-gets]
?    [            CLEAR GETS]
?    [         ELSE]
?    [            * ---EOF encontrado.]
?    [            GOTO ultreg]
?    [            @ ]+STR(ultlinha+1,2)+[,0 SAY limpalin]
?    [            @ ]+STR(ultlinha+1,2)+[,0 ]+                  [SAY "END-OF-FILE encontrado"]
?    [            STORE " " TO selecao]
?    [            @ ]+STR(ultlinha+2,2)+[,0 ]+                     [SAY "Pressione qualquer tecla para continuar...";]
?    [                   GET selecao]
?    [            READ NOUPDATE]
?    [            @ ]+STR(ultlinha+1,2)+[,0 SAY limpalin]
?    [            @ ]+STR(ultlinha+2,2)+[,0 SAY limpalin]
?    [         ENDIF]
?    [   ENDCASE]
?    [ENDDO]
?    [* EOF: ] + arqsai 
?    
SET  ALTERNATE OFF
ERAS 
STOR arqsist + "-PACK" + extensao TO arqsai 
SET  ALTERNATE TO &arqsai 
SET  ALTERNATE ON
?    [* Programa.: ]+arqsai 
?    [* Autor....: ]+seunome
?    [* Data.....: ]+DATE()
?    [* Nota.....: Copyright 19]+$(DATE(),7,2)+                [, Direitos Reservados]
?    [*]
?    [ERASE]
?    [@ 2, 0 SAY "C O M P A C T A R    ]+expan+["]
?    [@ 2,72 SAY DATE()]
?    [@ 3, 0 SAY "========================================"]
?    [@ 3,40 SAY "========================================"]
?    [STORE "NAO" TO selecao]
?    '@ 5,0 SAY "COMPACTAR o arquivo todo? [SIM/NAO] ";'
?    [      GET selecao PICTURE "!!!"]
?    [READ NOUPDATE]
?    [IF selecao <> "SIM"]
?    [   RETURN]
?    [ENDIF]
?    [@ 6,0 SAY "]+arqent+[.BKP ser  o backup de seu arquivo de dados."]
?    [IF FILE( "] + arqent + [.BKP" )]
?    [   STORE " " TO selecao]
?    [   @ $+1,0 SAY "Apaga o arquivo de backup antigo? (S/N) ";]
?    [           GET selecao PICTURE "!"]
?    [   READ NOUPDATE]
?    [   IF selecao <> "S"]
?    [      RETURN]
?    [   ENDIF]
?    [   DELETE FILE ]+arqent+[.BKP]
?    [ENDIF]
?    [USE]
?    [RENAME ]+arqent+[.DBF TO ]+arqent+[.BKP]
?    [@ $+1,0 SAY " "]
?    [*]
?    [USE ] + arqent + [.BKP]
?    [SET TALK ON]
?    [SET ECHO ON]
?    [COPY TO ] + arqent
?    [USE]
?    [USE ]+arqent
?    [* ---Recria o arquivo de indice.]
?    [INDEX ON ]+cpochave+[ TO ]+arqindice
?    [USE]
?    [SET ECHO OFF]
?    [SET TALK OFF]
?    [STORE " " TO selecao]
?    [@ 22,0 SAY limpalin]
?    [@ 22,0 SAY "Pressione qualquer tecla para continuar...";]
?    [       GET selecao]
?    [READ NOUPDATE]
?    [RETURN]
?    [* EOF: ] + arqsai 
?    
SET  ALTERNATE OFF
ERAS 
STOR arqsist + "-PESQ" + extensao TO arqsai 
SET  ALTERNATE TO &arqsai 
SET  ALTERNATE ON
?    [* Programa.: ]+arqsai 
?    [* Autor....: ]+seunome 
?    [* Data.....: ]+DATE()
?    [* Nota.....: Copyright 19]+$(DATE(),7,2)+                [, Direitos Reservados]
?    [*]
?    [STORE " " TO posopc]
?    [@ ]+STR(ultlinha+1,2)+[,0 SAY limpalin]
?    [@ ]+STR(ultlinha+1,2)+      [,0 SAY "COMANDO: (M)ostrar, (P)esquisar, "+;]
?    [           "(L)ocalizar, (C)ontinuar, (A)van‡ar ";]
?    [       GET posopc PICTURE "!"]
?    [READ NOUPDATE]
?    [CLEAR GETS]
?    [@ ]+STR(ultlinha+1,2)+[,0 SAY limpalin]
?    [IF .NOT. (posopc $ "MPLCA")]
?    [   RETURN]
?    [ENDIF]
?    [IF posopc = "P"]
?    [   * ---(P)esquisar]
?    [   @ ]+STR(ultlinha,2)+[,0 SAY "-"]
?    [   ACCEPT "Digite ]+cpochave+[ " TO chavepesq]
?    [   @ ]+STR(ultlinha+1,2)+[,0 SAY limpalin]
?    [   STORE TRIM(chavepesq) TO chavepesq]
?    [   IF chavepesq = " "]
?    [      RETURN]
?    [   ENDIF]
?    [   STORE # TO ultreg]
?    [   FIND &chavepesq]
?    [   IF (# <> 0)]
?    [      DO ] + arqsist + [-gets]
?    [      CLEAR GETS]
?    [   ELSE]
?    [      * ---NAO ENCONTROU.]
?    [      GOTO ultreg]
?    [      @ ]+STR(ultlinha+1,2)+[,0 SAY limpalin]
?    [      @ ]+STR(ultlinha+1,2)+            [,0 SAY '"'+chavepesq+'"'+" n±o est  no ¡ndice"]
?    [      STORE " " TO selecao]
?    [      @ ]+STR(ultlinha+2,2)+            [,0 SAY "Pressione qualquer tecla para continuar...";]
?    [             GET selecao]
?    [      READ NOUPDATE]
?    [      @ ]+STR(ultlinha+1,2)+[,0 SAY limpalin]
?    [      @ ]+STR(ultlinha+2,2)+[,0 SAY limpalin]
?    [   ENDIF]
?    [ELSE]
?    [   * ---(A)van‡ar, (C)ontinuar, (M)ostrar, ou (L)ocalizar]
?    [   STORE # TO ultreg]
?    [   DO CASE]
?    [      CASE posopc = "A"]
?    [      * ---(A)vancar.]
?    [         SKIP]
?    [      CASE posopc = "C"]
?    [      * ---(C)ontinuar.]
?    [         CONTINUE]
?    [      OTHERWISE]
?    [      * ---(M)ostrar ou (L)ocalizar]
?    [         DO ] + arqsist + [-locat]
?    [         IF expressao = " "]
?    [            RETURN]
?    [         ENDIF]
?    [         IF posopc = "M"]
?    [            IF string = " "]
?    [               RETURN]
?    [            ENDIF]
?    [            DO ] + arqsist + [-telas]
?    [         ENDIF]
?    [   ENDCASE]
?    [   * ---Checa por END-OF-FILE.]
?    [   IF .NOT. EOF]
?    [      DO ] + arqsist + [-gets]
?    [      CLEAR GETS]
?    [   ELSE]
?    [      * ---EOF encontrado.]
?    [      GOTO ultreg]
?    [      @ ]+STR(ultlinha+1,2)+[,0 SAY limpalin]
?    [      @ ]+STR(ultlinha+1,2)+[,0 SAY "END-OF-FILE encontrado"]
?    [      STORE " " TO selecao]
?    [      @ ]+STR(ultlinha+2,2)+            [,0 SAY "Pressione qualquer tecla para continuar...";]
?    [             GET selecao]
?    [      READ NOUPDATE]
?    [      @ ]+STR(ultlinha+1,2)+[,0 SAY limpalin]
?    [      @ ]+STR(ultlinha+2,2)+[,0 SAY limpalin]
?    [   ENDIF]
?    [ENDIF]
?    [RETURN]
?    [* EOF: ] + arqsai 
?    
ERAS 
STOR arqsist + "-LOCAT" + extensao TO arqsai 
SET  ALTERNATE TO &arqsai 
SET  ALTERNATE ON
?    [* Programa.: ]+arqsai 
?    [* Autor....: ]+seunome 
?    [* Data.....: ]+DATE()
?    [* Nota.....: Copyright 19]+$(DATE(),7,2)+                [, Direitos Reservados]
?    [*]
?    [STORE " " TO expressao,string]
?    [* ---LOCALIZACAO.]
?    [DO WHILE expressao = " "]
?    [   @ ]+STR(ultlinha-1,2)+[,0 SAY 'EXEMPLO :ESTADO="SP"']
?    [   @ ]+STR(ultlinha+1,2)+[,0 SAY limpalin]
?    [   *]
?    [   @ ]+STR(ultlinha,2)+[,0 SAY "-"]
?    [   ACCEPT "Digite expressao para LOCALIZACAO " TO expressao]
?    [   @ ]+STR(ultlinha-1,2)+[,0 SAY limpalin]
?    [   STORE TRIM(expressao) TO expressao]
?    [   DO CASE]
?    [      CASE expressao = " "]
?    [      * ---Sair.]
?    [         RETURN]
?    [      CASE 0 = TEST(&expressao)]
?    [      * ---EXPRESSAO INVALIDA.]
?    [         @ ]+STR(ultlinha+2,2)+[,0 SAY limpalin]
?    [         STORE " " TO selecao]
?    [         @ ]+STR(ultlinha+2,2)+               [,0 SAY "EXPRESS°O INV„LIDA: "+;]
?    [                    "Pressione qualquer tecla para continuar... ";]
?    [                GET selecao]
?    [         READ NOUPDATE]
?    [         @ ]+STR(ultlinha+2,2)+[,0 SAY limpalin]
?    [         STORE " " TO expressao]
?    [      OTHERWISE]
?    [      * ---LOCALIZAR o registro.]
?    [         * ---Fecha o arquivo de indice para um LOCATE mais rapido.]
?    [         SET INDEX TO]
?    [         LOCATE FOR &expressao]
?    [         IF .NOT. EOF]
?    [            * ---Encontra um registro que satisfaz a condicao.]
?    [            STORE # TO ultreg]
?    [            * ---Reabre o arquivo de indice.]
?    [            SET INDEX TO ] + arqindice
?    [            GOTO ultreg]
?    [         ELSE]
?    [            * ---Reabre o arquivo de indice e]
?    [            * ---forca a marca do EOF para verdadeiro (.T.).]
?    [            SET INDEX TO ] + arqindice
?    [            GO BOTTOM]
?    [            SKIP]
?    [         ENDIF]
?    [   ENDCASE]
?    [ENDDO]
?    [IF posopc = "L"]
?    [   * ---Retorna para o programa chamador]+           [se somente o LOCATE foi desejado.]
?    [   RETURN]
?    [ENDIF]
?    [*]
?    [* ---MOSTRAR.]
?    [STORE " " TO string]
?    [DO WHILE string = " "]
?    [   @ ]+STR(ultlinha-1,2)+         [,0 SAY "EXEMPLO :NOME+ENDERECO"]
?    [   @ ]+STR(ultlinha+1,2)+[,0 SAY "E"]
?    [   ACCEPT "Digite a string a ser mostrada " TO string]
?    [   @ ]+STR(ultlinha-1,2)+[,0 SAY limpalin]
?    [   STORE TRIM(string) TO string]
?    [   DO CASE]
?    [      CASE string = " "]
?    [      * ---Sair.]
?    [         @ ]+STR(ultlinha+1,2)+[,0 SAY limpalin]
?    [         @ ]+STR(ultlinha+2,2)+[,0 SAY limpalin]
?    [         RETURN]
?    [      CASE 0 = TEST(&string)]
?    [      * ---EXPRESSAO INVALIDA.]
?    [         @ ]+STR(ultlinha+2,2)+[,0 SAY limpalin]
?    [         STORE " " TO selecao]
?    [         @ ]+STR(ultlinha+2,2)+               [,0 SAY "EXPRESS°O PARA DISPLAY INV„LIDA: "+;]
?    [                    "
Pressione qualquer tecla para continuar... ";]
?    [                GET selecao]
?    [         READ NOUPDATE]
?    [         @ ]+STR(ultlinha+2,2)+[,0 SAY limpalin]
?    [         STORE " " TO string]
?    [   ENDCASE]
?    [ENDDO]
?    [* ---Agora, o DISPLAY da expressao.]
?    [STORE F TO e:eof,qualq]
?    [DO WHILE .NOT. e:eof]
?    [   * ---O conjunto seguinte de comandos dBASE II, sao para]
?    [   * ---limpar ate o fim da tela. Se voce tem um IBM-PC,]
?    [   * ---voce pode trocar estes comandos por um unico:]
?    [   * ---@ 4,0 ERASE]
?    [   *]
STOR 4 TO lin
DO WHILE lin <= 22
   ?    [   @ ]+STR(lin,2)+[,0 SAY limpalin]
   STOR lin + 1 TO lin
ENDD 
?    [   @ 20, 0 SAY "----------------------------------------"]
?    [   @ 20,40 SAY "----------------------------------------"]
?    [   STORE 4 TO lin]
?    [   DO WHILE .NOT. EOF .AND. lin-3 <= 15]
?    [      STORE T TO qualq]
?    [      @ lin,0 SAY ]+[&]+[string]
?    [      STORE lin + 1 TO lin]
?    [      CONTINUE]
?    [   ENDDO]
?    [   * ---Uma variavel de memoria logica e usada para    ]
?    [   * ---detectar o fim de arquivo. A funcao EOF nao pode ser]
?    [   * ---resetado devido ao comando READ abaixo.]
?    [   STORE EOF TO e:eof]
?    [   IF .NOT. qualq]
?    [      * ---Registros nao satisfazem a condicao.]
?    [      @ 4,0 SAY "*** N°O H„ REGISTRO QUE SATISFA€A A CONDI€°O ***"]
?    [   ENDIF]
?    [   STORE " " TO selecao]
?    [   @ 21,0 SAY "Pressione qualquer tecla para continuar... ";]
?    [          GET selecao]
?    [   READ NOUPDATE]
?    [   @ 21,0 SAY limpalin]
?    [ENDDO]
?    [* ---O conjunto seguinte de comandos do dBASE II sao para]
?    [* ---limpar ate o fim da tela. Se voce tem um IBM-PC,]
?    [* ---voce pode trocar estes comandos por um unico:]
?    [* ---@ 4,0 ERASE]
?    [*]
STOR 4 TO lin
DO WHILE lin <= 22
   ?    [@ ]+STR(lin,2)+[,0 SAY limpalin]
   STOR lin + 1 TO lin
ENDD 
?    [GO TOP]
?    [RETURN]
?    [* EOF: ] + arqsai 
?    
SET ALTERNATE OFF
SET  ALTERNATE TO
USE  
DELE FILE &arqent..$$$
ERAS 
?    [PARA INICIAR O SISTEMA "]+arqent+["  , DIGITE O SEGUINTE:]
?    [.]
?    [.]
?    [. DO ]+arqsist+[-PRIN]
?    [.]
CLEA 
SET  RAW OFF
SET  TALK ON
SET  ESCAPE ON
RETU 

