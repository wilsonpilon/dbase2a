package eval

import (
	"math"
	"testing"
)

func TestEvalArithmetic(t *testing.T) {
	tests := []struct {
		expr     string
		expected float64
	}{
		{"10 + 20", 30},
		{"100 - 45", 55},
		{"5 * 6", 30},
		{"50 / 2", 25},
		{"2 + 3 * 4", 14},
		{"(2 + 3) * 4", 20},
		{"-10 + 25", 15},
		{"+5 * +4", 20},
		{"100 / (2 + 3)", 20},
	}

	for _, tc := range tests {
		val, err := Eval(tc.expr, nil)
		if err != nil {
			t.Fatalf("Erro ao avaliar '%s': %v", tc.expr, err)
		}
		num, err := val.AsNumber()
		if err != nil {
			t.Fatalf("Erro ao converter resultado para número em '%s': %v", tc.expr, err)
		}
		if num != tc.expected {
			t.Errorf("Para '%s': esperava %f, obteve %f", tc.expr, tc.expected, num)
		}
	}
}

func TestEvalRelationalAndLogical(t *testing.T) {
	tests := []struct {
		expr     string
		expected bool
	}{
		{"10 > 5", true},
		{"10 < 5", false},
		{"10 >= 10", true},
		{"10 <= 9", false},
		{"10 = 10", true},
		{"10 <> 5", true},
		{"10 != 10", false},
		{"10 # 5", true},
		{".T. .AND. .T.", true},
		{".T. .AND. .F.", false},
		{".F. .OR. .T.", true},
		{".NOT. .F.", true},
		{"! .T.", false},
		{"10 > 5 .AND. 20 < 30", true},
		{"10 > 5 .AND. 20 > 30", false},
		{"10 < 5 .OR. 20 > 30 .OR. 1 = 1", true},
		{".NOT. (10 < 5)", true},
	}

	for _, tc := range tests {
		b, err := EvalBool(tc.expr, nil)
		if err != nil {
			t.Fatalf("Erro ao avaliar '%s': %v", tc.expr, err)
		}
		if b != tc.expected {
			t.Errorf("Para '%s': esperava %v, obteve %v", tc.expr, tc.expected, b)
		}
	}
}

func TestEvalStringsAndSubstring(t *testing.T) {
	ctx := &Context{
		GetFieldFunc: func(name string) (Value, bool) {
			if name == "NOME" {
				return NewStringValue("CARLOS DA SILVA"), true
			}
			if name == "CIDADE" {
				return NewStringValue("SAO PAULO"), true
			}
			return NewNullValue(), false
		},
	}

	tests := []struct {
		expr     string
		expected string
		isBool   bool
		boolExp  bool
	}{
		{expr: `"DBASE" + " " + "II"`, expected: "DBASE II"},
		{expr: `"SILVA" $ NOME`, isBool: true, boolExp: true},
		{expr: `"RIO" $ CIDADE`, isBool: true, boolExp: false},
		{expr: `UPPER("teste")`, expected: "TESTE"},
		{expr: `LOWER("TESTE")`, expected: "teste"},
		{expr: `TRIM("TEXTO   ")`, expected: "TEXTO"},
		{expr: `SUBSTR(NOME, 1, 6)`, expected: "CARLOS"},
		{expr: `LEN(CIDADE)`, expected: "9"},
		{expr: `IIF(10 > 5, "SIM", "NAO")`, expected: "SIM"},
	}

	for _, tc := range tests {
		val, err := Eval(tc.expr, ctx)
		if err != nil {
			t.Fatalf("Erro ao avaliar '%s': %v", tc.expr, err)
		}
		if tc.isBool {
			if val.AsBool() != tc.boolExp {
				t.Errorf("Para '%s': esperava bool %v, obteve %v", tc.expr, tc.boolExp, val.AsBool())
			}
		} else {
			if val.AsString() != tc.expected {
				t.Errorf("Para '%s': esperava '%s', obteve '%s'", tc.expr, tc.expected, val.AsString())
			}
		}
	}
}

func TestEvalContextVariablesAndFunctions(t *testing.T) {
	ctx := &Context{
		GetFieldFunc: func(name string) (Value, bool) {
			switch name {
			case "SALARIO":
				return NewNumberValue(3500.00), true
			case "ATIVO":
				return NewLogicalValue(true), true
			}
			return NewNullValue(), false
		},
		GetVarFunc: func(name string) (Value, bool) {
			if name == "TAXA" {
				return NewNumberValue(0.10), true
			}
			return NewNullValue(), false
		},
		DeletedFunc: func() bool {
			return false
		},
		RecNoFunc: func() int {
			return 42
		},
	}

	res, err := Eval("SALARIO * (1 + TAXA)", ctx)
	if err != nil {
		t.Fatalf("Erro no calculo: %v", err)
	}
	num, _ := res.AsNumber()
	if math.Abs(num-3850.00) > 1e-4 {
		t.Errorf("Esperava 3850.00, obteve %f", num)
	}

	bRes, err := EvalBool("ATIVO .AND. .NOT. DELETED() .AND. RECNO() = 42", ctx)
	if err != nil {
		t.Fatalf("Erro no calculo bool: %v", err)
	}
	if !bRes {
		t.Errorf("Esperava true para checagem com DELETED() e RECNO()")
	}
}
