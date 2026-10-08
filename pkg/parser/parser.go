package parser

import (
	"fmt"
	"strconv"
	"strings"
)

// Parse processa uma string de entrada e retorna uma lista de instruções Statement.
func Parse(input string) ([]*Statement, error) {
	input = strings.TrimSpace(input)
	if input == "" {
		return nil, nil
	}

	// Remove eventual ponto '.' inicial característico do prompt do dBASE II
	if strings.HasPrefix(input, ".") {
		input = strings.TrimSpace(input[1:])
	}

	// Comentários no dBASE II iniciam com '*' ou 'NOTE'
	if strings.HasPrefix(input, "*") || strings.HasPrefix(strings.ToUpper(input), "NOTE ") {
		return nil, nil
	}

	// Suporte a múltiplos comandos separados por ponto-e-vírgula ';'
	rawStatements := strings.Split(input, ";")
	var statements []*Statement

	for _, rawStmt := range rawStatements {
		rawStmt = strings.TrimSpace(rawStmt)
		if rawStmt == "" {
			continue
		}

		stmt, err := parseSingleStatement(rawStmt)
		if err != nil {
			return nil, err
		}
		statements = append(statements, stmt)
	}

	return statements, nil
}

func parseSingleStatement(raw string) (*Statement, error) {
	lexer := NewLexer(raw)
	var tokens []Token
	for {
		t := lexer.NextToken()
		if t.Type == TokenEOF {
			break
		}
		tokens = append(tokens, t)
	}

	if len(tokens) == 0 {
		return nil, nil
	}

	verb := strings.ToUpper(tokens[0].Value)
	stmt := &Statement{
		Raw:     raw,
		Options: make(map[string]string),
	}

	switch verb {
	case "USE":
		stmt.Type = CmdUse
		if len(tokens) > 1 {
			stmt.Target = tokens[1].Value
		}
	case "CREATE":
		stmt.Type = CmdCreate
		if len(tokens) > 1 {
			stmt.Target = tokens[1].Value
		}
	case "LIST":
		stmt.Type = CmdList
		parseListOrDisplay(tokens[1:], stmt)
	case "DISPLAY":
		stmt.Type = CmdDisplay
		parseListOrDisplay(tokens[1:], stmt)
	case "APPEND":
		stmt.Type = CmdAppend
		if len(tokens) > 1 && strings.ToUpper(tokens[1].Value) == "BLANK" {
			stmt.Options["BLANK"] = "true"
		}
	case "EDIT":
		stmt.Type = CmdEdit
		if len(tokens) > 1 {
			if n, err := strconv.Atoi(tokens[1].Value); err == nil {
				stmt.RecNo = n
			}
		}
	case "BROWSE":
		stmt.Type = CmdBrowse
	case "DELETE":
		stmt.Type = CmdDelete
		parseScopeAndFor(tokens[1:], stmt)
	case "RECALL":
		stmt.Type = CmdRecall
		parseScopeAndFor(tokens[1:], stmt)
	case "PACK":
		stmt.Type = CmdPack
	case "ZAP":
		stmt.Type = CmdZap
	case "REPLACE":
		stmt.Type = CmdReplace
		parseReplace(tokens[1:], stmt)
	case "LOCATE":
		stmt.Type = CmdLocate
		parseLocate(tokens[1:], stmt)
	case "CONTINUE":
		stmt.Type = CmdContinue
	case "COUNT":
		stmt.Type = CmdCount
		parseScopeAndFor(tokens[1:], stmt)
	case "SUM":
		stmt.Type = CmdSum
		parseSum(tokens[1:], stmt)
	case "DO":
		stmt.Type = CmdDo
		if len(tokens) > 1 {
			stmt.Target = tokens[1].Value
		}
	case "SET":
		stmt.Type = CmdSet
		if len(tokens) > 1 {
			stmt.SetParam = strings.ToUpper(tokens[1].Value)
		}
		if len(tokens) > 2 {
			stmt.SetValue = strings.ToUpper(tokens[2].Value)
		}
	case "QUIT", "EXIT":
		stmt.Type = CmdQuit
	case "HELP", "?":
		stmt.Type = CmdHelp
	case "CLEAR", "CLS":
		stmt.Type = CmdClear
	default:
		stmt.Type = CmdUnknown
		stmt.Target = verb
	}

	return stmt, nil
}

func parseListOrDisplay(tokens []Token, stmt *Statement) {
	for i := 0; i < len(tokens); i++ {
		tokVal := strings.ToUpper(tokens[i].Value)
		if tokVal == "ALL" {
			stmt.Scope = "ALL"
			continue
		}
		if tokVal == "STRUCTURE" || tokVal == "STRUC" {
			stmt.Options["STRUCTURE"] = "true"
			continue
		}
		if tokVal == "FOR" {
			var forParts []string
			for j := i + 1; j < len(tokens); j++ {
				forParts = append(forParts, tokens[j].Value)
			}
			stmt.ForClause = strings.Join(forParts, " ")
			break
		}
		if tokens[i].Type != TokenComma {
			stmt.Fields = append(stmt.Fields, tokVal)
		}
	}
}

func parseScopeAndFor(tokens []Token, stmt *Statement) {
	for i := 0; i < len(tokens); i++ {
		tokVal := strings.ToUpper(tokens[i].Value)
		if tokVal == "ALL" {
			stmt.Scope = "ALL"
			continue
		}
		if tokVal == "FOR" {
			var forParts []string
			for j := i + 1; j < len(tokens); j++ {
				forParts = append(forParts, tokens[j].Value)
			}
			stmt.ForClause = strings.Join(forParts, " ")
			break
		}
	}
}

func parseReplace(tokens []Token, stmt *Statement) {
	if len(tokens) == 0 {
		return
	}
	stmt.Target = strings.ToUpper(tokens[0].Value) // Campo a ser substituído
	for i := 1; i < len(tokens); i++ {
		if strings.ToUpper(tokens[i].Value) == "WITH" {
			var withParts []string
			for j := i + 1; j < len(tokens); j++ {
				if strings.ToUpper(tokens[j].Value) == "FOR" {
					stmt.ForClause = strings.Join(extractRemaining(tokens, j+1), " ")
					break
				}
				withParts = append(withParts, tokens[j].Value)
			}
			stmt.WithClause = strings.Join(withParts, " ")
			break
		}
	}
}

func parseLocate(tokens []Token, stmt *Statement) {
	for i := 0; i < len(tokens); i++ {
		if strings.ToUpper(tokens[i].Value) == "FOR" {
			stmt.ForClause = strings.Join(extractRemaining(tokens, i+1), " ")
			break
		}
	}
}

func parseSum(tokens []Token, stmt *Statement) {
	for i := 0; i < len(tokens); i++ {
		tokVal := strings.ToUpper(tokens[i].Value)
		if tokVal == "FOR" {
			stmt.ForClause = strings.Join(extractRemaining(tokens, i+1), " ")
			break
		}
		if tokens[i].Type != TokenComma {
			stmt.Fields = append(stmt.Fields, tokVal)
		}
	}
}

func extractRemaining(tokens []Token, start int) []string {
	var parts []string
	for j := start; j < len(tokens); j++ {
		parts = append(parts, tokens[j].Value)
	}
	return parts
}

// FormatSummary retorna uma representação legível do comando analisado.
func (s *Statement) FormatSummary() string {
	return fmt.Sprintf("Cmd: %s | Target: %s | Scope: %s | For: %s", s.Type, s.Target, s.Scope, s.ForClause)
}
