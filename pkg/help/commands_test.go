package help

import (
	"bytes"
	"strings"
	"testing"
)

func TestHelpCommandsTopic(t *testing.T) {
	hs := NewHelpSystem()
	_ = hs.LoadDefault()

	// 1. HELP COMMANDS
	cmdHelp, ok := hs.GetTopic("COMMANDS")
	if !ok {
		t.Fatalf("Tópico COMMANDS não foi encontrado")
	}
	if !strings.Contains(cmdHelp, "RESUMO DE COMANDOS") || !strings.Contains(cmdHelp, "APPEND") || !strings.Contains(cmdHelp, "ZAP") {
		t.Errorf("Conteúdo de COMMANDS incompleto: %s", cmdHelp)
	}

	// 2. HELP COMANDOS (alias em português)
	cmdHelpPt, ok := hs.GetTopic("COMANDOS")
	if !ok {
		t.Fatalf("Tópico COMANDOS não foi encontrado")
	}
	if !strings.Contains(cmdHelpPt, "RESUMO DE COMANDOS") {
		t.Errorf("Conteúdo de COMANDOS incorreto: %s", cmdHelpPt)
	}
}

func TestPagerNonInteractive(t *testing.T) {
	text := "Linha 1\nLinha 2\nLinha 3\nLinha 4\nLinha 5"
	in := strings.NewReader("")
	out := new(bytes.Buffer)

	pager := NewPager(in, out)
	if err := pager.Display(text); err != nil {
		t.Fatalf("Erro ao exibir com paginador: %v", err)
	}

	res := out.String()
	for _, l := range []string{"Linha 1", "Linha 2", "Linha 3", "Linha 4", "Linha 5"} {
		if !strings.Contains(res, l) {
			t.Errorf("Saída do paginador não contém %q: %s", l, res)
		}
	}
}
