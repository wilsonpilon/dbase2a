package help

import (
	"os"
	"path/filepath"
	"strings"
	"testing"
)

func TestGenerateAndVerifyUTF8HelpFile(t *testing.T) {
	rawPath := filepath.Join("..", "..", "resource", "files", "DBASEMSG.TXT")
	data, err := os.ReadFile(rawPath)
	if err != nil {
		t.Fatalf("Erro ao ler DBASEMSG.TXT original: %v", err)
	}

	utf8Content := MSXToUTF8(data)
	utf8Path := filepath.Join("..", "..", "resource", "files", "DBASEMSG_UTF8.TXT")

	if err := os.WriteFile(utf8Path, []byte(utf8Content), 0644); err != nil {
		t.Fatalf("Erro ao gravar DBASEMSG_UTF8.TXT: %v", err)
	}

	// Verifica se arquivo UTF-8 pode ser lido pelo HelpSystem
	hs := NewHelpSystem()
	if err := hs.LoadFile(utf8Path); err != nil {
		t.Fatalf("Erro ao carregar DBASEMSG_UTF8.TXT: %v", err)
	}

	topics := []string{"DBASE", "HELP", "CREATE", "USE", "LIST", "FUNÇÕES", "TELA-CHEIA", "ERROS", "LIMITES"}
	for _, top := range topics {
		txt, ok := hs.GetTopic(top)
		if !ok || txt == "" {
			t.Errorf("Tópico %s não recuperado do arquivo UTF-8", top)
		}
	}
}

func TestHelpSystemFallback(t *testing.T) {
	hs := NewHelpSystem()
	hs.loadFallbackHelp()

	txt, ok := hs.GetTopic("")
	if !ok || !strings.Contains(txt, "COMO USAR O COMANDO \"HELP\"") {
		t.Errorf("Fallback help falhou: %s", txt)
	}
}
