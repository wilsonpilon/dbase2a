// Package sqlite implementa a camada de armazenamento relacional e persistência
// de configurações, histórico de comandos, catálogo de metadados DBF e logs de builds MSX.
package sqlite

import (
	"database/sql"
	"fmt"
	"os"
	"path/filepath"
	"sync"
	"time"

	_ "modernc.org/sqlite"
)

// Storage gerencia a conexão com o banco de dados SQLite do dBase II++.
type Storage struct {
	db   *sql.DB
	path string
	mu   sync.RWMutex
}

// CommandLogEntry representa um comando executado pelo usuário no CLI, TUI ou GUI.
type CommandLogEntry struct {
	ID         int64     `json:"id"`
	Command    string    `json:"command"`
	ExecutedAt time.Time `json:"executed_at"`
	DurationMs int64     `json:"duration_ms"`
	Success    bool      `json:"success"`
	ErrorMsg   string    `json:"error_msg"`
}

// TableCatalogEntry armazena informações catalogadas sobre tabelas .DBF no disco.
type TableCatalogEntry struct {
	ID          int64     `json:"id"`
	TableName   string    `json:"table_name"`
	FilePath    string    `json:"file_path"`
	RecordCount int       `json:"record_count"`
	FieldCount  int       `json:"field_count"`
	RecordSize  int       `json:"record_size"`
	LastUpdate  time.Time `json:"last_update"`
}

// MSXBuildLog registra o histórico de compilações geradas para a plataforma MSX.
type MSXBuildLog struct {
	ID          int64     `json:"id"`
	ProjectName string    `json:"project_name"`
	Target      string    `json:"target"`
	SDK         string    `json:"sdk"`
	OutputPath  string    `json:"output_path"`
	CreatedAt   time.Time `json:"created_at"`
	Status      string    `json:"status"`
	LogOutput   string    `json:"log_output"`
}

// Open inicializa a conexão com o banco de dados SQLite e aplica as migrações de esquema.
func Open(dbPath string) (*Storage, error) {
	if dbPath == "" {
		dbPath = "dbase2plus.db"
	}
	if dir := filepath.Dir(dbPath); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}

	db, err := sql.Open("sqlite", dbPath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir banco sqlite (%s): %w", dbPath, err)
	}

	s := &Storage{
		db:   db,
		path: dbPath,
	}

	if err := s.migrate(); err != nil {
		_ = db.Close()
		return nil, fmt.Errorf("erro ao aplicar migrações no banco sqlite: %w", err)
	}

	return s, nil
}

// Close finaliza com segurança a conexão com o banco SQLite.
func (s *Storage) Close() error {
	s.mu.Lock()
	defer s.mu.Unlock()
	if s.db != nil {
		return s.db.Close()
	}
	return nil
}

// migrate cria as tabelas necessárias no SQLite caso ainda não existam.
func (s *Storage) migrate() error {
	schema := `
	CREATE TABLE IF NOT EXISTS config_store (
		key TEXT PRIMARY KEY,
		value TEXT NOT NULL,
		updated_at DATETIME DEFAULT CURRENT_TIMESTAMP
	);

	CREATE TABLE IF NOT EXISTS command_history (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		command TEXT NOT NULL,
		executed_at DATETIME NOT NULL,
		duration_ms INTEGER DEFAULT 0,
		success INTEGER DEFAULT 1,
		error_msg TEXT DEFAULT ''
	);

	CREATE TABLE IF NOT EXISTS table_catalog (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		table_name TEXT NOT NULL,
		file_path TEXT NOT NULL UNIQUE,
		record_count INTEGER DEFAULT 0,
		field_count INTEGER DEFAULT 0,
		record_size INTEGER DEFAULT 0,
		last_update DATETIME NOT NULL
	);

	CREATE TABLE IF NOT EXISTS msx_builds (
		id INTEGER PRIMARY KEY AUTOINCREMENT,
		project_name TEXT NOT NULL,
		target TEXT NOT NULL,
		sdk TEXT NOT NULL,
		output_path TEXT NOT NULL,
		created_at DATETIME NOT NULL,
		status TEXT NOT NULL,
		log_output TEXT DEFAULT ''
	);
	`
	_, err := s.db.Exec(schema)
	return err
}

// SetConfig grava ou atualiza um parâmetro de configuração no banco.
func (s *Storage) SetConfig(key, value string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	query := `INSERT INTO config_store(key, value, updated_at) VALUES (?, ?, CURRENT_TIMESTAMP)
	          ON CONFLICT(key) DO UPDATE SET value = excluded.value, updated_at = CURRENT_TIMESTAMP`
	_, err := s.db.Exec(query, key, value)
	return err
}

// GetConfig recupera o valor de uma configuração armazenada.
func (s *Storage) GetConfig(key string) (string, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	var val string
	err := s.db.QueryRow(`SELECT value FROM config_store WHERE key = ?`, key).Scan(&val)
	if err != nil {
		return "", err
	}
	return val, nil
}

// LogCommand insere uma entrada no histórico de comandos executados.
func (s *Storage) LogCommand(cmd string, dur time.Duration, success bool, errStr string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	query := `INSERT INTO command_history(command, executed_at, duration_ms, success, error_msg) VALUES (?, ?, ?, ?, ?)`
	succInt := 1
	if !success {
		succInt = 0
	}
	_, err := s.db.Exec(query, cmd, time.Now(), dur.Milliseconds(), succInt, errStr)
	return err
}

// GetRecentCommands busca os últimos comandos executados pelo usuário.
func (s *Storage) GetRecentCommands(limit int) ([]CommandLogEntry, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if limit <= 0 {
		limit = 50
	}
	rows, err := s.db.Query(`SELECT id, command, executed_at, duration_ms, success, error_msg 
	                         FROM command_history ORDER BY id DESC LIMIT ?`, limit)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var entries []CommandLogEntry
	for rows.Next() {
		var e CommandLogEntry
		var succInt int
		if err := rows.Scan(&e.ID, &e.Command, &e.ExecutedAt, &e.DurationMs, &succInt, &e.ErrorMsg); err != nil {
			return nil, err
		}
		e.Success = (succInt == 1)
		entries = append(entries, e)
	}
	return entries, nil
}

// CatalogTable cadastra ou atualiza os metadados de uma tabela .DBF no catálogo SQLite.
func (s *Storage) CatalogTable(name, path string, recCount, fieldCount, recSize int, lastUpdate time.Time) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	query := `INSERT INTO table_catalog(table_name, file_path, record_count, field_count, record_size, last_update)
	          VALUES (?, ?, ?, ?, ?, ?)
	          ON CONFLICT(file_path) DO UPDATE SET 
	             table_name = excluded.table_name,
	             record_count = excluded.record_count,
	             field_count = excluded.field_count,
	             record_size = excluded.record_size,
	             last_update = excluded.last_update`
	_, err := s.db.Exec(query, name, path, recCount, fieldCount, recSize, lastUpdate)
	return err
}

// LogMSXBuild registra um build gerado para a plataforma MSX.
func (s *Storage) LogMSXBuild(project, target, sdk, outPath, status, logOutput string) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	query := `INSERT INTO msx_builds(project_name, target, sdk, output_path, created_at, status, log_output)
	          VALUES (?, ?, ?, ?, ?, ?, ?)`
	_, err := s.db.Exec(query, project, target, sdk, outPath, time.Now(), status, logOutput)
	return err
}
