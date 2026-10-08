// Package dbf implementa a especificação completa de leitura, gravação,
// manipulação e compatibilidade binária de arquivos de dados dBASE II (.DBF).
package dbf

import (
	"fmt"
	"strings"
)

// FieldType representa o tipo de dado de uma coluna no padrão dBASE II.
type FieldType byte

const (
	// TypeCharacter representa dados alfanuméricos / texto (C).
	TypeCharacter FieldType = 'C'
	// TypeNumeric representa números inteiros ou com ponto flutuante (N).
	TypeNumeric FieldType = 'N'
	// TypeLogical representa valores booleanos T/t/Y/y/F/f/N/n (L).
	TypeLogical FieldType = 'L'
)

// Field define a estrutura de um campo de tabela no dBASE II (16 bytes no cabeçalho).
type Field struct {
	Name     string    `json:"name"`     // Nome do campo (até 10 caracteres)
	Type     FieldType `json:"type"`     // 'C', 'N', ou 'L'
	Length   uint8     `json:"length"`   // Tamanho do campo em bytes (1 a 254)
	Decimals uint8     `json:"decimals"` // Casas decimais (0 a 19 para tipo 'N')
	Offset   int       `json:"offset"`   // Deslocamento interno dentro do registro de dados
}

// Validate verifica se as dimensões e o nome do campo respeitam as regras do dBASE II.
func (f Field) Validate() error {
	name := strings.TrimSpace(f.Name)
	if name == "" {
		return fmt.Errorf("nome do campo não pode ser vazio")
	}
	if len(name) > 10 {
		return fmt.Errorf("nome do campo '%s' ultrapassa o limite de 10 caracteres do dBASE II", name)
	}

	switch f.Type {
	case TypeCharacter:
		if f.Length == 0 {
			return fmt.Errorf("campo caractere '%s' deve ter tamanho >= 1", name)
		}
	case TypeNumeric:
		if f.Length == 0 {
			return fmt.Errorf("campo numérico '%s' deve ter tamanho >= 1", name)
		}
		if int(f.Decimals) >= int(f.Length) {
			return fmt.Errorf("campo numérico '%s': casas decimais (%d) devem ser menores que o tamanho total (%d)", name, f.Decimals, f.Length)
		}
	case TypeLogical:
		if f.Length != 1 {
			return fmt.Errorf("campo lógico '%s' deve ter tamanho exatamente 1", name)
		}
	default:
		return fmt.Errorf("tipo de campo inválido '%c' para '%s' (suportados: C, N, L)", f.Type, name)
	}
	return nil
}
