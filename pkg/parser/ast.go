// Package parser implementa a análise léxica e sintática dos comandos da linguagem dBASE II.
package parser

import "fmt"

// CommandType identifica o comando dBASE II analisado.
type CommandType string

const (
	CmdUse       CommandType = "USE"
	CmdCreate    CommandType = "CREATE"
	CmdList      CommandType = "LIST"
	CmdDisplay   CommandType = "DISPLAY"
	CmdAppend    CommandType = "APPEND"
	CmdEdit      CommandType = "EDIT"
	CmdBrowse    CommandType = "BROWSE"
	CmdDelete    CommandType = "DELETE"
	CmdRecall    CommandType = "RECALL"
	CmdPack      CommandType = "PACK"
	CmdZap       CommandType = "ZAP"
	CmdReplace   CommandType = "REPLACE"
	CmdLocate    CommandType = "LOCATE"
	CmdContinue  CommandType = "CONTINUE"
	CmdCount     CommandType = "COUNT"
	CmdSum       CommandType = "SUM"
	CmdStore     CommandType = "STORE"
	CmdPrint     CommandType = "?"
	CmdPrintSame CommandType = "??"
	CmdGoto      CommandType = "GOTO"
	CmdSkip      CommandType = "SKIP"
	CmdDo        CommandType = "DO"
	CmdSet       CommandType = "SET"
	CmdQuit      CommandType = "QUIT"
	CmdHelp      CommandType = "HELP"
	CmdClear     CommandType = "CLEAR"
	CmdUnknown   CommandType = "UNKNOWN"
)

// Statement representa um comando estruturado na AST do interpretador dBASE II.
type Statement struct {
	Type        CommandType       `json:"type"`
	Target      string            `json:"target"`       // Ex: nome do arquivo no USE / CREATE / DO, ou variável no STORE
	Fields      []string          `json:"fields"`       // Lista de campos (LIST NOME, VALOR)
	Scope       string            `json:"scope"`        // ALL, RECORD n, NEXT n, REST
	ForClause   string            `json:"for_clause"`   // Expressão da cláusula FOR
	WhileClause string            `json:"while_clause"` // Expressão da cláusula WHILE
	WithClause  string            `json:"with_clause"`  // Expressão da cláusula WITH (REPLACE)
	Expression  string            `json:"expression"`   // Expressão avaliada no ? ou STORE
	SetParam    string            `json:"set_param"`    // Parâmetro do SET (ex: TALK, ECHO)
	SetValue    string            `json:"set_value"`    // Valor do SET (ex: ON, OFF)
	RecNo       int               `json:"rec_no"`       // Número do registro (ex: EDIT 5, GOTO 10)
	Raw         string            `json:"raw"`          // Comando original bruto
	Options     map[string]string `json:"options"`      // Opções extras
}

// FormatSummary retorna uma representação legível do comando analisado.
func (s *Statement) FormatSummary() string {
	return fmt.Sprintf("Cmd: %s | Target: %s | Scope: %s | For: %s | While: %s", s.Type, s.Target, s.Scope, s.ForClause, s.WhileClause)
}
