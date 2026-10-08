package dbf

import (
	"encoding/binary"
	"fmt"
	"io"
	"time"
)

const (
	// DBase2VersionByte identifica o arquivo como dBASE II (0x02).
	DBase2VersionByte = 0x02

	// HeaderSize tamanho total fixo do cabeçalho dBASE II (521 bytes).
	HeaderSize = 521

	// MaxFields limite máximo de campos por tabela no padrão dBASE II (32 campos).
	MaxFields = 32

	// FieldDescriptorSize tamanho de cada entrada de campo no cabeçalho (16 bytes).
	FieldDescriptorSize = 16

	// HeaderTerminator terminador de cabeçalho (0x0D / Carriage Return).
	HeaderTerminator = 0x0D

	// FileTerminator marcador End-Of-File do CP/M / MSX-DOS (0x1A / Ctrl+Z).
	FileTerminator = 0x1A
)

// Header representa o cabeçalho binário de 521 bytes do dBASE II.
type Header struct {
	Version     byte      // 0x02
	RecordCount uint16    // Contagem de registros (máx 65535)
	LastUpdate  time.Time // YY MM DD
	RecordSize  uint16    // Tamanho do registro em bytes (incluindo marcador de deleção)
	Fields      []Field   // Lista de descritores de campo (até 32)
}

// ReadHeader lê e valida o cabeçalho binário de 521 bytes a partir de um leitor.
func ReadHeader(r io.Reader) (*Header, error) {
	buf := make([]byte, HeaderSize)
	if _, err := io.ReadFull(r, buf); err != nil {
		return nil, fmt.Errorf("falha ao ler cabeçalho do dBASE II: %w", err)
	}

	version := buf[0]
	if version != DBase2VersionByte {
		return nil, fmt.Errorf("versão inválida do dBASE: 0x%02X (esperado dBASE II: 0x02)", version)
	}

	recCount := binary.LittleEndian.Uint16(buf[1:3])
	year := int(buf[3])
	month := int(buf[4])
	day := int(buf[5])

	// Ajuste do ano (dBase II armazenava 2 dígitos de ano)
	fullYear := 1900 + year
	if fullYear < 1980 {
		fullYear += 100 // Ex: ano 26 -> 2026
	}
	if month < 1 || month > 12 {
		month = 1
	}
	if day < 1 || day > 31 {
		day = 1
	}
	lastUpdate := time.Date(fullYear, time.Month(month), day, 0, 0, 0, 0, time.UTC)

	recSize := binary.LittleEndian.Uint16(buf[6:8])

	var fields []Field
	offset := 1 // O primeiro byte de cada registro é o indicador de deleção

	for i := 0; i < MaxFields; i++ {
		fieldOffset := 8 + (i * FieldDescriptorSize)
		fieldBuf := buf[fieldOffset : fieldOffset+FieldDescriptorSize]

		// Se o primeiro byte do nome for nulo ou espaço, encerrou os campos válidos
		if fieldBuf[0] == 0x00 || fieldBuf[0] == 0x0D {
			break
		}

		// Nome do campo (até 10 caracteres ASCII)
		rawName := fieldBuf[0:10]
		nameLen := 0
		for nameLen < 10 && rawName[nameLen] != 0x00 {
			nameLen++
		}
		name := string(rawName[:nameLen])

		fType := FieldType(fieldBuf[11])
		fLen := fieldBuf[12]
		fDec := fieldBuf[15]

		field := Field{
			Name:     name,
			Type:     fType,
			Length:   fLen,
			Decimals: fDec,
			Offset:   offset,
		}

		if err := field.Validate(); err != nil {
			return nil, fmt.Errorf("campo %d inválido no cabeçalho: %w", i+1, err)
		}

		fields = append(fields, field)
		offset += int(fLen)
	}

	return &Header{
		Version:     version,
		RecordCount: recCount,
		LastUpdate:  lastUpdate,
		RecordSize:  recSize,
		Fields:      fields,
	}, nil
}

// WriteHeader grava o cabeçalho binário de 521 bytes no escritor informado.
func (h *Header) WriteHeader(w io.Writer) error {
	buf := make([]byte, HeaderSize)

	buf[0] = DBase2VersionByte
	binary.LittleEndian.PutUint16(buf[1:3], h.RecordCount)

	year := h.LastUpdate.Year() % 100
	buf[3] = byte(year)
	buf[4] = byte(h.LastUpdate.Month())
	buf[5] = byte(h.LastUpdate.Day())

	binary.LittleEndian.PutUint16(buf[6:8], h.RecordSize)

	for i, f := range h.Fields {
		if i >= MaxFields {
			break
		}
		fieldOffset := 8 + (i * FieldDescriptorSize)
		// Nome
		nameBytes := []byte(f.Name)
		copy(buf[fieldOffset:fieldOffset+10], nameBytes)
		// Tipo
		buf[fieldOffset+11] = byte(f.Type)
		// Tamanho
		buf[fieldOffset+12] = f.Length
		// Offset / endereço em runtime
		binary.LittleEndian.PutUint16(buf[fieldOffset+13:fieldOffset+15], uint16(f.Offset))
		// Decimais
		buf[fieldOffset+15] = f.Decimals
	}

	// Terminador de cabeçalho no byte 520
	buf[520] = HeaderTerminator

	_, err := w.Write(buf)
	return err
}
