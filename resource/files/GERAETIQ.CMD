SET  ESCAPE OFF
SET  TALK OFF
SET  BELL OFF
STOR "Seu Nome" TO seunome
STOR ".CMD" TO extensao
ERAS
@    2, 0 SAY "G E R A D O R  D E  E T I Q U E T A S"
@    2,72 SAY DATE()
@    3, 0 SAY "========================================"
@    3,40 SAY "========================================"
ACCE "Digite o nome do arquivo de BANCO DE DADOS" TO arqent 
STOR !( TRIM(ARQENT) ) + "." TO arqent
STOR $(arqent,1,@(".",arqent)-1) TO arqent
DO CASE
   CASE arqent = " "
       ERAS 
       SET  BELL ON
       SET  TALK ON
       SET  ESCAPE ON
       CLEA 
       RETU 
   CASE .NOT. FILE(arqent+".DBF")
          ?    "*** ARQUIVO N°O EXISTE"
          SET  BELL ON
          SET  TALK ON
	  SET  ESCAPE ON
	  CLEA 
	  RETU 
ENDCASE
USE &arqent
ACCE "Digite o nome do arquivo de ETIQUETAS " TO arqsai 
STOR !( TRIM(arqsai) ) + "." TO arqsai  
STOR $(arqsai,1,@(".",arqsai)-1) TO arqsai 
DO CASE
   CASE arqsai  = " "
        ERAS 
	SET  BELL ON
	SET  TALK ON
	SET  ESCAPE ON
	CLEA 
	RETU 
   CASE FILE(arqsai + extensao)
	STOR "N" TO selecao
	@    7,0 SAY "ARQUIVO DE COMANDOS J„ EXISTE. "+                "Apaga-o? (S/N) "            GET selecao PICTURE "!"
	READ 
	@    7,0 SAY "C"
	IF   selecao <> "S"
	   SET  BELL ON
	   SET  TALK ON
	   SET  ESCAPE ON
	   CLEA 
	   RETU 
        ENDI 
ENDCASE
STOR arqsai + extensao TO arqsai
STOR $(arqent,1,1) + " " TO expan
STOR 2 TO pos
DO WHILE pos <= LEN(arqent)
   STOR expan + $(arqent,pos,1) + " " TO expan
   STOR pos + 1 TO pos
ENDD 
STOR TRIM(expan) TO expan
?    
?    "Digite as linhas da ETIQUETA (exemplo):"
?    " 1:NOME"
?    " 2:[COMPANHIA]      <--- (imprimir  somente quando n±o-branco)"
?    " 3:ENDERE€O"                            
?    [ 4:TRIM(CIDADE)+", "+ESTADO+"  "+CEP]
?    " 5:"
?    
?    "Digite as linhas das ETIQUETAS:"
STOR "11" TO item
STOR "X" TO linha&item
DO WHILE linha&item <> " " .AND. VAL(item)-10 <= 10
   STOR STR(VAL(item)-10,2) TO prompt
   ACCE "&prompt" TO linha&item
   STOR TRIM( linha&item ) TO string
   DO WHILE  $(string,1,1) = " " .AND. LEN(string) > 1
      STOR $(string,2) TO string
   ENDD 
   IF  linha&item = " " .OR. 0 = TEST(&string)
      LOOP 
   ENDI 
   STOR STR(VAL(item)+1,2) TO item
   STOR "X" TO linha&item
ENDD 
STOR VAL(item)-11 TO contlin
IF  linha11=" "
   SET  BELL ON
   SET  TALK ON
   SET  ESCAPE ON
   CLEA 
   RETU 
ENDI 
ERAS 
SET  ALTERNATE TO &arqsai 
SET  ALTERNATE ON
SET  RAW ON
?    [* Programa.: ]+arqsai
?    [* Autor....: ]+seunome 
?    [* Data.....: ]+DATE()
?    [* Nota.....: Copyright 19]+$(DATE(),7,2)+                [, Direitos Reservados]
?    [* Pal.reser: selecao, condicao, extra]
?    [*]
?    [SET TALK OFF]
?    [SET BELL OFF]
?    [STORE " " TO selecao]
?    [USE ]+arqent
?    [ERASE]
?    [@ 2, 0 SAY "]+expan+[    M A L A   D I R E T A       "]
?    [@ 2,72 SAY DATE()]
?    [@ 3, 0 SAY "========================================"]
?    [@ 3,40 SAY "========================================"]
?    [STORE " " TO selecao]
?    '@ 5,0 SAY "Saida para tela ou impressora? [T/I] ";'
?    [      GET selecao PICTURE "!"]
?    [READ]
?    [DO CASE]
?    [   CASE selecao= "T"]
?    [      ERASE]
?    [   CASE selecao= "I"]
?    [      SET CONSOLE OFF]
?    [      SET PRINT ON]
?    [   OTHERWISE]
?    [      ERASE]
?    [      SET BELL ON]
?    [      SET TALK ON]
?    [      RETURN]
?    [ENDCASE]
?    [* ---Digite FOR <expressao> para as etiquetas, como por exemplo,]
?    [* ---STORE "ESTADO = 'SP'" TO condicao]
?    [STORE " " TO condicao]
?    [DO WHILE .NOT. EOF]
?    [   IF condicao <> " "]
?    [      IF .NOT. (&condicao)]
?    [         SKIP]
?    [         LOOP]
?    [      ENDIF]
?    [   ENDIF]
?    [   STORE 0 TO extra]
STOR "11" TO item
DO WHILE VAL(item)-10 <= contlin
   STOR TRIM(linha&item ) TO string
   IF   $(string,1,1) = "[" .AND.      $(string,LEN(string),1) = "]" 
      STOR $(string,2,LEN(string)-2) TO string
?    [   IF TRIM(]+string+[) = " "]
?    [      STORE extra + 1 TO extra]
?    [   ELSE]
?    [      ? ]+string
?    [   ENDIF]
   ELSE 
?    [   ? ]+string
   ENDI 
   STOR STR(VAL(item)+1,2) TO item
ENDD 
?    [   ?]
?    [   ?]
?    [   DO WHILE extra > 0]
?    [      ?]
?    [      STORE extra - 1 TO extra]
?    [   ENDDO]
?    [   SKIP]
?    [ENDDO]
?    [*]
?    [SET PRINT OFF]
?    [SET CONSOLE ON]
?    [?]
?    [? "ISTO E TUDO..."]
?    [CLEAR]
?    [SET TALK ON]
?    [SET BELL ON]
?    [RETURN]
?    [* EOF: ] + arqsai
?    
SET  ALTERNATE OFF
SET  ALTERNATE TO
CLEA 
SET  RAW OFF
SET  BELL ON
SET  TALK ON
SET  ESCAPE ON
RETU 
