// Package eval implementa o analisador e avaliador de expressões da linguagem dBASE II,
// suportando tipos (Character, Numeric, Logical), operadores aritméticos, relacionais,
// lógicos (.AND., .OR., .NOT.), concatenação, operador de substring ($) e funções nativas.
package eval

import (
	"fmt"
	"strconv"
	"strings"
)

// DataType define o tipo primitivo de dado avaliado.
type DataType int

const (
	TypeNull DataType = iota
	TypeString
	TypeNumber
	TypeLogical
)

// Value representa um valor polimórfico manipulado pelo dBASE II.
type Value struct {
	Type DataType
	Str  string
	Num  float64
	Bool bool
}

// NewStringValue cria um Value do tipo Character (string).
func NewStringValue(s string) Value {
	return Value{Type: TypeString, Str: s}
}

// NewNumberValue cria um Value do tipo Numeric (float64).
func NewNumberValue(n float64) Value {
	return Value{Type: TypeNumber, Num: n}
}

// NewLogicalValue cria um Value do tipo Logical (booleano).
func NewLogicalValue(b bool) Value {
	return Value{Type: TypeLogical, Bool: b}
}

// NewNullValue cria um Value nulo.
func NewNullValue() Value {
	return Value{Type: TypeNull}
}

// AsString converte o valor para representação em string.
func (v Value) AsString() string {
	switch v.Type {
	case TypeString:
		return v.Str
	case TypeNumber:
		if v.Num == float64(int64(v.Num)) {
			return fmt.Sprintf("%d", int64(v.Num))
		}
		return strconv.FormatFloat(v.Num, 'f', -1, 64)
	case TypeLogical:
		if v.Bool {
			return ".T."
		}
		return ".F."
	default:
		return ""
	}
}

// AsNumber converte o valor para float64 se possível.
func (v Value) AsNumber() (float64, error) {
	switch v.Type {
	case TypeNumber:
		return v.Num, nil
	case TypeString:
		n, err := strconv.ParseFloat(strings.TrimSpace(v.Str), 64)
		if err != nil {
			return 0, fmt.Errorf("não é possível converter '%s' para numérico", v.Str)
		}
		return n, nil
	case TypeLogical:
		if v.Bool {
			return 1, nil
		}
		return 0, nil
	default:
		return 0, fmt.Errorf("valor nulo")
	}
}

// AsBool converte o valor para booleano.
func (v Value) AsBool() bool {
	switch v.Type {
	case TypeLogical:
		return v.Bool
	case TypeNumber:
		return v.Num != 0
	case TypeString:
		upper := strings.ToUpper(strings.TrimSpace(v.Str))
		return upper == "T" || upper == "Y" || upper == ".T." || upper == ".Y." || upper == "TRUE"
	default:
		return false
	}
}

func (v Value) String() string {
	return v.AsString()
}
