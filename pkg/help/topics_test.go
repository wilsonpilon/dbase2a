package help

import (
	"bufio"
	"fmt"
	"os"
	"strings"
	"testing"
)

func TestParseTopics(t *testing.T) {
	data, err := os.ReadFile("../../resource/files/DBASEMSG.TXT")
	if err != nil {
		t.Fatalf("Erro ao ler DBASEMSG.TXT: %v", err)
	}

	utf8Text := MSXToUTF8(data)

	// Agora analisar a estrutura de tópicos
	// Tópicos começam com *NOME_DO_TOPICO
	// e terminam com *EXIT ou próximo *TOPICO
	scanner := bufio.NewScanner(strings.NewReader(utf8Text))
	var currentTopic string
	var currentAliases []string
	var currentBody strings.Builder
	topics := make(map[string]string)

	flushTopic := func() {
		if len(currentAliases) > 0 {
			body := strings.TrimSpace(currentBody.String())
			for _, alias := range currentAliases {
				norm := strings.ToUpper(strings.TrimSpace(alias))
				topics[norm] = body
			}
			currentAliases = nil
			currentBody.Reset()
		}
	}

	for scanner.Scan() {
		line := scanner.Text()
		trimmed := strings.TrimSpace(line)

		if strings.HasPrefix(trimmed, "*") {
			tag := strings.TrimPrefix(trimmed, "*")
			tag = strings.TrimSpace(tag)
			if strings.EqualFold(tag, "EXIT") {
				flushTopic()
				currentTopic = ""
			} else {
				// Pode ser uma lista de aliases como *COPIA seguido de *COPIA4
				if currentBody.Len() > 0 {
					flushTopic()
				}
				currentAliases = append(currentAliases, tag)
				currentTopic = tag
			}
		} else {
			if currentTopic != "" || len(currentAliases) > 0 {
				currentBody.WriteString(line)
				currentBody.WriteString("\n")
			}
		}
	}
	flushTopic()

	fmt.Printf("Total de tópicos mapeados: %d\n", len(topics))
	var topicNames []string
	for name := range topics {
		topicNames = append(topicNames, name)
	}
	fmt.Printf("Tópicos: %s\n", strings.Join(topicNames, ", "))

	// Testa alguns tópicos conhecidos
	for _, k := range []string{"DBASE", "HELP", "?", "CREATE", "USE", "LIST", "UTILITÁRIOS", "TELA-CHEIA", "ERROS", "FUNÇÕES"} {
		if body, ok := topics[strings.ToUpper(k)]; ok {
			fmt.Printf("=== TÓPICO [%s] ===\n%s\n\n", k, body[:min(len(body), 150)])
		} else {
			t.Errorf("Tópico esperado não encontrado: %s", k)
		}
	}
}

func min(a, b int) int {
	if a < b {
		return a
	}
	return b
}
