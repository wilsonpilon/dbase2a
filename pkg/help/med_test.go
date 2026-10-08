package help

import (
	"fmt"
	"os"
	"testing"
)

func TestInspectMedFiles(t *testing.T) {
	for _, fname := range []string{"MED.HLP", "MED.MES"} {
		path := "../../resource/files/" + fname
		data, err := os.ReadFile(path)
		if err != nil {
			t.Logf("Aviso: %s não pôde ser lido: %v", fname, err)
			continue
		}
		utf8Text := MSXToUTF8(data)
		fmt.Printf("=== ARQUIVO %s (%d bytes) ===\n%s\n\n", fname, len(data), utf8Text)
	}
}
