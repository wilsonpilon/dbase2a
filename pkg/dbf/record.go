package dbf

import (
	"fmt"
	"strings"
)

// Record representa uma linha de dados de uma tabela dBASE II.
type Record struct {
	Deleted bool              `json:"deleted"` // true se o registro estiver marcado para exclusão ('*')
	Values  map[string]string `json:"values"`  // Mapa de nome do campo (em maiúsculas) para valor texto
}

// NewRecord cria uma nova instância de registro em branco associada aos campos informados.
func NewRecord(fields []Field) *Record {
	r := &Record{
		Deleted: false,
		Values:  make(map[string]string),
	}
	for _, f := range fields {
		name := strings.ToUpper(strings.TrimSpace(f.Name))
		r.Values[name] = ""
	}
	return r
}

// Get obtém o valor do campo como string formatada.
func (r *Record) Get(fieldName string) string {
	if r.Values == nil {
		return ""
	}
	return r.Values[strings.ToUpper(strings.TrimSpace(fieldName))]
}

// Set define o valor de um campo.
func (r *Record) Set(fieldName, val string) {
	if r.Values == nil {
		r.Values = make(map[string]string)
	}
	r.Values[strings.ToUpper(strings.TrimSpace(fieldName))] = val
}

// FormatRecordBytes converte o registro para os bytes binários do dBASE II com base nos descritores de campo.
func (r *Record) FormatRecordBytes(fields []Field, totalSize int) []byte {
	buf := make([]byte, totalSize)

	// Marcador de exclusão no byte 0
	if r.Deleted {
		buf[0] = '*' // 0x2A
	} else {
		buf[0] = ' ' // 0x20
	}

	for _, f := range fields {
		val := r.Get(f.Name)
		fLen := int(f.Length)
		fOffset := f.Offset

		var fieldBytes []byte
		switch f.Type {
		case TypeNumeric:
			// Números no dBASE II são alinhados à direita com espaços à esquerda
			val = strings.TrimSpace(val)
			if len(val) > fLen {
				val = val[:fLen]
			}
			fieldBytes = []byte(fmt.Sprintf("%*s", fLen, val))
		case TypeLogical:
			val = strings.TrimSpace(strings.ToUpper(val))
			if val == "T" || val == "Y" || val == "1" || val == "TRUE" {
				fieldBytes = []byte("T")
			} else if val == "F" || val == "N" || val == "0" || val == "FALSE" {
				fieldBytes = []byte("F")
			} else {
				fieldBytes = []byte("?")
			}
		default: // TypeCharacter
			// Texto no dBASE II é alinhado à esquerda com espaços à direita
			if len(val) > fLen {
				val = val[:fLen]
			}
			fieldBytes = []byte(fmt.Sprintf("%-*s", fLen, val))
		}

		copy(buf[fOffset:fOffset+fLen], fieldBytes)
	}

	return buf
}

// ParseRecordBytes reconstrói um registro a partir dos bytes lidos do arquivo .DBF.
func ParseRecordBytes(buf []byte, fields []Field) (*Record, error) {
	if len(buf) == 0 {
		return nil, fmt.Errorf("buffer de registro vazio")
	}

	deleted := (buf[0] == '*')

	rec := &Record{
		Deleted: deleted,
		Values:  make(map[string]string),
	}

	for _, f := range fields {
		fOffset := f.Offset
		fLen := int(f.Length)

		if fOffset+fLen > len(buf) {
			return nil, fmt.Errorf("leitura de campo '%s' ultrapassa o tamanho do registro", f.Name)
		}

		rawVal := string(buf[fOffset : fOffset+fLen])
		name := strings.ToUpper(strings.TrimSpace(f.Name))

		switch f.Type {
		case TypeNumeric:
			rec.Values[name] = strings.TrimSpace(rawVal)
		case TypeLogical:
			rec.Values[name] = strings.TrimSpace(rawVal)
		default:
			rec.Values[name] = strings.TrimRight(rawVal, " ")
		}
	}

	return rec, nil
}
