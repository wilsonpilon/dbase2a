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
	}{
		{"USE CLIENTES", CmdUse, "CLIENTES", ""},
		{".USE PROD.DBF", CmdUse, "PROD.DBF", ""},
		{"LIST ALL", CmdList, "", "ALL"},
		{"LIST NOME, VALOR FOR VALOR > 100", CmdList, "", ""},
		{"APPEND BLANK", CmdAppend, "", ""},
		{"PACK", CmdPack, "", ""},
		{"ZAP", CmdZap, "", ""},
		{"DELETE ALL", CmdDelete, "", "ALL"},
		{"RECALL ALL", CmdRecall, "", "ALL"},
		{"DO ROTINA.CMD", CmdDo, "ROTINA.CMD", ""},
		{"SET TALK ON", CmdSet, "", ""},
		{"QUIT", CmdQuit, "", ""},
	}

	for _, tc := range tests {
		stmts, err := Parse(tc.input)
		if err != nil {
			t.Fatalf("Erro inesperado no parse de '%s': %v", tc.input, err)
		}
		if len(stmts) != 1 {
			t.Fatalf("Esperava 1 statement para '%s', obteve %d", tc.input, len(stmts))
		}
		if stmts[0].Type != tc.expectedCmd {
			t.Errorf("Para '%s': esperava comando %s, obteve %s", tc.input, tc.expectedCmd, stmts[0].Type)
		}
		if tc.target != "" && stmts[0].Target != tc.target {
			t.Errorf("Para '%s': esperava target '%s', obteve '%s'", tc.input, tc.target, stmts[0].Target)
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
