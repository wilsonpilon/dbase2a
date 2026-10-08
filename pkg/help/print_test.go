package help

import (
	"fmt"
	"testing"
)

func TestPrintHelpAndIntro(t *testing.T) {
	hs := NewHelpSystem()
	if err := hs.LoadFile("../../resource/files/DBASEMSG.TXT"); err != nil {
		t.Fatalf("Erro ao carregar help: %v", err)
	}

	intro, _ := hs.GetTopic("DBASE")
	fmt.Printf("=== INTRO DBASE ===\n%s\n\n", intro)

	helpCmd, _ := hs.GetTopic("HELP")
	fmt.Printf("=== HELP COMMAND ===\n%s\n\n", helpCmd)
}
