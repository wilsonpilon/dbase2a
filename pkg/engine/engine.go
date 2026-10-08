// Package engine implementa o motor de execução de comandos dBASE II,
// gerenciando tabelas ativas, variáveis de ambiente, buffers e ponteiros de registros.
package engine

import (
	"bytes"
	"fmt"
	"io"
	"os"
	"strings"
	"sync"
	"time"

	"dbase2a/pkg/config"
	"dbase2a/pkg/dbf"
	"dbase2a/pkg/parser"
	"dbase2a/pkg/storage/sqlite"
)

// Engine gerencia o estado de execução de uma sessão do dBase II+.
type Engine struct {
	Config       *config.AppConfig
	Storage      *sqlite.Storage
	ActiveTable  *dbf.Table
	ActiveAlias  string
	Out          io.Writer
	LastLocateCond string
	QuitRequested bool
	mu           sync.Mutex
}

// NewEngine instancia um novo motor de execução com a configuração e o armazenamento fornecidos.
func NewEngine(cfg *config.AppConfig, store *sqlite.Storage, out io.Writer) *Engine {
	if out == nil {
		out = os.Stdout
	}
	return &Engine{
		Config:  cfg,
		Storage: store,
		Out:     out,
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
