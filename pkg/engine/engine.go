// Package engine implementa o motor de execução de comandos dBASE II,
// gerenciando tabelas ativas, variáveis de ambiente, buffers e ponteiros de registros.
package engine

import (
	"bufio"
	"bytes"
	"fmt"
	"io"
	"os"
	"strconv"
	"strings"
	"sync"
	"time"

	"dbase2a/pkg/config"
	"dbase2a/pkg/dbf"
	"dbase2a/pkg/eval"
	"dbase2a/pkg/help"
	"dbase2a/pkg/parser"
	"dbase2a/pkg/storage/sqlite"
)

// Engine gerencia o estado de execução de uma sessão do dBase II++.
type Engine struct {
	Config             *config.AppConfig
	Storage            *sqlite.Storage
	ActiveTable        *dbf.Table
	ActiveAlias        string
	CurrentRecordIndex int // 1-based (0 se tabela vazia ou sem seleção)
	Variables          map[string]eval.Value
	Help               *help.HelpSystem
	In                 io.Reader
	Out                io.Writer
	reader             *bufio.Reader
	LastLocateCond     string
	QuitRequested      bool
	mu                 sync.Mutex
}

// NewEngine instancia um novo motor de execução com a configuração e o armazenamento fornecidos.
func NewEngine(cfg *config.AppConfig, store *sqlite.Storage, out io.Writer) *Engine {
	if out == nil {
		out = os.Stdout
	}
	helpSys := help.NewHelpSystem()
	_ = helpSys.LoadDefault()

	return &Engine{
		Config:             cfg,
		Storage:            store,
		Variables:          make(map[string]eval.Value),
		CurrentRecordIndex: 1,
		Help:               helpSys,
		Out:                out,
	}
}

// Printf escreve texto formatado na saída configurada.
func (e *Engine) Printf(format string, a ...interface{}) {
	fmt.Fprintf(e.Out, format, a...)
}

// Println escreve uma linha na saída configurada.
func (e *Engine) Println(a ...interface{}) {
	fmt.Fprintln(e.Out, a...)
}

// GetEvalContext cria o contexto de avaliação de expressões no estado atual ou para um registro específico.
func (e *Engine) GetEvalContext(rec *dbf.Record, recNo int) *eval.Context {
	return &eval.Context{
		GetFieldFunc: func(name string) (eval.Value, bool) {
			if rec != nil {
				valStr := rec.Get(name)
				// Tenta inferir tipo com base no esquema do DBF se tabela ativa
				if e.ActiveTable != nil {
					for _, f := range e.ActiveTable.Header.Fields {
						if strings.EqualFold(f.Name, name) {
							switch f.Type {
							case dbf.TypeNumeric:
								if num, err := strconv.ParseFloat(strings.TrimSpace(valStr), 64); err == nil {
									return eval.NewNumberValue(num), true
								}
							case dbf.TypeLogical:
								upper := strings.ToUpper(strings.TrimSpace(valStr))
								return eval.NewLogicalValue(upper == "T" || upper == "Y"), true
							}
							return eval.NewStringValue(valStr), true
						}
					}
				}
				// Se não encontrou no header mas está no record
				if valStr != "" {
					if num, err := strconv.ParseFloat(strings.TrimSpace(valStr), 64); err == nil {
						return eval.NewNumberValue(num), true
					}
					return eval.NewStringValue(valStr), true
				}
			}
			return eval.NewNullValue(), false
		},
		GetVarFunc: func(name string) (eval.Value, bool) {
			if val, exists := e.Variables[strings.ToUpper(name)]; exists {
				return val, true
			}
			return eval.NewNullValue(), false
		},
		DeletedFunc: func() bool {
			if rec != nil {
				return rec.IsDeleted()
			}
			return false
		},
		EOFFunc: func() bool {
			if e.ActiveTable == nil || len(e.ActiveTable.Records) == 0 {
				return true
			}
			return recNo > len(e.ActiveTable.Records)
		},
		RecNoFunc: func() int {
			if recNo > 0 {
				return recNo
			}
			return e.CurrentRecordIndex
		},
	}
}

// ExecuteLine processa e executa uma linha completa de comando do usuário.
func (e *Engine) ExecuteLine(line string) error {
	e.mu.Lock()
	defer e.mu.Unlock()

	startTime := time.Now()
	stmts, err := parser.Parse(line)
	if err != nil {
		if e.Storage != nil {
			_ = e.Storage.LogCommand(line, time.Since(startTime), false, err.Error())
		}
		return err
	}

	for _, stmt := range stmts {
		if err := e.execStatement(stmt); err != nil {
			if e.Storage != nil {
				_ = e.Storage.LogCommand(line, time.Since(startTime), false, err.Error())
			}
			return err
		}
	}

	if e.Storage != nil && len(stmts) > 0 {
		_ = e.Storage.LogCommand(line, time.Since(startTime), true, "")
	}

	return nil
}

// ExecuteScript lê e executa linha por linha de um arquivo de script (.CMD ou .PRG).
func (e *Engine) ExecuteScript(scriptPath string) error {
	data, err := os.ReadFile(scriptPath)
	if err != nil {
		return fmt.Errorf("erro ao abrir script (%s): %w", scriptPath, err)
	}

	lines := strings.Split(string(data), "\n")
	for lineNum, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" || strings.HasPrefix(line, "*") || strings.HasPrefix(strings.ToUpper(line), "NOTE ") {
			continue
		}
		if e.Config.Echo {
			e.Printf("%s\n", line)
		}
		if err := e.ExecuteLine(line); err != nil {
			return fmt.Errorf("erro na linha %d de %s: %w", lineNum+1, scriptPath, err)
		}
		if e.QuitRequested {
			break
		}
	}
	return nil
}

// CaptureOutput helper para capturar a saída textual de um comando durante testes ou chamadas GUI.
func (e *Engine) CaptureOutput(fn func()) string {
	oldOut := e.Out
	buf := new(bytes.Buffer)
	e.Out = buf
	defer func() {
		e.Out = oldOut
	}()
	fn()
	return buf.String()
}
