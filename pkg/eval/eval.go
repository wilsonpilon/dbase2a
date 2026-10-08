package eval

import (
	"fmt"
	"math"
	"strings"
)

// Eval avalia uma string de expressão no contexto fornecido e retorna o Value resultante.
func Eval(exprStr string, ctx *Context) (Value, error) {
	node, err := ParseExpression(exprStr)
	if err != nil {
		return NewNullValue(), err
	}
	if node == nil {
		return NewNullValue(), nil
	}
	return node.Eval(ctx)
}

// EvalBool avalia a expressão e retorna diretamente um valor booleano.
func EvalBool(exprStr string, ctx *Context) (bool, error) {
	exprStr = strings.TrimSpace(exprStr)
	if exprStr == "" {
		return true, nil
	}
	val, err := Eval(exprStr, ctx)
	if err != nil {
		return false, err
	}
	return val.AsBool(), nil
}

func (n *UnaryOpNode) Eval(ctx *Context) (Value, error) {
	val, err := n.Expr.Eval(ctx)
	if err != nil {
		return NewNullValue(), err
	}

	op := strings.ToUpper(n.Op)
	switch op {
	case "-", "-UNARY":
		num, err := val.AsNumber()
		if err != nil {
			return NewNullValue(), fmt.Errorf("operador unário '-' requer numérico: %w", err)
		}
		return NewNumberValue(-num), nil
	case "+":
		num, err := val.AsNumber()
		if err != nil {
			return NewNullValue(), fmt.Errorf("operador unário '+' requer numérico: %w", err)
		}
		return NewNumberValue(num), nil
	case ".NOT.", "NOT", "!":
		return NewLogicalValue(!val.AsBool()), nil
	default:
		return NewNullValue(), fmt.Errorf("operador unário desconhecido '%s'", n.Op)
	}
}

func (n *BinaryOpNode) Eval(ctx *Context) (Value, error) {
	leftVal, err := n.Left.Eval(ctx)
	if err != nil {
		return NewNullValue(), err
	}

	op := strings.ToUpper(n.Op)

	// Curto-circuito lógico para .AND. e .OR.
	if op == ".AND." || op == "AND" {
		if !leftVal.AsBool() {
			return NewLogicalValue(false), nil
		}
		rightVal, err := n.Right.Eval(ctx)
		if err != nil {
			return NewNullValue(), err
		}
		return NewLogicalValue(rightVal.AsBool()), nil
	}

	if op == ".OR." || op == "OR" {
		if leftVal.AsBool() {
			return NewLogicalValue(true), nil
		}
		rightVal, err := n.Right.Eval(ctx)
		if err != nil {
			return NewNullValue(), err
		}
		return NewLogicalValue(rightVal.AsBool()), nil
	}

	rightVal, err := n.Right.Eval(ctx)
	if err != nil {
		return NewNullValue(), err
	}

	switch op {
	case "+":
		if leftVal.Type == TypeNumber && rightVal.Type == TypeNumber {
			return NewNumberValue(leftVal.Num + rightVal.Num), nil
		}
		return NewStringValue(leftVal.AsString() + rightVal.AsString()), nil

	case "-":
		if leftVal.Type == TypeNumber && rightVal.Type == TypeNumber {
			return NewNumberValue(leftVal.Num - rightVal.Num), nil
		}
		// Concatenação de string com remoção de espaços intermediários clássico do dBase
		lStr := strings.TrimRight(leftVal.AsString(), " ")
		trailing := len(leftVal.AsString()) - len(lStr)
		return NewStringValue(lStr + rightVal.AsString() + strings.Repeat(" ", trailing)), nil

	case "*":
		n1, err1 := leftVal.AsNumber()
		n2, err2 := rightVal.AsNumber()
		if err1 != nil || err2 != nil {
			return NewNullValue(), fmt.Errorf("multiplicação requer operandos numéricos")
		}
		return NewNumberValue(n1 * n2), nil

	case "/":
		n1, err1 := leftVal.AsNumber()
		n2, err2 := rightVal.AsNumber()
		if err1 != nil || err2 != nil {
			return NewNullValue(), fmt.Errorf("divisão requer operandos numéricos")
		}
		if n2 == 0 {
			return NewNumberValue(0), nil // dBase II não quebra na divisão por zero
		}
		return NewNumberValue(n1 / n2), nil

	case "$":
		// Substring: str1 $ str2 (str1 contido em str2)
		subStr := strings.ToUpper(leftVal.AsString())
		targetStr := strings.ToUpper(rightVal.AsString())
		return NewLogicalValue(strings.Contains(targetStr, subStr)), nil

	case "=":
		return NewLogicalValue(compareEqual(leftVal, rightVal)), nil

	case "<>", "!=", "#":
		return NewLogicalValue(!compareEqual(leftVal, rightVal)), nil

	case "<":
		return NewLogicalValue(compareLess(leftVal, rightVal)), nil

	case "<=":
		return NewLogicalValue(compareLess(leftVal, rightVal) || compareEqual(leftVal, rightVal)), nil

	case ">":
		return NewLogicalValue(!compareLess(leftVal, rightVal) && !compareEqual(leftVal, rightVal)), nil

	case ">=":
		return NewLogicalValue(!compareLess(leftVal, rightVal)), nil

	default:
		return NewNullValue(), fmt.Errorf("operador binário '%s' não suportado", n.Op)
	}
}

func compareEqual(a, b Value) bool {
	if a.Type == TypeNumber && b.Type == TypeNumber {
		return math.Abs(a.Num-b.Num) < 1e-9
	}
	if a.Type == TypeLogical && b.Type == TypeLogical {
		return a.Bool == b.Bool
	}
	// Comparação de string no padrão dBase (case-insensitive e prefixo)
	s1 := strings.TrimSpace(strings.ToUpper(a.AsString()))
	s2 := strings.TrimSpace(strings.ToUpper(b.AsString()))
	if s1 == s2 {
		return true
	}
	// Se a for numérico convertido ou b for, compara como float se ambos válidos
	if numA, errA := a.AsNumber(); errA == nil {
		if numB, errB := b.AsNumber(); errB == nil {
			return math.Abs(numA-numB) < 1e-9
		}
	}
	return strings.EqualFold(s1, s2)
}

func compareLess(a, b Value) bool {
	if a.Type == TypeNumber && b.Type == TypeNumber {
		return a.Num < b.Num
	}
	if numA, errA := a.AsNumber(); errA == nil {
		if numB, errB := b.AsNumber(); errB == nil {
			return numA < numB
		}
	}
	return strings.ToUpper(a.AsString()) < strings.ToUpper(b.AsString())
}
