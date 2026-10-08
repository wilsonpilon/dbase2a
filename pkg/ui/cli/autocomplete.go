// Package cli implementa o motor de autocompletar (autocomplete) com navegação cíclica por TAB.
package cli

import (
	"os"
	"path/filepath"
	"sort"
	"strings"
	"unicode"
)

var (
	// Lista ordenada de comandos principais para autocompletar
	allCommandsList = []string{
		"APPEND",
		"APPEND BLANK",
		"BROWSE",
		"CLEAR",
		"CLS",
		"CONTINUE",
		"COPY",
		"COUNT",
		"CREATE",
		"DELETE",
		"DISPLAY",
		"DO",
		"EDIT",
		"FIND",
		"GO",
		"GOTO",
		"HELP",
		"INDEX",
		"LIST",
		"LOCATE",
		"MODIFY",
		"PACK",
		"QUIT",
		"RECALL",
		"REPLACE",
		"RESTORE",
		"SAVE",
		"SELECT",
		"SET",
		"SKIP",
		"SORT",
		"STORE",
		"SUM",
		"TOTAL",
		"USE",
		"ZAP",
	}

	// Tópicos adicionais para sugestão após o comando HELP
	helpTopicsList = []string{
		"COMMANDS",
		"COMANDOS",
		"EDITOR",
		"MED",
		"FUNÇÕES",
		"FUNCOES",
		"TELA-CHEIA",
		"ERROS",
		"LIMITES",
		"UTILITÁRIOS",
		"UTILITARIOS",
	}

	// Parâmetros de SET
	setParamsList = []string{
		"TALK",
		"BELL",
		"ECHO",
		"EXACT",
		"HEADING",
		"INTENSITY",
		"SAFETY",
		"STATUS",
		"DEFAULT",
		"DEVICE",
		"ON",
		"OFF",
	}
)

// AutocompleteState gerencia o estado da iteração de TABs sucessivos.
type AutocompleteState struct {
	OriginalLine string
	Prefix       string
	Candidates   []string
	Index        int
	ReplaceStart int
	ReplaceEnd   int
	Active       bool
}

// Reset reinicia o estado de autocompletar.
func (a *AutocompleteState) Reset() {
	a.OriginalLine = ""
	a.Prefix = ""
	a.Candidates = nil
	a.Index = 0
	a.ReplaceStart = 0
	a.ReplaceEnd = 0
	a.Active = false
}

// FindCandidates procura todas as palavras candidatas para o prefixo no contexto atual.
func FindCandidates(line string, cursorPos int) (prefix string, candidates []string, replaceStart int, replaceEnd int) {
	if cursorPos > len(line) {
		cursorPos = len(line)
	}
	left := line[:cursorPos]

	// Encontra o início da palavra atual
	start := cursorPos
	for start > 0 && !unicode.IsSpace(rune(line[start-1])) && line[start-1] != ';' {
		start--
	}

	prefix = left[start:cursorPos]
	replaceStart = start
	replaceEnd = cursorPos

	upperPrefix := strings.ToUpper(prefix)
	beforeWord := strings.TrimSpace(left[:start])
	tokensBefore := strings.Fields(beforeWord)

	// Contexto 1: Primeiro comando da linha ou após ponto e vírgula
	if len(tokensBefore) == 0 || strings.HasSuffix(beforeWord, ";") {
		for _, cmd := range allCommandsList {
			if strings.HasPrefix(cmd, upperPrefix) {
				candidates = append(candidates, cmd)
			}
		}
		sort.Strings(candidates)
		return
	}

	firstVerb := strings.ToUpper(tokensBefore[0])

	// Contexto 2: Após HELP
	if firstVerb == "HELP" {
		seen := make(map[string]bool)
		for _, t := range helpTopicsList {
			if strings.HasPrefix(t, upperPrefix) && !seen[t] {
				candidates = append(candidates, t)
				seen[t] = true
			}
		}
		for _, cmd := range allCommandsList {
			if strings.HasPrefix(cmd, upperPrefix) && !seen[cmd] {
				candidates = append(candidates, cmd)
				seen[cmd] = true
			}
		}
		sort.Strings(candidates)
		return
	}

	// Contexto 3: Após SET
	if firstVerb == "SET" {
		for _, sp := range setParamsList {
			if strings.HasPrefix(sp, upperPrefix) {
				candidates = append(candidates, sp)
			}
		}
		sort.Strings(candidates)
		return
	}

	// Contexto 4: Após USE / DO (arquivos no diretório corrente)
	if firstVerb == "USE" || firstVerb == "DO" {
		extFilter := ".DBF"
		if firstVerb == "DO" {
			extFilter = ".CMD"
		}
		files, _ := os.ReadDir(".")
		for _, f := range files {
			if !f.IsDir() {
				name := f.Name()
				if strings.EqualFold(filepath.Ext(name), extFilter) || filepath.Ext(name) == "" {
					upperName := strings.ToUpper(name)
					if strings.HasPrefix(upperName, upperPrefix) {
						candidates = append(candidates, upperName)
					}
				}
			}
		}
		if len(candidates) > 0 {
			sort.Strings(candidates)
			return
		}
	}

	// Fallback geral: busca em todos os comandos e palavras-chave
	for _, cmd := range allCommandsList {
		if strings.HasPrefix(cmd, upperPrefix) {
			candidates = append(candidates, cmd)
		}
	}
	sort.Strings(candidates)
	return
}

// NextCompletion calcula a próxima linha e posição do cursor ao pressionar TAB.
func (a *AutocompleteState) NextCompletion(line string, cursorPos int) (newLine string, newCursorPos int, success bool) {
	// Se já está ativo e a linha atual corresponde ao candidato anterior, circula para o próximo
	if a.Active && len(a.Candidates) > 1 {
		a.Index = (a.Index + 1) % len(a.Candidates)
		candidate := a.Candidates[a.Index]

		// Substitui a palavra anterior pelo novo candidato
		newLine = a.OriginalLine[:a.ReplaceStart] + candidate + a.OriginalLine[a.ReplaceEnd:]
		newCursorPos = a.ReplaceStart + len(candidate)
		return newLine, newCursorPos, true
	}

	// Primeira pressão de TAB para esta palavra
	prefix, candidates, replaceStart, replaceEnd := FindCandidates(line, cursorPos)
	if len(candidates) == 0 {
		return line, cursorPos, false
	}

	a.OriginalLine = line
	a.Prefix = prefix
	a.Candidates = candidates
	a.Index = 0
	a.ReplaceStart = replaceStart
	a.ReplaceEnd = replaceEnd
	a.Active = true

	candidate := candidates[0]

	// Se houver apenas 1 candidato, adiciona espaço no final para agilizar a digitação
	suffixSpace := ""
	if len(candidates) == 1 {
		// Se já não houver espaço depois
		if replaceEnd >= len(line) || line[replaceEnd] != ' ' {
			suffixSpace = " "
		}
	}

	newLine = line[:replaceStart] + candidate + suffixSpace + line[replaceEnd:]
	newCursorPos = replaceStart + len(candidate) + len(suffixSpace)
	return newLine, newCursorPos, true
}
