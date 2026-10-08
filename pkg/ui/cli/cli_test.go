package cli

import (
	"strings"
	"testing"
)

func TestHighlightLine(t *testing.T) {
	tests := []struct {
		input    string
		contains []string
	}{
		{
			input:    "USE CLIENTES.DBF",
			contains: []string{AnsiCyan, "USE", AnsiGreen, "CLIENTES.DBF", AnsiReset},
		},
		{
			input:    "LIST ALL FOR PRECO > 100.50 .AND. TRIM(NOME) = \"MARIA\"",
			contains: []string{AnsiCyan, "LIST", AnsiWhite, "ALL", AnsiWhite, "FOR", AnsiYellow, ">", AnsiBlue, "100.50", AnsiYellow, ".AND.", AnsiMagenta, "TRIM", AnsiGreen, "\"MARIA\""},
		},
		{
			input:    "* Este é um comentário dBase",
			contains: []string{AnsiGray, "* Este é um comentário dBase", AnsiReset},
		},
		{
			input:    "NOTE Outro comentário",
			contains: []string{AnsiGray, "NOTE Outro comentário", AnsiReset},
		},
		{
			input:    "? RECNO() + 10",
			contains: []string{AnsiCyan, "?", AnsiMagenta, "RECNO", AnsiYellow, "+", AnsiBlue, "10"},
		},
	}

	for _, tt := range tests {
		got := HighlightLine(tt.input)
		for _, exp := range tt.contains {
			if !strings.Contains(got, exp) {
				t.Errorf("HighlightLine(%q) não contém %q. Resultado: %q", tt.input, exp, got)
			}
		}
	}
}

func TestFormatError(t *testing.T) {
	errStr := "Arquivo não encontrado"
	formatted := FormatError(errStr)
	if !strings.Contains(formatted, AnsiRed) || !strings.Contains(formatted, "*** ERRO: Arquivo não encontrado") {
		t.Errorf("FormatError não aplicou vermelho ou formato correto: %s", formatted)
	}
}

func TestAutocomplete(t *testing.T) {
	// Teste 1: Autocomplete com único candidato ("CRE" -> "CREATE ")
	var state AutocompleteState
	line, pos, ok := state.NextCompletion("cre", 3)
	if !ok || line != "CREATE " || pos != 7 {
		t.Errorf("Falha no autocomplete de 'cre': esperado 'CREATE ', obteve %q (pos %d)", line, pos)
	}

	// Teste 2: Autocomplete cíclico ("CO" -> "CONTINUE" -> "COPY" -> "COUNT")
	state.Reset()
	line1, _, ok1 := state.NextCompletion("co", 2)
	if !ok1 || line1 != "CONTINUE" {
		t.Errorf("TAB 1 em 'co' esperava 'CONTINUE', obteve %q", line1)
	}

	line2, _, ok2 := state.NextCompletion(line1, len(line1))
	if !ok2 || line2 != "COPY" {
		t.Errorf("TAB 2 em 'co' esperava 'COPY', obteve %q", line2)
	}

	line3, _, ok3 := state.NextCompletion(line2, len(line2))
	if !ok3 || line3 != "COUNT" {
		t.Errorf("TAB 3 em 'co' esperava 'COUNT', obteve %q", line3)
	}

	// Ciclo volta para o início
	line4, _, ok4 := state.NextCompletion(line3, len(line3))
	if !ok4 || line4 != "CONTINUE" {
		t.Errorf("TAB 4 (ciclo) em 'co' esperava 'CONTINUE', obteve %q", line4)
	}

	// Teste 3: Autocomplete no contexto de HELP ("HELP COM" -> "HELP COMMANDS" ou "HELP COMANDOS")
	state.Reset()
	lineHelp, _, okHelp := state.NextCompletion("HELP com", 8)
	if !okHelp || (!strings.Contains(lineHelp, "COMMANDS") && !strings.Contains(lineHelp, "COMANDOS")) {
		t.Errorf("Autocomplete após HELP falhou: obteve %q", lineHelp)
	}

	// Teste 4: Autocomplete no contexto de SET ("SET TAL" -> "SET TALK ")
	state.Reset()
	lineSet, _, okSet := state.NextCompletion("SET tal", 7)
	if !okSet || lineSet != "SET TALK " {
		t.Errorf("Autocomplete após SET falhou: obteve %q", lineSet)
	}
}

func TestLineEditorNonInteractive(t *testing.T) {
	input := "USE TESTE\nLIST ALL\nQUIT\n"
	in := strings.NewReader(input)
	out := new(strings.Builder)

	editor := NewLineEditor(in, out)

	l1, err := editor.ReadLine(".")
	if err != nil || l1 != "USE TESTE" {
		t.Errorf("Erro ao ler linha 1: %v, obtido: %q", err, l1)
	}

	l2, err := editor.ReadLine(".")
	if err != nil || l2 != "LIST ALL" {
		t.Errorf("Erro ao ler linha 2: %v, obtido: %q", err, l2)
	}

	l3, err := editor.ReadLine(".")
	if err != nil || l3 != "QUIT" {
		t.Errorf("Erro ao ler linha 3: %v, obtido: %q", err, l3)
	}
}
