CLEA
SET  ESCAPE OFF
SET  TALK OFF
SET  BELL OFF
SET  COLON OFF
STOR ".CMD" TO extensao
STOR "Seu Nome" TO seunome
STOR " " TO escolha
DO WHILE escolha <> "G"
   ERAS 
   @    2, 0 SAY "G E R A D O R   D E   M E N U"
   @    2,72 SAY DATE()
   @    3, 0 SAY "========================================"
   @    3,40 SAY "========================================"
   ACCE "Digite o nome do PROGRAMA " TO arqsai
   STOR !( TRIM(arqsai) ) + "." TO arqsai 
   STOR $(arqsai,1,@(".",arqsai)-1) TO arqsai 
   DO CASE
      CASE arqsai  = " "
           ERAS 
           SET  BELL ON
           SET  TALK ON
           SET  ESCAPE ON
           SET  COLON ON
           CLEA 
           RETU 
      CASE FILE(arqsai + extensao)
           STOR "N" TO selecao
           @ 6,0 SAY "ARQUIVO DE COMANDOS J„ EXISTE. "+                    "Apaga-o? [S/N] "                GET selecao PICTURE "!"
	   READ 
           @    $,0 SAY "C"
           IF   selecao <> "S"
	      SET  BELL ON
              SET  TALK ON
	      SET  ESCAPE ON
              SET  COLON ON
              CLEA 
              RETU 
           ENDI 
   ENDCASE
   STOR arqsai + extensao TO arqsai
   ACCE "Digite o cabe‡alho do MENU " TO cabec  
   IF   cabec = " "
      SET  BELL ON
      SET  TALK ON
      SET  ESCAPE ON
      SET  COLON ON
      CLEA 
      RETU 
   ENDI 
   STOR TRIM( !(cabec) ) TO cabec
   STOR " " TO expan
   SET  EXACT ON
   DO WHILE cabec <> " "
      STOR @(" ",cabec) TO pos
      IF   pos > 0
         STOR $(cabec,1,pos-1) TO palavra
         STOR $(cabec,pos+1,LEN(cabec)-pos) TO cabec
      ELSE 
         STOR cabec TO palavra
         STOR " " TO cabec  
      ENDI 
      STOR 1 TO item
      DO WHILE item <= LEN(palavra)
         STOR expan+$(palavra,item,1)+" " TO expan
         STOR item+1 TO item
      ENDD 
      STOR expan+"   " TO expan
   ENDD 
   STOR TRIM(expan) TO expan
   STOR $(expan,2,LEN(expan)) TO expan
   STOR (80-LEN(expan))/2 TO colcab 
   SET EXACT OFF
   ?    
   ?    "Digite as op‡µes do menu:"
   ?    "   0:sair"
   STOR "sair" to opcao10
   STOR 9 TO tam
   STOR "11" TO item
   STOR "X" TO opcao&item
   DO WHILE opcao&item <> " " .AND. VAL(item)-10 <= 14
      STOR STR(VAL(item)-10,2) TO prompt
      ACCE "  &prompt" TO opcao&item
      IF   opcao&item <> " "
         IF   tam < LEN(opcao&item)+5
            STOR LEN(opcao&item)+5 TO tam
         ENDI 
         STOR STR(VAL(item)+1,2) TO item
         STOR "X" TO opcao&item
      ENDI 
   ENDD 
   STOR VAL(item)-11 TO totalopcs
   IF   opcao11=" "
      SET  BELL ON
      SET  TALK ON
      SET  ESCAPE ON
      SET  COLON ON
      CLEA 
      RETU 
   ENDI 
   STOR (80-tam)/2 TO colopcs 
   ERAS 
   @    1, 0 SAY "========================================"
   @    1,40 SAY "========================================"
   @    2, 0 SAY "||"
   @    2,colcab  SAY expan
   @    2,78 SAY "||"
   @    3, 0 SAY "========================================"
   @    3,40 SAY "========================================"
   STOR 4 TO cont
   DO WHILE cont <= totalopcs + 7
      @    cont,0  SAY "||"
      @    cont,78 SAY "||"
      STOR cont+1 TO cont
   ENDD 
   @    cont, 0 SAY "========================================"
   @    cont,40 SAY "========================================"
   STOR "10" TO item
   DO WHILE VAL(item)-10 <= totalopcs
      @    VAL(item)-5, colopcs  SAY            STR(VAL(item)-10,2)+". "+opcao&item
      STOR STR(VAL(item)+1,2) TO item
   ENDD 
   STOR " " TO escolha
   DO WHILE .NOT. (escolha $ "SRG")
      STOR " " TO escolha
      @    22,0  SAY "COMANDO: (S)air, (R)efazer, (G)ravar"              GET escolha PICTURE "!"
      READ 
   ENDD 
   IF escolha <>"G"
      SET  BELL ON
      SET  TALK ON
      SET  ESCAPE ON
      SET  COLON ON
      CLEA 
      RETU 
   ENDI 
ENDD opcao <> "G"
SET  RAW ON
ERAS 
SET  ALTERNATE TO &arqsai 
SET  ALTERNATE ON
?    [* Programa.: ]+arqsai 
?    [* Autor....: ]+seunome 
?    [* Data.....: ]+DATE()
?    [* Nota.....: Copyright 19]+$(DATE(),7,2)+                [, Direitos Reservados]
?    [* Observa‡µes.:]
?    [* Pal.reser: selecao, selenum]
?    [*]
?    [SET TALK OFF]
?    [SET BELL OFF]
?    [SET COLON ON]
?    
?    [DO WHILE T]
?    [set talk off]
?    [ERASE]
?    [@  1, 0 SAY "========================================"]
?    [@  1,40 SAY "========================================"]
?    [@  2, 0 SAY "||"]
?    [@  2,] + STR(colcab,2) + [ SAY "] + expan + ["]
?    [@  2,78 SAY "||"]
?    [@  3, 0 SAY "========================================"]
?    [@  3,40 SAY "========================================"]
STOR 4 TO cont
DO WHILE cont <= totalopcs + 7
?    [@ ] + STR(cont,2) + [, 0 SAY "||"]
?    [@ ] + STR(cont,2) + [,78 SAY "||"]
   STOR cont + 1 to cont 
ENDD 
?    [@ ] + STR(cont,2) +  [, 0 SAY "========================================"]
?    [@ ] + STR(cont,2) +  [,40 SAY "========================================"]
STOR "10" TO item
DO WHILE VAL(item)-10 <= totalopcs
?    [@ ]+STR(VAL(item)-5,2)+[,]+STR(colopcs,2)+[ SAY ]+     ["]+STR(VAL(item)-10,2)+[. ] + opcao&item + ["]
   STOR STR(VAL(item)+1,2) TO item
ENDD 
?    [STORE ]+STR(totalopcs+1,2)+[ TO selenum]
?    [DO WHILE selenum < 0 .OR. selenum > ]+STR(totalopcs,2)
IF   totalopcs < 10
?    [   STORE " " TO selecao]
?    [   @ ]+STR(totalopcs+8,2)+[,33 SAY " selecione     "]
?    [   @ ]+STR(totalopcs+8,2)+[,44 GET selecao PICTURE "#"]
ELSE 
?    [   STORE "  " TO selecao]
?    [   @ ]+STR(totalopcs+8,2)+[,33 SAY " selecione :  : "]
?    [   @ ]+STR(totalopcs+8,2)+[,45 GET selecao PICTURE "##"]
ENDI 
?    [   READ]
?    [   STORE VAL(selecao) TO selenum]
?    [ENDDO]
?    
?    [DO CASE]
?    [   CASE selenum= 0]
?    [      SET COLON ON]
?    [      SET BELL ON]
?    [      SET TALK ON]
?    [      CLEAR]
?    [      RETURN]
STOR "11" TO item
DO WHILE VAL(item)-10 <= totalopcs
?    [   CASE selenum=] + STR(VAL(item)-10,2)
?    [   *  DO ]+opcao&item
   STOR STR(VAL(item)+1,2) TO item
ENDD 
?    [ENDCASE]
?    
?    [ENDDO T]
?    [* EOF: ]+arqsai 
?    
SET  ALTERNATE OFF
SET  ALTERNATE TO
SET  COLON ON
SET  BELL ON
SET  TALK ON
SET  ESCAPE ON
SET  RAW OFF
CLEA 
RETU 
