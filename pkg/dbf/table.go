package dbf

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"
)

// Table gerencia uma tabela dBASE II aberta em memória e em disco.
type Table struct {
	FilePath string    `json:"file_path"`
	Header   *Header   `json:"header"`
	Records  []*Record `json:"records"`
	RecNo    int       `json:"rec_no"` // Ponteiro do registro atual (1-based no dBASE II)
}

// CreateTable inicializa uma nova tabela dBASE II com a lista de campos especificada e salva no disco.
func CreateTable(filePath string, fields []Field) (*Table, error) {
	if len(fields) == 0 {
		return nil, fmt.Errorf("a tabela deve possuir ao menos 1 campo")
	}
	if len(fields) > MaxFields {
		return nil, fmt.Errorf("a tabela ultrapassa o limite de %d campos do dBASE II", MaxFields)
	}

	offset := 1 // Byte 0 reservado para o marcador de exclusão (' ' ou '*')
	for i := range fields {
		fields[i].Offset = offset
		if err := fields[i].Validate(); err != nil {
			return nil, err
		}
		offset += int(fields[i].Length)
	}

	totalRecordSize := uint16(offset)

	hdr := &Header{
		Version:     DBase2VersionByte,
		RecordCount: 0,
		LastUpdate:  time.Now(),
		RecordSize:  totalRecordSize,
		Fields:      fields,
	}

	tbl := &Table{
		FilePath: filePath,
		Header:   hdr,
		Records:  make([]*Record, 0),
		RecNo:    0,
	}

	if err := tbl.Save(); err != nil {
		return nil, fmt.Errorf("falha ao salvar nova tabela .DBF: %w", err)
	}

	return tbl, nil
}

// OpenTable abre e carrega um arquivo .DBF existente no formato dBASE II.
func OpenTable(filePath string) (*Table, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return nil, fmt.Errorf("erro ao abrir arquivo .DBF (%s): %w", filePath, err)
	}
	defer file.Close()

	hdr, err := ReadHeader(file)
	if err != nil {
		return nil, err
	}

	records := make([]*Record, 0, hdr.RecordCount)
	recBuf := make([]byte, hdr.RecordSize)

	for i := 0; i < int(hdr.RecordCount); i++ {
		_, err := io.ReadFull(file, recBuf)
		if err != nil {
			if err == io.EOF || err == io.ErrUnexpectedEOF {
				break
			}
			return nil, fmt.Errorf("erro ao ler registro %d: %w", i+1, err)
		}

		// Se encontrar marcador de final de arquivo CP/M 0x1A antes do registro
		if recBuf[0] == FileTerminator {
			break
		}

		rec, err := ParseRecordBytes(recBuf, hdr.Fields)
		if err != nil {
			return nil, fmt.Errorf("erro ao decodificar registro %d: %w", i+1, err)
		}
		records = append(records, rec)
	}

	tbl := &Table{
		FilePath: filePath,
		Header:   hdr,
		Records:  records,
		RecNo:    0,
	}

	if len(records) > 0 {
		tbl.RecNo = 1
	}

	return tbl, nil
}

// Save grava todas as alterações no arquivo .DBF no disco no formato nativo do dBASE II.
func (t *Table) Save() error {
	if t.FilePath == "" {
		return fmt.Errorf("caminho de arquivo .DBF indefinido")
	}

	if dir := filepath.Dir(t.FilePath); dir != "." && dir != "" {
		_ = os.MkdirAll(dir, 0755)
	}

	file, err := os.Create(t.FilePath)
	if err != nil {
		return fmt.Errorf("erro ao criar arquivo .DBF (%s): %w", t.FilePath, err)
	}
	defer file.Close()

	t.Header.RecordCount = uint16(len(t.Records))
	t.Header.LastUpdate = time.Now()

	if err := t.Header.WriteHeader(file); err != nil {
		return fmt.Errorf("erro ao gravar cabeçalho: %w", err)
	}

	for _, r := range t.Records {
		recBytes := r.FormatRecordBytes(t.Header.Fields, int(t.Header.RecordSize))
		if _, err := file.Write(recBytes); err != nil {
			return fmt.Errorf("erro ao gravar registro: %w", err)
		}
	}

	// Gravador do terminador End-Of-File (0x1A / Ctrl+Z)
	if _, err := file.Write([]byte{FileTerminator}); err != nil {
		return fmt.Errorf("erro ao gravar terminador de arquivo: %w", err)
	}

	return nil
}

// Append insere um novo registro de dados na tabela.
func (t *Table) Append(rec *Record) error {
	t.Records = append(t.Records, rec)
	t.Header.RecordCount = uint16(len(t.Records))
	t.RecNo = len(t.Records)
	return nil
}

// AppendBlank insere um registro em branco e posiciona o ponteiro nele.
func (t *Table) AppendBlank() *Record {
	rec := NewRecord(t.Header.Fields)
	_ = t.Append(rec)
	return rec
}

// CurrentRecord retorna o registro apontado por RecNo ou nil caso esteja fora dos limites.
func (t *Table) CurrentRecord() *Record {
	if t.RecNo < 1 || t.RecNo > len(t.Records) {
		return nil
	}
	return t.Records[t.RecNo-1]
}

// Goto posiciona o ponteiro no registro indicado (1-based).
func (t *Table) Goto(recNo int) error {
	if recNo < 1 || recNo > len(t.Records) {
		return fmt.Errorf("registro %d fora dos limites (1 a %d)", recNo, len(t.Records))
	}
	t.RecNo = recNo
	return nil
}

// GoTop posiciona no primeiro registro.
func (t *Table) GoTop() {
	if len(t.Records) > 0 {
		t.RecNo = 1
	} else {
		t.RecNo = 0
	}
}

// GoBottom posiciona no último registro.
func (t *Table) GoBottom() {
	if len(t.Records) > 0 {
		t.RecNo = len(t.Records)
	} else {
		t.RecNo = 0
	}
}

// Skip avança ou recua o ponteiro de registro relativo ao atual.
func (t *Table) Skip(n int) {
	newRec := t.RecNo + n
	if newRec < 1 {
		t.RecNo = 1
	} else if newRec > len(t.Records) {
		t.RecNo = len(t.Records) + 1 // EOF
	} else {
		t.RecNo = newRec
	}
}

// EOF retorna true se o ponteiro de registro ultrapassou o final da tabela.
func (t *Table) EOF() bool {
	return t.RecNo > len(t.Records) || len(t.Records) == 0
}

// BOF retorna true se o ponteiro de registro estiver antes do primeiro registro.
func (t *Table) BOF() bool {
	return t.RecNo <= 1
}

// Delete marca o registro atual como excluído.
func (t *Table) Delete() {
	if rec := t.CurrentRecord(); rec != nil {
		rec.Deleted = true
	}
}

// Recall desmarca a exclusão do registro atual.
func (t *Table) Recall() {
	if rec := t.CurrentRecord(); rec != nil {
		rec.Deleted = false
	}
}

// Pack remove fisicamente todos os registros marcados como excluídos e reindexa.
func (t *Table) Pack() int {
	var kept []*Record
	removed := 0
	for _, r := range t.Records {
		if r.Deleted {
			removed++
		} else {
			kept = append(kept, r)
		}
	}
	t.Records = kept
	t.Header.RecordCount = uint16(len(t.Records))
	if t.RecNo > len(t.Records) {
		t.RecNo = len(t.Records)
	}
	return removed
}

// Zap elimina todos os registros da tabela ativa.
func (t *Table) Zap() {
	t.Records = make([]*Record, 0)
	t.Header.RecordCount = 0
	t.RecNo = 0
}

// Close fecha a tabela e persiste dados se necessário.
func (t *Table) Close() error {
	return t.Save()
}

// FieldNames retorna a lista de nomes dos campos em maiúsculas.
func (t *Table) FieldNames() []string {
	names := make([]string, len(t.Header.Fields))
	for i, f := range t.Header.Fields {
		names[i] = strings.ToUpper(f.Name)
	}
	return names
}
