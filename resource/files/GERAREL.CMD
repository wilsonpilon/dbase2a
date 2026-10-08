clear
SET  ESCAPE OFF
SET  TALK OFF
SET  BELL OFF
STOR ".CMD" TO extensao
STOR "Seu nome " TO seunome
STOR "N" TO s:n
STOR "========================================"+      "========================================" TO iguais
STOR [VAL($(opcao&item,1,@(",",opcao&item)-1))] TO Mlarg 
STOR [opcao&item,@(",",opcao&item)+1,]+      [LEN(opcao&item)-@(",",opcao&item)] TO Mconteudo
ERAS 
@    2, 0 SAY "REPORT  FORM  gerador de relat¢rios"
@    2,72 SAY DATE()
@    3, 0 SAY "========================================"
@    3,40 SAY "========================================"
ACCE "Digite o nome do arquivo de BANCO DE DADOS " TO arqdados
STOR !( TRIM(arqdados) ) + "." TO arqdados
STOR $(arqdados,1,@(".",arqdados)-1) TO arqdados
DO CASE
   CASE arqdados = " "
	ERAS 
	SET  BELL ON
	SET  TALK ON
	SET  ESCAPE ON
	CLEA 
	RETU 
   CASE .NOT. FILE(arqdados + ".DBF")
	?    "*** ARQUIVO N±O EXISTE" 
	SET  BELL ON
	SET  TALK ON
	SET  ESCAPE ON
	CLEA 
	RETU 
ENDCASE
USE &arqdados
ACCE "Digite o nome do arquivo de formato de relat¢rio " TO arqform 
STOR !( TRIM(arqform) ) + "." TO arqform 
STOR $(arqform,1,@(".",arqform)-1) TO arqform
DO CASE
   CASE arqform = " "
	ERAS 
	SET  BELL ON
	SET  TALK ON
	SET  ESCAPE ON
	CLEA 
	RETU 
   CASE FILE(arqform  + extensao)
	STOR "N" TO selecao
	@    7,0 SAY "ARQUIVO DE COMANDOS J„ EXISTE. "+                "Apaga?  (S/N) "            GET selecao PICTURE "!"
	READ 
	@    7,0 SAY "C"
	IF  selecao <> "S"
	   SET  BELL ON
	   SET  TALK ON
	   SET  ESCAPE ON
	   CLEA 
	   RETU 
	ENDI 
ENDCASE
STOR arqform  + extensao  TO arqform 
?    
?    "DIGITE OP€´ES:"
ACCE "   Margem esq....<1>." TO margeme
ACCE "   Linhas/Pag...<56>." TO tampag 
ACCE "   Largura pag..<80>." TO largpag  
IF   VAL(margeme) = 0
   STOR "1" TO margeme
ENDI 
IF   VAL(tampag) = 0
   STOR "56" TO tampag 
ENDI 
IF   VAL(largpag) = 0
   STOR "79" TO largpag
ENDI 
?    
ACCE "Digite Cabe‡alho da p gina." TO cabpag 
ACCE "Totais? (S/N) " TO string
STOR @(string,"Ss")>0 TO temtotal
ACCE "Subtotais no Relat¢rio? (S/N) " TO string
STOR @(string,"Ss") > 0 TO temsubtot 
STOR " " TO totstack,substack
IF   temsubtot 
   STOR 1 TO cont
   STOR " " TO cposubtot
   DO WHILE cposubtot = " " .AND. cont <= 3
      ACCE "Digite campo de subtotal" TO cposubtot
      STOR !(cposubtot) TO cposubtot
      IF   0 = TEST(&cposubtot)
         STOR " " TO cposubtot
      ENDI 
      STOR cont + 1 TO cont
   ENDD 
   IF   cont > 3
      CLEA 
      SET  BELL ON
      SET  TALK ON
      SET  ESCAPE ON
      RETU 
   ENDI 
ENDI 
?    
?    "DIGITE A DESCRI€°O DAS COLUNAS:"
STOR "11" TO item
STOR "X" TO opcao&item
DO WHILE opcao&item <> " " .AND. VAL(item) <= 22
   STOR STR(VAL(item)-10,2)+". Tamanho, Conteudo." TO prompt
   ACCE "&prompt" TO opcao&item
   STOR $(&Mconteudo) TO string
   IF   @(",",opcao&item)>3 .OR. @(",",opcao&item)=0 .OR.      opcao&item=" " .OR. 0=TEST(&string) 
      LOOP 
   ENDI 
   IF   TYPE(&string)="L"
      LOOP 
   ENDI 
   ACCE "    Cabe‡alho......" TO cabec&item
   IF   LEN(cabec&item) > &Mlarg 
      STOR $(cabec&item,1,&Mlarg) TO cabec&item
   ENDI 
   IF   0 > TEST(&string) .AND. temtotal
      ACCE "    Totais? (S/N)" TO s:n
      IF   !(s:n)="S"
         STOR totstack+"&item" TO totstack
      ENDI 
   ENDI 
   IF   0 > TEST(&string) .AND. temsubtot
      ACCE "    Subtotais? (S/N)" TO s:n
      IF   !(s:n)="S"
         STOR substack+"&item" TO substack
      ENDI 
   ENDI 
   ?    
   STOR STR(VAL(item)+1,2) TO item
   STOR "X" TO opcao&item
ENDD 
STOR VAL(item)-1 TO totalopcs
IF   opcao11=" "
   SET  BELL ON
   SET  TALK ON
   SET  ESCAPE ON
   CLEA 
   RETU 
ENDI 
COPY STRUCTURE EXTENDED TO &arqdados..$$$ 
USE &arqdados..$$$
ERAS 
SET  RAW ON
SET  ALTERNATE TO &arqform 
SET  ALTERNATE ON
?    [* Programa.: ]+arqform 
?    [* Autor....: ]+seunome 
?    [* Data.....: ]+DATE()
?    [* Nota.....: Copyright 19]+$(DATE(),7,2)+                [, Direitos Reservados]
?    [* Pal.reser: numpag, linha, cabpag, colcab, condicao,]
?    [*            ultreg]
?    [*]
?    [SET TALK OFF]
?    [SET BELL OFF]
?    [SET MARGIN TO ]+margeme
?    [STORE 1 TO numpag]
?    [STORE 254 TO linha]
?    [STORE "] + cabpag + [" TO cabpag]
?    [STORE (] + largpag + [-LEN(cabpag))/2 TO colcab]
STOR "11" TO item
IF   temtotal .AND. LEN(totstack) <> 1
?    [*]
?    [* ---Inicializa acumuladores.]
   STOR $(totstack,2,LEN(totstack)) TO totstack
   STOR "11" TO item
   STOR 1  TO stackcont
   DO WHILE stackcont < LEN(totstack)
      IF   item=$(totstack,stackcont,2)
?    [STORE 0 TO total&item]
         STOR stackcont+2 TO stackcont
      ENDI 
      STOR STR(VAL(item)+1,2) TO item
   ENDD 
ENDI 
IF   temsubtot  .AND. LEN(substack) <> 1
   STOR $(substack,2,LEN(substack)) TO substack
   STOR "11" TO item
   STOR 1 TO stackcont
   DO WHILE stackcont < LEN(substack)
      IF   item=$(substack,stackcont,2)
?    [STORE 0 TO subtot&item]
         STOR stackcont+2 TO stackcont
      ENDI 
      STOR STR(VAL(item)+1,2) TO item
   ENDD 
ENDI 
?    [*]
?    [* ---Abre o arq. de dados e imprime o relat¢rio.]
?    [USE ]+arqdados
?    [ERASE]
?    [@ 2, 0 SAY cabpag]
?    [@ 2,72 SAY DATE()]
?    [@ 3, 0 SAY "========================================"]
?    [@ 3,40 SAY "========================================"]
?    [STORE " " TO selecao]
?    [@ 5,0 SAY "Saida em tela ou impressora? (T/I) ";]
?    [      GET selecao PICTURE "!"]
?    [READ]
?    [DO CASE]
?    [   CASE selecao = "T"] 
?    [      ERASE]
?    [      STORE 30 TO LINHA]
?    [      STORE 20 TO tampag]
?    [   CASE selecao = "I"]
?    [      SET FORMAT TO PRINT]
?    [      STORE ]+tampag+[ TO tampag]
?    [   OTHERWISE]
?    [      ERASE]
?    [      SET BELL ON]
?    [      SET TALK ON]
?    [      RETURN]
?    [ENDCASE]
?    [* ---Digite FOR <express±o> para o relat¢rio, como por exemplo,]
?    [* ---STORE "ESTADO = 'SP'" TO condicao] 
?    [STORE " " TO condicao]
?    [DO WHILE .NOT. EOF]
?    [   IF linha > tampag]
?    [      IF selecao <>"T"]
?    [         EJECT]
?    [      ENDIF]
?    [      @ 0,0 SAY "PAGINA N."]
?    [      @ 0,9 SAY STR(numpag,3)]
?    [      @ 2,colcab SAY cabpag]
?    [      *]
?    [      * ---Gerador de cabe‡alhos de colunas.]
STOR STR(totalopcs,2) TO colcont
DO WHILE VAL(colcont) >= 11
   STOR "11" TO item
   STOR 0 TO col&colcont
   DO WHILE VAL(colcont) > VAL(item)
      STOR col&colcont + &Mlarg + 1 TO col&colcont
      STOR STR(VAL(item)+1,2) TO item
   ENDD 
   STOR col&colcont + ((VAL(colcont)-11)*2) TO col&colcont
   STOR STR(VAL(colcont)-1,2) TO colcont
ENDD 
STOR "11" TO item
DO WHILE VAL(item) <= totalopcs
?    [      @ 4,]+STR(col&item,3)+[ SAY "]+cabec&item+["]
   STOR STR(VAL(item)+1,2) TO item
ENDD 
STOR "11" TO item
DO WHILE VAL(item) <= totalopcs
?    [      @ 5,]+STR(col&item,3)+[ SAY "]+     $(iguais,1,&Mlarg)+["]
   STOR STR(VAL(item)+1,2) TO item
ENDD 
?    [      STORE  numpag+1 TO numpag]
?    [      STORE 7 TO linha]
?    [   ENDIF]
?    [   * ---Teste para checar se a condicao existe.]
?    [   IF condicao <> " "]
?    [      IF .NOT. (]+"&"+[condicao)]
?    [         SKIP]
?    [         LOOP]
?    [      ENDIF]
?    [   ENDIF]
IF   temsubtot  .AND. LEN(substack) <> 1
?    [   IF 0=TEST(ultreg)]
?    [      * ---Campo nao foi inicializado.]
?    [      STORE ]+cposubtot+[ TO ultreg]
?    [   ENDIF] 
?    [   *]
?    [   * ---Imprimir subtotais e limpar acumuladores.]
?    [   * ---por causa do controle de quebra.]
?    [   IF ultreg  <> ]+cposubtot
   STOR "11" TO item 
   STOR 1  TO stackcont 
?    [      STORE linha + 1 TO linha]
   LOCA FOR Field:name= cposubtot
   IF   Field:type="N"
?    [      @ linha,2 SAY "** SUBTOTAL PARA "]+        [+STR(ultreg,]+STR(Field:len,3)+        [,]+STR(Field:dec,2)+[)+" **"]
   ELSE 
?    [      @ linha,2 SAY "** SUBTOTAL PARA "+TRIM(ultreg)+" **"]
   ENDI 
?    [      STORE linha+1 TO linha]
   DO WHILE stackcont < LEN(substack) 
      IF   item=$(substack,stackcont,2) 
         LOCA FOR Field:name=$(&Mconteudo)
         IF   .NOT. EOF
?    [      @ linha,]+STR(col&item,3)+              [ SAY STR(subtot&item,]+              STR(&Mlarg,3)+[,]+STR(Field:dec,1)+[)]
         ELSE 
?    [      @ linha,]+STR(col&item,3)+[ SAY ]+              [STR(subtot&item,]+STR(&Mlarg,3)+[,2)]
         ENDI 
?    [      STORE 0 TO subtot&item]
         STOR stackcont+2 TO stackcont 
      ENDI 
      STOR STR(VAL(item)+1,2) TO item
   ENDD 
?    [      STORE linha + 2 TO linha]
?    [      STORE ]+cposubtot+[ TO ultreg]
?    [   ENDIF]
ENDI 
?    [   *]
?    [   * ---Imprimir linha de detalhe.]
STOR "11" TO item
DO WHILE VAL(item) <= totalopcs
   STOR $(&Mconteudo) TO string
   STOR &Mlarg TO larg 
   LOCA FOR Field:name=string
   IF   .NOT. EOF
      IF   Field:type="C"
?    [   @ linha,]+STR(col&item,3)+[ SAY ]+           [$(]+string+[,1,]+STR(larg,3)+[)]
      ELSE 
?    [   @ linha,]+STR(col&item,3)+[ SAY ]+           [$(STR(]+string+[,]+STR(larg,3)+           [,]+STR(Field:dec,2)+[),1,]+STR(larg,3)+[)]
      ENDI 
   ELSE 
      USE &arqdados
      IF   0 > TEST(&string)
?    [   @ linha,]+STR(col&item,3)+[ SAY ]+           [$(STR(]+string+[,10,2),1,]+STR(larg,3)+[)]
      ELSE 
?    [   @ linha,]+STR(col&item,3)+[ SAY ]+           [$(]+string+[,1,]+STR(larg,3)+[)]
      ENDI 
      USE &arqdados..$$$
   ENDI 
   STOR STR(VAL(item)+1,2) TO item
ENDD 
?    [   STORE linha + 1 TO linha]
IF   temtotal .AND. LEN(totstack) <> 1
?    [   *]
?    [   * ---Acumular totais e/ou subtotais.]
   STOR "11" TO item
   STOR 1 TO stackcont
   DO WHILE stackcont < LEN(totstack)
      IF   item=$(totstack,stackcont,2)
?    [   STORE total&item+]+$(&Mconteudo)+           [ TO total&item]
         STOR stackcont+2 TO stackcont
      ENDI 
      STOR STR(VAL(item)+1,2) TO item
   ENDD 
ENDI 
IF   temsubtot  .AND. LEN(substack) <> 1
   STOR "11" TO item
   STOR 1 TO stackcont
   DO WHILE stackcont < LEN(substack)
      IF   item=$(substack,stackcont,2)
?    [   STORE subtot&item+]+$(&Mconteudo)+           [ TO subtot&item]
         STOR stackcont+2 TO stackcont
      ENDI 
      STOR STR(VAL(item)+1,2) TO item
   ENDD 
ENDI 
?    [   SKIP]
?    [   IF linha > tampag]
?    [      IF selecao = "T"]
?    [         @ 23,15 say "pressione qualquer tecla para continuar " ]
?    [         wait]
?    [         ERASE]
?    [      endif]
?    [   ENDIF]
?    [ENDDO]
IF   temsubtot  .AND. LEN(substack) <> 1
?    [*]
?    [* ---Imprimir o ultimo registro do subtotal apos o end-of-file.]
?    [STORE linha + 1 TO linha]
   LOCA FOR Field:name= cposubtot
   IF   Field:type="N"
?    [@ linha,2 SAY "** SUBTOTAL PARA "]+[+STR(ultreg,]+        STR(Field:len,3)+[,]+STR(Field:dec,2)+[)+" **"]
   ELSE 
?    [@ linha,2 SAY "** SUBTOTAL PARA "+TRIM(ultreg)+" **"]
   ENDI 
?    [STORE linha + 1 TO linha]
   STOR "11" TO item 
   STOR 1  TO stackcont 
   DO WHILE stackcont < LEN(substack)
      IF   item=$(substack,stackcont,2)
         STOR $(&Mconteudo) TO string
         LOCA FOR Field:name=string
         IF   .NOT. EOF
?    [@ linha,]+STR(col&item,3)+              [ SAY STR(subtot&item,],&Mlarg,[,]+              STR(Field:dec,1)+[)]
         ELSE 
?    [@ linha,]+STR(col&item,3)+[ SAY ]+              [STR(subtot&item,],&Mlarg,[,2)]
         ENDI 
         STOR stackcont+2 TO stackcont 
      ENDI 
      STOR STR(VAL(item)+1,2) TO item 
   ENDD 
ENDI 
IF   temtotal .AND. LEN(totstack) <> 1
?    [*]
?    [* ---Imprimir totais finais.]
   STOR "11" TO item 
   STOR 1 TO stackcont 
?    [STORE linha + 2 TO linha]
?    [@ linha,2 SAY "*** TOTAL FINAL  ***"]
?    [STORE linha+1 TO linha]
   DO WHILE stackcont < LEN(totstack) 
      IF   item = $(totstack,stackcont,2)
         STOR $(&Mconteudo) TO string
         LOCA FOR Field:name=string
         IF   .NOT. EOF
?    [@ linha,]+STR(col&item,3)+              [ SAY STR(total&item,],&Mlarg,[,]+              STR(Field:dec,1)+[)]
         ELSE 
?    [@ linha,]+STR(col&item,3)+              [ SAY STR(total&item,],&Mlarg,[,2)]
         ENDI 
         STOR stackcont+2 TO stackcont 
      ENDI 
      STOR STR(VAL(item)+1,2) TO item 
   ENDD 
ENDI 
?    [@ linha+1,0 SAY " "]
?    [IF selecao = "T"]
?    [   @ 23,15 say "pressione qualquer tecla para continuar " ]
?    [   wait]
?    [   ERASE]
?    [ENDIF]
?    [SET FORMAT TO SCREEN]
?    [RELEASE ALL]
?    [SET TALK ON]
?    [SET BELL ON]
?    [RETURN]
?    [* EOF: ]+arqform 
?    
SET  ALTERNATE OFF
SET  ALTERNATE TO
USE 
DELE FILE &arqdados..$$$
SET  RAW OFF
SET  BELL ON
SET  TALK ON
SET  ESCAPE ON
CLEA 
RETU 
