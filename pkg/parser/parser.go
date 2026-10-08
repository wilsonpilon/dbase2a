package parser

import (
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
	if strings.HasPrefix(input, ".") && !strings.HasPrefix(input, ".T.") && !strings.HasPrefix(input, ".F.") && !strings.HasPrefix(input, ".NOT.") && !strings.HasPrefix(input, ".AND.") && !strings.HasPrefix(input, ".OR.") {
		input = strings.TrimSpace(input[1:])
	}

	// Comentários no dBASE II iniciam com '*' ou 'NOTE'
	if strings.HasPrefix(input, "*") || strings.HasPrefix(strings.ToUpper(input), "NOTE ") || strings.ToUpper(input) == "NOTE" {
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
		if stmt != nil {
			statements = append(statements, stmt)
		}
	}

	return statements, nil
}

func parseSingleStatement(raw string) (*Statement, error) {
	rawTrimmed := strings.TrimSpace(raw)
	if rawTrimmed == "" {
		return nil, nil
	}

	// Verificação direta para comandos de impressão ? e ??
	if strings.HasPrefix(rawTrimmed, "??") {
		return &Statement{
			Type:       CmdPrintSame,
			Expression: strings.TrimSpace(rawTrimmed[2:]),
			Raw:        raw,
			Options:    make(map[string]string),
		}, nil
	}
	if strings.HasPrefix(rawTrimmed, "?") {
		return &Statement{
			Type:       CmdPrint,
			Expression: strings.TrimSpace(rawTrimmed[1:]),
			Raw:        raw,
			Options:    make(map[string]string),
		}, nil
	}

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

	// Checagem de atribuição direta: VAR = EXPR
	if len(tokens) >= 3 && (tokens[1].Type == TokenEquals || tokens[1].Value == "=") {
		varName := strings.ToUpper(tokens[0].Value)
		var exprParts []string
		for _, tok := range tokens[2:] {
			exprParts = append(exprParts, tok.Raw)
		}
		return &Statement{
			Type:       CmdStore,
			Target:     varName,
			Expression: strings.Join(exprParts, " "),
			Raw:        raw,
			Options:    make(map[string]string),
		}, nil
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
		parseClauses(tokens[1:], stmt, true)
	case "DISPLAY":
		stmt.Type = CmdDisplay
		parseClauses(tokens[1:], stmt, true)
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
		parseClauses(tokens[1:], stmt, false)
	case "RECALL":
		stmt.Type = CmdRecall
		parseClauses(tokens[1:], stmt, false)
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
		parseClauses(tokens[1:], stmt, false)
	case "SUM":
		stmt.Type = CmdSum
		parseSum(tokens[1:], stmt)
	case "STORE":
		stmt.Type = CmdStore
		parseStore(tokens[1:], stmt)
	case "GO", "GOTO":
		stmt.Type = CmdGoto
		if len(tokens) > 1 {
			target := strings.ToUpper(tokens[1].Value)
			if target == "TOP" || target == "BOTTOM" {
				stmt.Target = target
			} else if n, err := strconv.Atoi(target); err == nil {
				stmt.RecNo = n
			}
		}
	case "SKIP":
		stmt.Type = CmdSkip
		stmt.RecNo = 1 // default skip +1
		if len(tokens) > 1 {
			if n, err := strconv.Atoi(tokens[1].Value); err == nil {
				stmt.RecNo = n
			}
		}
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
	case "HELP":
		stmt.Type = CmdHelp
		if len(tokens) > 1 {
			var parts []string
			for _, t := range tokens[1:] {
				parts = append(parts, t.Raw)
			}
			stmt.Target = strings.TrimSpace(strings.Join(parts, " "))
		}
	case "CLEAR", "CLS":
		stmt.Type = CmdClear
	default:
		// Se começar por número inteiro, no dBase II '5' equivale a 'GOTO 5'
		if num, err := strconv.Atoi(verb); err == nil {
			stmt.Type = CmdGoto
			stmt.RecNo = num
			return stmt, nil
		}
		stmt.Type = CmdUnknown
		stmt.Target = verb
	}

	return stmt, nil
}

// parseClauses analisa parâmetros genéricos de comandos: escopo (ALL, RECORD n, NEXT n, REST), campos, FOR e WHILE.
func parseClauses(tokens []Token, stmt *Statement, allowFields bool) {
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
			stmt.ForClause, i = extractClauseUntilNextKeyword(tokens, i+1)
			continue
		}
		if tokVal == "WHILE" {
			stmt.WhileClause, i = extractClauseUntilNextKeyword(tokens, i+1)
			continue
		}
		if tokVal == "TO" && i+1 < len(tokens) {
			stmt.Target = strings.ToUpper(tokens[i+1].Value)
			i++
			continue
		}
		if allowFields && tokens[i].Type != TokenComma {
			stmt.Fields = append(stmt.Fields, tokVal)
		}
	}
}

func parseReplace(tokens []Token, stmt *Statement) {
	if len(tokens) == 0 {
		return
	}

	// REPLACE [scope] FIELD WITH expr [FOR expr] [WHILE expr]
	idx := 0
	if strings.ToUpper(tokens[idx].Value) == "ALL" {
		stmt.Scope = "ALL"
		idx++
	}
	if idx >= len(tokens) {
		return
	}

	stmt.Target = strings.ToUpper(tokens[idx].Value) // Campo alvo
	idx++

	for idx < len(tokens) {
		tokVal := strings.ToUpper(tokens[idx].Value)
		if tokVal == "WITH" {
			stmt.WithClause, idx = extractClauseUntilNextKeyword(tokens, idx+1)
			continue
		}
		if tokVal == "FOR" {
			stmt.ForClause, idx = extractClauseUntilNextKeyword(tokens, idx+1)
			continue
		}
		if tokVal == "WHILE" {
			stmt.WhileClause, idx = extractClauseUntilNextKeyword(tokens, idx+1)
			continue
		}
		idx++
	}
}

func parseLocate(tokens []Token, stmt *Statement) {
	for i := 0; i < len(tokens); i++ {
		tokVal := strings.ToUpper(tokens[i].Value)
		if tokVal == "FOR" {
			stmt.ForClause, i = extractClauseUntilNextKeyword(tokens, i+1)
			continue
		}
		if tokVal == "WHILE" {
			stmt.WhileClause, i = extractClauseUntilNextKeyword(tokens, i+1)
			continue
		}
	}
}

func parseSum(tokens []Token, stmt *Statement) {
	for i := 0; i < len(tokens); i++ {
		tokVal := strings.ToUpper(tokens[i].Value)
		if tokVal == "FOR" {
			stmt.ForClause, i = extractClauseUntilNextKeyword(tokens, i+1)
			continue
		}
		if tokVal == "WHILE" {
			stmt.WhileClause, i = extractClauseUntilNextKeyword(tokens, i+1)
			continue
		}
		if tokVal == "TO" && i+1 < len(tokens) {
			stmt.Target = strings.ToUpper(tokens[i+1].Value)
			i++
			continue
		}
		if tokVal == "ALL" {
			stmt.Scope = "ALL"
			continue
		}
		if tokens[i].Type != TokenComma {
			stmt.Fields = append(stmt.Fields, tokVal)
		}
	}
}

func parseStore(tokens []Token, stmt *Statement) {
	// STORE <expr> TO <var>
	var exprParts []string
	toFound := false
	for i := 0; i < len(tokens); i++ {
		tokVal := strings.ToUpper(tokens[i].Value)
		if tokVal == "TO" {
			toFound = true
			if i+1 < len(tokens) {
				stmt.Target = strings.ToUpper(tokens[i+1].Value)
			}
			break
		}
		exprParts = append(exprParts, tokens[i].Raw)
	}
	if toFound {
		stmt.Expression = strings.Join(exprParts, " ")
	}
}

func extractClauseUntilNextKeyword(tokens []Token, start int) (string, int) {
	var parts []string
	curr := start
	for curr < len(tokens) {
		upper := strings.ToUpper(tokens[curr].Value)
		if upper == "FOR" || upper == "WHILE" || upper == "WITH" || upper == "TO" {
			curr--
			break
		}
		parts = append(parts, tokens[curr].Raw)
		curr++
	}
	return strings.Join(parts, " "), curr
}
