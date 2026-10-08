// Package help provê o sistema de ajuda interativa do dBASE II++ (DBASEMSG.TXT / MED.HLP),
// com conversão transparente do conjunto de caracteres do MSX para UTF-8 no PC,
// indexação de tópicos, comandos, utilitários, mensagens de erro e modo de tela cheia.
package help

import (
	"bufio"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"unicode"
	"unicode/utf8"
)

// MSXToUTF8 converte um fluxo de bytes codificado na ROM de caracteres do MSX nacional (Gradiente/Sharp)
// para uma string UTF-8 padrão no PC. Se os bytes já forem UTF-8 válido, são retornados diretamente.
func MSXToUTF8(data []byte) string {
	if utf8.Valid(data) {
		// Se já for UTF-8 válido (como arquivos editados no PC), retorna como string diretamente
		// removendo eventual byte 0x1A (EOF do MSX-DOS/CP/M) no final
		res := string(data)
		res = strings.TrimRight(res, "\x1a")
		return res
	}

	var buf strings.Builder
	buf.Grow(len(data))

	for i := 0; i < len(data); i++ {
		b := data[i]
		switch b {
		case 0x80:
			buf.WriteRune('Ç')
		case 0x81:
			buf.WriteRune('ü')
		case 0x82:
			buf.WriteRune('é')
		case 0x83:
			buf.WriteRune('â')
		case 0x84:
			buf.WriteRune('Á')
		case 0x85:
			buf.WriteRune('à')
		case 0x86:
			buf.WriteRune('É')
		case 0x87:
			buf.WriteRune('ç')
		case 0x88:
			buf.WriteRune('ê')
		case 0x89:
			buf.WriteRune('Ê')
		case 0x8A:
			buf.WriteRune('è')
		case 0x8B:
			buf.WriteRune('Í')
		case 0x8C:
			buf.WriteRune('Ó')
		case 0x8D:
			buf.WriteRune('Ú')
		case 0x8E:
			buf.WriteRune('Ä')
		case 0x8F:
			buf.WriteRune('Â')
		case 0x90:
			buf.WriteRune('É')
		case 0x91:
			buf.WriteRune('Ô')
		case 0x92, 0x93:
			buf.WriteRune('ô')
		case 0x94:
			buf.WriteRune('ö')
		case 0x95:
			buf.WriteRune('ò')
		case 0x96:
			buf.WriteRune('û')
		case 0x97:
			buf.WriteRune('ù')
		case 0x98:
			buf.WriteRune('ÿ')
		case 0x99:
			buf.WriteRune('Ö')
		case 0x9A:
			buf.WriteRune('Ü')
		case 0xA0:
			buf.WriteRune('á')
		case 0xA1:
			buf.WriteRune('í')
		case 0xA2:
			buf.WriteRune('ó')
		case 0xA3:
			buf.WriteRune('ú')
		case 0xA4:
			buf.WriteRune('ñ')
		case 0xA5:
			buf.WriteRune('Ñ')
		case 0xA6:
			buf.WriteRune('ª')
		case 0xA7:
			buf.WriteRune('º')
		case 0xA8:
			buf.WriteRune('¿')
		case 0xB0:
			buf.WriteRune('Ã')
		case 0xB1:
			buf.WriteRune('ã')
		case 0xB2, 0xB4:
			buf.WriteRune('Õ')
		case 0xB3, 0xB5:
			buf.WriteRune('õ')
		case 0x1A: // EOF CP/M / MSX-DOS
			// Ignora caractere de fim de arquivo do MSX-DOS
			break
		default:
			if b < 0x80 {
				buf.WriteByte(b)
			} else {
				// Caractere não mapeado preservado
				buf.WriteRune(rune(b))
			}
		}
	}

	return buf.String()
}

// HelpSystem gerencia o catálogo de tópicos de ajuda do dBASE II.
type HelpSystem struct {
	mu           sync.RWMutex
	topics       map[string]string // Chave normalizada -> Conteúdo do tópico
	topicList    []string          // Lista de tópicos originais disponíveis
	defaultIntro string            // Tópico introdutório padrão (DBASE)
	loaded       bool
	sourcePath   string
}

// NewHelpSystem cria uma nova instância do sistema de ajuda.
func NewHelpSystem() *HelpSystem {
	return &HelpSystem{
		topics:    make(map[string]string),
		topicList: make([]string, 0),
	}
}

// normalizeKey normaliza a chave de busca (maiúsculas, sem acentos, sem hífens e sem espaços).
func normalizeKey(s string) string {
	s = strings.ToUpper(strings.TrimSpace(s))
	var b strings.Builder
	for _, r := range s {
		switch r {
		case 'Á', 'À', 'Ã', 'Â', 'Ä':
			b.WriteRune('A')
		case 'É', 'È', 'Ê', 'Ë':
			b.WriteRune('E')
		case 'Í', 'Ì', 'Î', 'Ï':
			b.WriteRune('I')
		case 'Ó', 'Ò', 'Õ', 'Ô', 'Ö':
			b.WriteRune('O')
		case 'Ú', 'Ù', 'Û', 'Ü':
			b.WriteRune('U')
		case 'Ç':
			b.WriteRune('C')
		case '-', '_', ' ', '/':
			// Ignora delimitadores para casar ex: TELA-CHEIA com TELA CHEIA ou TELACHEIA
			continue
		default:
			if unicode.IsLetter(r) || unicode.IsDigit(r) || r == '?' || r == '@' || r == '*' || r == '$' {
				b.WriteRune(r)
			}
		}
	}
	return b.String()
}

// LoadFile carrega e indexa o arquivo DBASEMSG.TXT especificado.
func (h *HelpSystem) LoadFile(path string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return fmt.Errorf("não foi possível abrir o arquivo de ajuda (%s): %w", path, err)
	}

	h.mu.Lock()
	defer h.mu.Unlock()

	h.sourcePath = path
	if err := h.parseContent(data); err != nil {
		return err
	}

	// Tenta carregar também o MED.HLP do mesmo diretório se existir
	dir := filepath.Dir(path)
	medPath := filepath.Join(dir, "MED.HLP")
	if medData, err := os.ReadFile(medPath); err == nil {
		h.registerMedHelp(medData)
	}

	// Registra tópicos virtuais e comandos gerais
	h.registerVirtualTopics()

	return nil
}

// LoadDefault tenta carregar DBASEMSG.TXT dos locais padrão do sistema.
func (h *HelpSystem) LoadDefault() error {
	candidatePaths := []string{
		"DBASEMSG.TXT",
		"dbasemsg.txt",
		"resource/files/DBASEMSG.TXT",
		"resource/files/dbasemsg.txt",
		"../resource/files/DBASEMSG.TXT",
		"../../resource/files/DBASEMSG.TXT",
		filepath.Join(os.Getenv("DBASE2_HOME"), "DBASEMSG.TXT"),
	}

	for _, p := range candidatePaths {
		if p == "" {
			continue
		}
		if _, err := os.Stat(p); err == nil {
			return h.LoadFile(p)
		}
	}

	// Se não encontrar em disco, carrega a base embutida essencial
	h.mu.Lock()
	defer h.mu.Unlock()
	h.loadFallbackHelp()
	return nil
}

// parseContent decodifica e indexa as seções do DBASEMSG.TXT.
func (h *HelpSystem) parseContent(rawData []byte) error {
	utf8Text := MSXToUTF8(rawData)
	scanner := bufio.NewScanner(strings.NewReader(utf8Text))

	var currentAliases []string
	var currentBody strings.Builder
	h.topics = make(map[string]string)
	h.topicList = make([]string, 0)

	flushTopic := func() {
		if len(currentAliases) > 0 {
			body := strings.TrimSpace(currentBody.String())
			for _, alias := range currentAliases {
				rawAlias := strings.ToUpper(strings.TrimSpace(alias))
				normAlias := normalizeKey(rawAlias)

				h.topics[rawAlias] = body
				h.topics[normAlias] = body
				h.topicList = append(h.topicList, rawAlias)
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
			} else {
				if currentBody.Len() > 0 {
					flushTopic()
				}
				currentAliases = append(currentAliases, tag)
			}
		} else {
			if len(currentAliases) > 0 {
				currentBody.WriteString(line)
				currentBody.WriteString("\n")
			}
		}
	}
	flushTopic()

	// Aliases adicionais convenientes para navegação
	h.addTopicAliases()

	if intro, ok := h.topics["DBASE"]; ok {
		h.defaultIntro = intro
	} else if intro, ok := h.topics["HELP"]; ok {
		h.defaultIntro = intro
	}

	h.loaded = true
	return nil
}

// registerVirtualTopics registra comandos gerais e resumos sintéticos como COMMANDS/COMANDOS.
func (h *HelpSystem) registerVirtualTopics() {
	cmdKeys := []string{"COMMANDS", "COMANDOS", "COMMAND", "COMANDO", "LISTA DE COMANDOS", "COMANDOS DBASE"}
	for _, k := range cmdKeys {
		h.topics[k] = CommandsSummaryText
		h.topics[normalizeKey(k)] = CommandsSummaryText
	}
	h.topicList = append(h.topicList, "COMMANDS (COMANDOS)")
}

// addTopicAliases cria aliases úteis para comandos e tópicos comuns.
func (h *HelpSystem) addTopicAliases() {
	aliases := map[string][]string{
		"TELA-CHEIA":  {"TELACHEIA", "TELA", "FULLSCREEN", "CURSOR"},
		"FUNÇÕES":     {"FUNCOES", "FUNCAO", "FUNÇÃO", "FUNCTIONS"},
		"UTILITÁRIOS": {"UTILITARIOS", "UTILITARIO", "UTILITÁRIO", "UTILS"},
		"ERROS":       {"ERRO", "ERROR", "ERRORS", "MENSAGENS"},
		"LIMITES":     {"LIMITE", "LIMITS"},
		"GO":          {"GOTO"},
		"GOTO":        {"GO"},
		"DO":          {"DOWHILE", "DO WHILE"},
		"NOTE":        {"*"},
		"?":           {"PRINT", "AVALIAR"},
		"??":          {"PRINTSAME"},
		"MODIFY":      {"MODIFY COMMAND", "MODIFY STRUCTURE", "MODICOMM"},
		"COMMANDS":    {"COMANDOS", "COMMAND", "COMANDO"},
	}

	for mainTopic, aliasList := range aliases {
		if content, ok := h.topics[mainTopic]; ok {
			for _, a := range aliasList {
				h.topics[strings.ToUpper(a)] = content
				h.topics[normalizeKey(a)] = content
			}
		}
	}
}

// registerMedHelp cadastra a ajuda do editor de texto/programas MED (MSX-DOS Screen Editor).
func (h *HelpSystem) registerMedHelp(medData []byte) {
	medText := MSXToUTF8(medData)
	// Limpa caracteres binários de fim de bloco se houver
	if idx := strings.Index(medText, "HELP end."); idx != -1 {
		medText = medText[:idx+len("HELP end.\nHit any key !!")]
	}
	medText = strings.TrimSpace(medText)

	header := "================================================================================\n" +
		"  MSX-DOS SCREEN EDITOR (MED) - AJUDA DO EDITOR DE PROGRAMAS E PROCEDURES      \n" +
		"================================================================================\n\n"

	fullEditorHelp := header + medText

	editorKeys := []string{"EDITOR", "MED", "MED.COM", "MED.HLP", "SCREEN EDITOR", "EDITAR", "PROGRAM EDITOR"}
	for _, k := range editorKeys {
		h.topics[k] = fullEditorHelp
		h.topics[normalizeKey(k)] = fullEditorHelp
	}
	h.topicList = append(h.topicList, "EDITOR (MED)")
}

// GetTopic busca o texto de ajuda para um comando ou tópico específico.
func (h *HelpSystem) GetTopic(query string) (string, bool) {
	h.mu.RLock()
	defer h.mu.RUnlock()

	q := strings.TrimSpace(query)
	if q == "" {
		if h.defaultIntro != "" {
			return h.defaultIntro, true
		}
		return "Não há mensagem HELP.", false
	}

	upperQ := strings.ToUpper(q)
	normQ := normalizeKey(upperQ)

	// 1. Busca exata pela chave bruta
	if text, ok := h.topics[upperQ]; ok {
		return text, true
	}
	// 2. Busca pela chave normalizada (sem acentos / pontuação)
	if text, ok := h.topics[normQ]; ok {
		return text, true
	}

	// 3. Busca por prefixo normalizado
	for key, text := range h.topics {
		if strings.HasPrefix(key, normQ) || strings.HasPrefix(key, upperQ) {
			return text, true
		}
	}

	return "", false
}

// Topics retorna a lista de tópicos registrados.
func (h *HelpSystem) Topics() []string {
	h.mu.RLock()
	defer h.mu.RUnlock()
	res := make([]string, len(h.topicList))
	copy(res, h.topicList)
	return res
}

// IsLoaded informa se os arquivos de ajuda foram carregados com sucesso.
func (h *HelpSystem) IsLoaded() bool {
	h.mu.RLock()
	defer h.mu.RUnlock()
	return h.loaded
}

// loadFallbackHelp provê mensagens essenciais de ajuda caso DBASEMSG.TXT não seja encontrado.
func (h *HelpSystem) loadFallbackHelp() {
	h.defaultIntro = `HELP - ARQUIVO TEXTO VERSAO 1.0 PARA dBASE II++
Copyright (C) 1987 DATALOGICA / 2026 dBase II++ PC Edition.

COMO USAR O COMANDO "HELP":
Digite HELP <comando> para obter ajuda detalhada sobre um comando específico.
Exemplos:
  HELP CREATE
  HELP USE
  HELP LIST
  HELP DISPLAY
  HELP APPEND
  HELP REPLACE
  HELP DELETE
  HELP RECALL
  HELP PACK
  HELP ZAP
  HELP TELA-CHEIA
  HELP ERROS
  HELP FUNÇÕES
  HELP EDITOR
`
	h.topics["DBASE"] = h.defaultIntro
	h.topics["HELP"] = h.defaultIntro
	h.registerVirtualTopics()
	h.loaded = true
}
