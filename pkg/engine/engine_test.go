package engine

import (
	"bytes"
	"path/filepath"
	"strings"
	"testing"

	"dbase2a/pkg/config"
	"dbase2a/pkg/dbf"
)

func TestEngineFullExecutionCycle(t *testing.T) {
	tempDir := t.TempDir()
	dbfPath := filepath.Join(tempDir, "LIVROS.DBF")

	cfg := config.DefaultConfig()
	cfg.Talk = false // Silencia para testar saída limpa
	buf := new(bytes.Buffer)

	eng := NewEngine(cfg, nil, buf)

	// 1. Criação manual da tabela
	fields := []dbf.Field{
		{Name: "ID", Type: dbf.TypeNumeric, Length: 4, Decimals: 0},
		{Name: "TITULO", Type: dbf.TypeCharacter, Length: 25, Decimals: 0},
		{Name: "PRECO", Type: dbf.TypeNumeric, Length: 8, Decimals: 2},
	}
	tbl, err := dbf.CreateTable(dbfPath, fields)
	if err != nil {
		t.Fatalf("Erro ao criar tabela: %v", err)
	}

	r1 := tbl.AppendBlank()
	r1.Set("ID", "1")
	r1.Set("TITULO", "DRACULA")
	r1.Set("PRECO", "45.00")

	r2 := tbl.AppendBlank()
	r2.Set("ID", "2")
	r2.Set("TITULO", "FRANKENSTEIN")
	r2.Set("PRECO", "50.00")

	_ = tbl.Save()

	// 2. Executa USE
	err = eng.ExecuteLine("USE " + dbfPath)
	if err != nil {
		t.Fatalf("Erro no USE: %v", err)
	}

	// 3. Executa LIST
	buf.Reset()
	err = eng.ExecuteLine("LIST ALL")
	if err != nil {
		t.Fatalf("Erro no LIST: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "DRACULA") || !strings.Contains(out, "FRANKENSTEIN") {
		t.Errorf("LIST ALL não retornou os registros esperados: %s", out)
	}

	// 4. Executa COUNT e SUM
	buf.Reset()
	_ = eng.ExecuteLine("COUNT FOR PRECO > 40")
	if !strings.Contains(buf.String(), "2 REGISTROS") {
		t.Errorf("COUNT incorreto: %s", buf.String())
	}

	buf.Reset()
	_ = eng.ExecuteLine("SUM PRECO")
	if !strings.Contains(buf.String(), "95.00") {
		t.Errorf("SUM incorreto: %s", buf.String())
	}

	// 5. Executa REPLACE e LOCATE
	_ = eng.ExecuteLine("REPLACE TITULO WITH NOSFERATU FOR ID = 1")
	if eng.ActiveTable.Records[0].Get("TITULO") != "NOSFERATU" {
		t.Errorf("REPLACE falhou, valor é %s", eng.ActiveTable.Records[0].Get("TITULO"))
	}

	// 6. Executa DELETE e PACK
	_ = eng.ExecuteLine("DELETE ALL FOR ID = 2")
	_ = eng.ExecuteLine("PACK")
	if len(eng.ActiveTable.Records) != 1 {
		t.Errorf("Após PACK esperava 1 registro, obteve %d", len(eng.ActiveTable.Records))
	}

	// 7. QUIT
	_ = eng.ExecuteLine("QUIT")
	if !eng.QuitRequested {
		t.Errorf("QUIT deveria ter setado QuitRequested como true")
	}
}

func TestEngineExpressionsAndClauses(t *testing.T) {
	tempDir := t.TempDir()
	dbfPath := filepath.Join(tempDir, "PRODUTOS.DBF")

	cfg := config.DefaultConfig()
	cfg.Talk = true
	buf := new(bytes.Buffer)

	eng := NewEngine(cfg, nil, buf)

	fields := []dbf.Field{
		{Name: "CODIGO", Type: dbf.TypeNumeric, Length: 4, Decimals: 0},
		{Name: "DESCRICAO", Type: dbf.TypeCharacter, Length: 30, Decimals: 0},
		{Name: "PRECO", Type: dbf.TypeNumeric, Length: 10, Decimals: 2},
		{Name: "QTD", Type: dbf.TypeNumeric, Length: 5, Decimals: 0},
	}
	tbl, err := dbf.CreateTable(dbfPath, fields)
	if err != nil {
		t.Fatalf("Erro ao criar tabela: %v", err)
	}

	items := []struct {
		cod   string
		desc  string
		preco string
		qtd   string
	}{
		{"101", "MSX 2+ GRADIENTE", "1500.00", "5"},
		{"102", "MSX 1 GRADIENTE EXPERT", "800.00", "12"},
		{"103", "DRIVE 720KB MSX2", "350.00", "20"},
		{"104", "CARTUCHO MEGARAM 256K", "250.00", "8"},
	}

	for _, item := range items {
		r := tbl.AppendBlank()
		r.Set("CODIGO", item.cod)
		r.Set("DESCRICAO", item.desc)
		r.Set("PRECO", item.preco)
		r.Set("QTD", item.qtd)
	}
	_ = tbl.Save()

	_ = eng.ExecuteLine("USE " + dbfPath)

	// Teste 1: STORE e Expressão com ?
	buf.Reset()
	_ = eng.ExecuteLine("STORE 0.10 TO DESCONTO")
	if eng.Variables["DESCONTO"].Num != 0.10 {
		t.Errorf("Variável DESCONTO não armazenada corretamente")
	}

	buf.Reset()
	_ = eng.ExecuteLine("? 100 * (1 - DESCONTO)")
	if !strings.Contains(buf.String(), "90") {
		t.Errorf("Expressão '?' falhou: %s", buf.String())
	}

	// Teste 2: LIST com FOR usando operador de substring $ e relacional
	buf.Reset()
	_ = eng.ExecuteLine(`LIST ALL FOR "GRADIENTE" $ DESCRICAO .AND. PRECO >= 1000`)
	out := buf.String()
	if !strings.Contains(out, "MSX 2+ GRADIENTE") || strings.Contains(out, "EXPERT") {
		t.Errorf("LIST com FOR composto falhou: %s", out)
	}

	// Teste 3: COUNT com FOR e armazenamento TO VAR
	buf.Reset()
	_ = eng.ExecuteLine("COUNT FOR QTD >= 10 TO ESTOQUE_ALTO")
	if eng.Variables["ESTOQUE_ALTO"].Num != 2 {
		t.Errorf("COUNT TO VAR esperava 2, obteve %f", eng.Variables["ESTOQUE_ALTO"].Num)
	}

	// Teste 4: SUM com FOR e TO VAR
	buf.Reset()
	_ = eng.ExecuteLine("SUM PRECO FOR CODIGO > 102 TO SOMA_PARCIAL")
	if eng.Variables["SOMA_PARCIAL"].Num != 600.00 {
		t.Errorf("SUM TO VAR esperava 600.00, obteve %f", eng.Variables["SOMA_PARCIAL"].Num)
	}

	// Teste 5: REPLACE com cálculo matemático na cláusula WITH
	buf.Reset()
	_ = eng.ExecuteLine("REPLACE ALL PRECO WITH PRECO * 1.10 FOR CODIGO = 104")
	if eng.ActiveTable.Records[3].Get("PRECO") != "275.00" && eng.ActiveTable.Records[3].Get("PRECO") != "275" {
		t.Errorf("REPLACE com expressão matemática falhou: %s", eng.ActiveTable.Records[3].Get("PRECO"))
	}

	// Teste 6: LOCATE FOR com CONTINUE
	buf.Reset()
	_ = eng.ExecuteLine(`LOCATE FOR "GRADIENTE" $ DESCRICAO`)
	if eng.CurrentRecordIndex != 1 {
		t.Errorf("LOCATE deveria achar registro 1, achou %d", eng.CurrentRecordIndex)
	}

	_ = eng.ExecuteLine("CONTINUE")
	if eng.CurrentRecordIndex != 2 {
		t.Errorf("CONTINUE deveria achar registro 2, achou %d", eng.CurrentRecordIndex)
	}
}

func TestEngineHelpCommands(t *testing.T) {
	cfg := config.DefaultConfig()
	buf := new(bytes.Buffer)
	eng := NewEngine(cfg, nil, buf)

	// 1. HELP sem argumentos
	buf.Reset()
	if err := eng.ExecuteLine("HELP"); err != nil {
		t.Fatalf("Erro no HELP: %v", err)
	}
	out := buf.String()
	if !strings.Contains(out, "INTRODUÇÃO") || !strings.Contains(out, "COMO USAR O COMANDO \"HELP\"") {
		t.Errorf("HELP sem argumentos não exibiu introdução esperada: %s", out)
	}

	// 2. HELP CREATE
	buf.Reset()
	_ = eng.ExecuteLine("HELP CREATE")
	out = buf.String()
	if !strings.Contains(out, "CREATE") || !strings.Contains(out, "Cria um novo arquivo") {
		t.Errorf("HELP CREATE falhou: %s", out)
	}

	// 3. HELP ?
	buf.Reset()
	_ = eng.ExecuteLine("HELP ?")
	out = buf.String()
	if !strings.Contains(out, "Avalia e mostra na tela o valor de uma expressão") {
		t.Errorf("HELP ? falhou: %s", out)
	}

	// 4. HELP FUNÇÕES (e sem acento HELP FUNCOES)
	buf.Reset()
	_ = eng.ExecuteLine("HELP FUNÇÕES")
	out = buf.String()
	if !strings.Contains(out, "@(<string1>,<string2>)") && !strings.Contains(out, "FUNÇÕES") {
		t.Errorf("HELP FUNÇÕES falhou: %s", out)
	}

	buf.Reset()
	_ = eng.ExecuteLine("HELP FUNCOES")
	out = buf.String()
	if !strings.Contains(out, "@(<string1>,<string2>)") && !strings.Contains(out, "FUNÇÕES") {
		t.Errorf("HELP FUNCOES falhou: %s", out)
	}

	// 5. HELP TELA-CHEIA
	buf.Reset()
	_ = eng.ExecuteLine("HELP TELA-CHEIA")
	out = buf.String()
	if !strings.Contains(out, "CODIGOS DE MOVIMENTO DO CURSOR NA TELA") {
		t.Errorf("HELP TELA-CHEIA falhou: %s", out)
	}

	// 6. HELP ERROS
	buf.Reset()
	_ = eng.ExecuteLine("HELP ERROS")
	out = buf.String()
	if !strings.Contains(out, "Erro de sintaxe") || !strings.Contains(out, "inválido") {
		t.Errorf("HELP ERROS falhou: %s", out)
	}

	// 7. HELP EDITOR (MED)
	buf.Reset()
	_ = eng.ExecuteLine("HELP EDITOR")
	out = buf.String()
	if !strings.Contains(out, "MSX-DOS SCREEN EDITOR (MED)") && !strings.Contains(out, "Direct command summary") {
		t.Errorf("HELP EDITOR falhou: %s", out)
	}
}

func TestParseFieldDefinition(t *testing.T) {
	tests := []struct {
		input    string
		valid    bool
		name     string
		fType    dbf.FieldType
		length   uint8
		decimals uint8
	}{
		{"NOME, C, 30", true, "NOME", dbf.TypeCharacter, 30, 0},
		{"NOME,C,30", true, "NOME", dbf.TypeCharacter, 30, 0},
		{"IDADE, N, 3", true, "IDADE", dbf.TypeNumeric, 3, 0},
		{"SALARIO, N, 10, 2", true, "SALARIO", dbf.TypeNumeric, 10, 2},
		{"ATIVO, L", true, "ATIVO", dbf.TypeLogical, 1, 0},
		{"ATIVO, L, 1", true, "ATIVO", dbf.TypeLogical, 1, 0},
		{"CIDADE C 25", true, "CIDADE", dbf.TypeCharacter, 25, 0},
		{"PRECO N 8 2", true, "PRECO", dbf.TypeNumeric, 8, 2},
		{"INVALIDO_MUITO_LONGO, C, 10", false, "", 0, 0, 0},
		{"123NOME, C, 10", false, "", 0, 0, 0},
		{"NOME, X, 10", false, "", 0, 0, 0},
		{"VALOR, N, 4, 5", false, "", 0, 0, 0}, // decimals >= length
	}

	for _, tt := range tests {
		f, err := ParseFieldDefinition(tt.input)
		if tt.valid {
			if err != nil {
				t.Errorf("ParseFieldDefinition(%q) falhou: %v", tt.input, err)
				continue
			}
			if f.Name != tt.name || f.Type != tt.fType || f.Length != tt.length || f.Decimals != tt.decimals {
				t.Errorf("ParseFieldDefinition(%q) = %+v, esperava Name=%s Type=%c Length=%d Dec=%d",
					tt.input, f, tt.name, tt.fType, tt.length, tt.decimals)
			}
		} else {
			if err == nil {
				t.Errorf("ParseFieldDefinition(%q) deveria falhar, mas obteve sucesso: %+v", tt.input, f)
			}
		}
	}
}

func TestInteractiveCreateWithDataInput(t *testing.T) {
	tempDir := t.TempDir()
	dbfPath := filepath.Join(tempDir, "CLIENTES.DBF")

	cfg := config.DefaultConfig()
	cfg.Talk = true
	buf := new(bytes.Buffer)
	eng := NewEngine(cfg, nil, buf)

	// Simula a entrada do usuário:
	// Campo 1: NOME, C, 20
	// Campo 2: IDADE, N, 3
	// Campo 3: ATIVO, L
	// Campo 4: <Enter vazio para finalizar estrutura>
	// Pergunta INPUT DATA NOW?: S (Sim)
	// Registro 1:
	//   NOME: ALICE
	//   IDADE: 28
	//   ATIVO: T
	// Registro 2:
	//   NOME: BOB
	//   IDADE: 35
	//   ATIVO: F
	// Registro 3:
	//   <Enter vazio no primeiro campo para encerrar APPEND>
	inputScript := strings.Join([]string{
		"NOME, C, 20",
		"IDADE, N, 3",
		"ATIVO, L",
		"",
		"S",
		"ALICE",
		"28",
		"T",
		"BOB",
		"35",
		"F",
		"",
	}, "\n") + "\n"

	eng.SetIn(strings.NewReader(inputScript))

	err := eng.ExecuteLine("CREATE " + dbfPath)
	if err != nil {
		t.Fatalf("Erro no CREATE interativo: %v", err)
	}

	out := buf.String()
	if !strings.Contains(out, "ENTER RECORD STRUCTURE AS FOLLOWS") {
		t.Errorf("Saída não contém o cabeçalho de estrutura: %s", out)
	}
	if !strings.Contains(out, "INPUT DATA NOW?") {
		t.Errorf("Saída não contém a pergunta INPUT DATA NOW?: %s", out)
	}

	// Verifica se a tabela foi aberta e possui os registros
	if eng.ActiveTable == nil {
		t.Fatalf("Tabela ativa não foi definida após CREATE")
	}

	if len(eng.ActiveTable.Records) != 2 {
		t.Fatalf("Esperava 2 registros na tabela, obteve %d", len(eng.ActiveTable.Records))
	}

	rec1 := eng.ActiveTable.Records[0]
	if rec1.Get("NOME") != "ALICE" || rec1.Get("IDADE") != "28" || rec1.Get("ATIVO") != "T" {
		t.Errorf("Registro 1 incorreto: NOME=%s IDADE=%s ATIVO=%s", rec1.Get("NOME"), rec1.Get("IDADE"), rec1.Get("ATIVO"))
	}

	rec2 := eng.ActiveTable.Records[1]
	if rec2.Get("NOME") != "BOB" || rec2.Get("IDADE") != "35" || rec2.Get("ATIVO") != "F" {
		t.Errorf("Registro 2 incorreto: NOME=%s IDADE=%s ATIVO=%s", rec2.Get("NOME"), rec2.Get("IDADE"), rec2.Get("ATIVO"))
	}

	// Testa LIST ALL na tabela recém-criada
	buf.Reset()
	if err := eng.ExecuteLine("LIST ALL"); err != nil {
		t.Fatalf("Erro ao listar registros: %v", err)
	}
	listOut := buf.String()
	if !strings.Contains(listOut, "ALICE") || !strings.Contains(listOut, "BOB") {
		t.Errorf("LIST ALL não listou registros inseridos: %s", listOut)
	}
}

func TestInteractiveCreateWithoutDataInput(t *testing.T) {
	tempDir := t.TempDir()
	dbfPath := filepath.Join(tempDir, "VAZIO.DBF")

	cfg := config.DefaultConfig()
	buf := new(bytes.Buffer)
	eng := NewEngine(cfg, nil, buf)

	// Cria estrutura mas responde N para INPUT DATA NOW?
	inputScript := strings.Join([]string{
		"CODIGO, N, 5",
		"DESCRICAO, C, 30",
		"",
		"N",
	}, "\n") + "\n"

	eng.SetIn(strings.NewReader(inputScript))

	err := eng.ExecuteLine("CREATE " + dbfPath)
	if err != nil {
		t.Fatalf("Erro no CREATE sem dados: %v", err)
	}

	if eng.ActiveTable == nil {
		t.Fatalf("Tabela ativa não definida")
	}

	if len(eng.ActiveTable.Records) != 0 {
		t.Errorf("Tabela deveria estar vazia, possui %d registros", len(eng.ActiveTable.Records))
	}
}

func TestInteractiveAppendAndEdit(t *testing.T) {
	tempDir := t.TempDir()
	dbfPath := filepath.Join(tempDir, "TESTE_EDIT.DBF")

	cfg := config.DefaultConfig()
	buf := new(bytes.Buffer)
	eng := NewEngine(cfg, nil, buf)

	// 1. Cria tabela com 1 registro inicial
	inputCreate := strings.Join([]string{
		"ID, N, 4",
		"NOME, C, 20",
		"",
		"S",
		"100",
		"VALOR INICIAL",
		"",
	}, "\n") + "\n"

	eng.SetIn(strings.NewReader(inputCreate))
	_ = eng.ExecuteLine("CREATE " + dbfPath)

	if len(eng.ActiveTable.Records) != 1 {
		t.Fatalf("Esperava 1 registro, obteve %d", len(eng.ActiveTable.Records))
	}

	// 2. Executa APPEND interativo para adicionar mais 1 registro
	inputAppend := strings.Join([]string{
		"200",
		"VALOR SEGUNDO",
		"",
	}, "\n") + "\n"
	eng.SetIn(strings.NewReader(inputAppend))
	if err := eng.ExecuteLine("APPEND"); err != nil {
		t.Fatalf("Erro no APPEND interativo: %v", err)
	}

	if len(eng.ActiveTable.Records) != 2 {
		t.Fatalf("Esperava 2 registros após APPEND, obteve %d", len(eng.ActiveTable.Records))
	}

	// 3. Executa EDIT 1 para alterar o primeiro registro
	inputEdit := strings.Join([]string{
		"101",
		"VALOR MODIFICADO",
	}, "\n") + "\n"
	eng.SetIn(strings.NewReader(inputEdit))
	if err := eng.ExecuteLine("EDIT 1"); err != nil {
		t.Fatalf("Erro no EDIT 1: %v", err)
	}

	if eng.ActiveTable.Records[0].Get("ID") != "101" {
		t.Errorf("ID esperado 101, obteve %s", eng.ActiveTable.Records[0].Get("ID"))
	}
	if eng.ActiveTable.Records[0].Get("NOME") != "VALOR MODIFICADO" {
		t.Errorf("NOME esperado VALOR MODIFICADO, obteve %s", eng.ActiveTable.Records[0].Get("NOME"))
	}
}
