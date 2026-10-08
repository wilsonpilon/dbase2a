package parser

import (
	"testing"
)

func TestParseCommands(t *testing.T) {
	tests := []struct {
		input       string
		expectedCmd CommandType
		target      string
		scope       string
		forClause   string
		whileClause string
	}{
		{input: "USE CLIENTES", expectedCmd: CmdUse, target: "CLIENTES"},
		{input: ".USE PROD.DBF", expectedCmd: CmdUse, target: "PROD.DBF"},
		{input: "LIST ALL", expectedCmd: CmdList, scope: "ALL"},
		{input: "LIST NOME, VALOR FOR VALOR > 100", expectedCmd: CmdList, forClause: "VALOR > 100"},
		{input: "LIST ALL FOR IDADE > 30 WHILE ATIVO = .T.", expectedCmd: CmdList, scope: "ALL", forClause: "IDADE > 30", whileClause: "ATIVO = .T."},
		{input: "APPEND BLANK", expectedCmd: CmdAppend},
		{input: "PACK", expectedCmd: CmdPack},
		{input: "ZAP", expectedCmd: CmdZap},
		{input: "DELETE ALL", expectedCmd: CmdDelete, scope: "ALL"},
		{input: "RECALL ALL", expectedCmd: CmdRecall, scope: "ALL"},
		{input: "STORE 100 TO TOTAL", expectedCmd: CmdStore, target: "TOTAL"},
		{input: "TAXA = 0.15", expectedCmd: CmdStore, target: "TAXA"},
		{input: "? 10 + 20", expectedCmd: CmdPrint},
		{input: "?? UPPER(NOME)", expectedCmd: CmdPrintSame},
		{input: "GOTO 15", expectedCmd: CmdGoto},
		{input: "SKIP 2", expectedCmd: CmdSkip},
		{input: "DO ROTINA.CMD", expectedCmd: CmdDo, target: "ROTINA.CMD"},
		{input: "SET TALK ON", expectedCmd: CmdSet},
		{input: "QUIT", expectedCmd: CmdQuit},
	}

	for _, tc := range tests {
		stmts, err := Parse(tc.input)
		if err != nil {
			t.Fatalf("Erro inesperado no parse de '%s': %v", tc.input, err)
		}
		if len(stmts) != 1 {
			t.Fatalf("Esperava 1 statement para '%s', obteve %d", tc.input, len(stmts))
		}
		stmt := stmts[0]
		if stmt.Type != tc.expectedCmd {
			t.Errorf("Para '%s': esperava comando %s, obteve %s", tc.input, tc.expectedCmd, stmt.Type)
		}
		if tc.target != "" && stmt.Target != tc.target {
			t.Errorf("Para '%s': esperava target '%s', obteve '%s'", tc.input, tc.target, stmt.Target)
		}
		if tc.scope != "" && stmt.Scope != tc.scope {
			t.Errorf("Para '%s': esperava scope '%s', obteve '%s'", tc.input, tc.scope, stmt.Scope)
		}
		if tc.forClause != "" && stmt.ForClause != tc.forClause {
			t.Errorf("Para '%s': esperava ForClause '%s', obteve '%s'", tc.input, tc.forClause, stmt.ForClause)
		}
		if tc.whileClause != "" && stmt.WhileClause != tc.whileClause {
			t.Errorf("Para '%s': esperava WhileClause '%s', obteve '%s'", tc.input, tc.whileClause, stmt.WhileClause)
		}
	}
}

func TestParseMultipleStatements(t *testing.T) {
	input := "USE CLIENTES; LIST ALL; QUIT"
	stmts, err := Parse(input)
	if err != nil {
		t.Fatalf("Erro no parse: %v", err)
	}
	if len(stmts) != 3 {
		t.Fatalf("Esperava 3 comandos separados por ';', obteve %d", len(stmts))
	}
	if stmts[0].Type != CmdUse || stmts[1].Type != CmdList || stmts[2].Type != CmdQuit {
		t.Errorf("Comandos decodificados incorretamente: %+v", stmts)
	}
}
