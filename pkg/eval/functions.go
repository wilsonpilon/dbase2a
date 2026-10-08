package eval

import (
	"fmt"
	"math"
	"strconv"
	"strings"
	"time"
)

func (n *FunctionCallNode) Eval(ctx *Context) (Value, error) {
	fnName := strings.ToUpper(n.Name)

	// Funções customizadas definidas no contexto
	if ctx != nil && ctx.CustomFuncs != nil {
		if fn, exists := ctx.CustomFuncs[fnName]; exists {
			var argVals []Value
			for _, arg := range n.Args {
				val, err := arg.Eval(ctx)
				if err != nil {
					return NewNullValue(), err
				}
				argVals = append(argVals, val)
			}
			return fn(argVals)
		}
	}

	// Funções intrínsecas do dBASE II
	switch fnName {
	case "DELETED":
		if ctx != nil && ctx.DeletedFunc != nil {
			return NewLogicalValue(ctx.DeletedFunc()), nil
		}
		return NewLogicalValue(false), nil

	case "EOF":
		if ctx != nil && ctx.EOFFunc != nil {
			return NewLogicalValue(ctx.EOFFunc()), nil
		}
		return NewLogicalValue(false), nil

	case "RECNO":
		if ctx != nil && ctx.RecNoFunc != nil {
			return NewNumberValue(float64(ctx.RecNoFunc())), nil
		}
		return NewNumberValue(0), nil

	case "TRIM", "RTRIM":
		if len(n.Args) < 1 {
			return NewNullValue(), fmt.Errorf("função %s requer 1 argumento", fnName)
		}
		val, err := n.Args[0].Eval(ctx)
		if err != nil {
			return NewNullValue(), err
		}
		return NewStringValue(strings.TrimRight(val.AsString(), " \t\r\n")), nil

	case "LTRIM":
		if len(n.Args) < 1 {
			return NewNullValue(), fmt.Errorf("função LTRIM requer 1 argumento")
		}
		val, err := n.Args[0].Eval(ctx)
		if err != nil {
			return NewNullValue(), err
		}
		return NewStringValue(strings.TrimLeft(val.AsString(), " \t\r\n")), nil

	case "ALLTRIM":
		if len(n.Args) < 1 {
			return NewNullValue(), fmt.Errorf("função ALLTRIM requer 1 argumento")
		}
		val, err := n.Args[0].Eval(ctx)
		if err != nil {
			return NewNullValue(), err
		}
		return NewStringValue(strings.TrimSpace(val.AsString())), nil

	case "UPPER":
		if len(n.Args) < 1 {
			return NewNullValue(), fmt.Errorf("função UPPER requer 1 argumento")
		}
		val, err := n.Args[0].Eval(ctx)
		if err != nil {
			return NewNullValue(), err
		}
		return NewStringValue(strings.ToUpper(val.AsString())), nil

	case "LOWER":
		if len(n.Args) < 1 {
			return NewNullValue(), fmt.Errorf("função LOWER requer 1 argumento")
		}
		val, err := n.Args[0].Eval(ctx)
		if err != nil {
			return NewNullValue(), err
		}
		return NewStringValue(strings.ToLower(val.AsString())), nil

	case "LEN":
		if len(n.Args) < 1 {
			return NewNullValue(), fmt.Errorf("função LEN requer 1 argumento")
		}
		val, err := n.Args[0].Eval(ctx)
		if err != nil {
			return NewNullValue(), err
		}
		return NewNumberValue(float64(len(val.AsString()))), nil

	case "SUBSTR":
		if len(n.Args) < 2 {
			return NewNullValue(), fmt.Errorf("função SUBSTR requer pelo menos 2 argumentos")
		}
		strVal, err := n.Args[0].Eval(ctx)
		if err != nil {
			return NewNullValue(), err
		}
		posVal, err := n.Args[1].Eval(ctx)
		if err != nil {
			return NewNullValue(), err
		}
		startPos := int(posVal.Num) - 1 // dBASE usa indexação 1-based
		if startPos < 0 {
			startPos = 0
		}
		runes := []rune(strVal.AsString())
		if startPos >= len(runes) {
			return NewStringValue(""), nil
		}

		length := len(runes) - startPos
		if len(n.Args) >= 3 {
			lenVal, err := n.Args[2].Eval(ctx)
			if err != nil {
				return NewNullValue(), err
			}
			reqLen := int(lenVal.Num)
			if reqLen < length {
				length = reqLen
			}
		}
		if length <= 0 {
			return NewStringValue(""), nil
		}

		return NewStringValue(string(runes[startPos : startPos+length])), nil

	case "VAL":
		if len(n.Args) < 1 {
			return NewNullValue(), fmt.Errorf("função VAL requer 1 argumento")
		}
		val, err := n.Args[0].Eval(ctx)
		if err != nil {
			return NewNullValue(), err
		}
		num, _ := strconv.ParseFloat(strings.TrimSpace(val.AsString()), 64)
		return NewNumberValue(num), nil

	case "STR":
		if len(n.Args) < 1 {
			return NewNullValue(), fmt.Errorf("função STR requer pelo menos 1 argumento")
		}
		val, err := n.Args[0].Eval(ctx)
		if err != nil {
			return NewNullValue(), err
		}
		num, _ := val.AsNumber()
		dec := 0
		if len(n.Args) >= 3 {
			decVal, _ := n.Args[2].Eval(ctx)
			dec = int(decVal.Num)
		}
		var str string
		if dec > 0 {
			str = strconv.FormatFloat(num, 'f', dec, 64)
		} else {
			str = fmt.Sprintf("%d", int64(num))
		}
		if len(n.Args) >= 2 {
			lenVal, _ := n.Args[1].Eval(ctx)
			totalLen := int(lenVal.Num)
			if len(str) < totalLen {
				str = strings.Repeat(" ", totalLen-len(str)) + str
			}
		}
		return NewStringValue(str), nil

	case "INT":
		if len(n.Args) < 1 {
			return NewNullValue(), fmt.Errorf("função INT requer 1 argumento")
		}
		val, err := n.Args[0].Eval(ctx)
		if err != nil {
			return NewNullValue(), err
		}
		num, _ := val.AsNumber()
		return NewNumberValue(math.Floor(num)), nil

	case "ABS":
		if len(n.Args) < 1 {
			return NewNullValue(), fmt.Errorf("função ABS requer 1 argumento")
		}
		val, err := n.Args[0].Eval(ctx)
		if err != nil {
			return NewNullValue(), err
		}
		num, _ := val.AsNumber()
		return NewNumberValue(math.Abs(num)), nil

	case "DATE":
		return NewStringValue(time.Now().Format("01/02/06")), nil

	case "TIME":
		return NewStringValue(time.Now().Format("15:04:05")), nil

	case "IIF", "IF":
		if len(n.Args) < 3 {
			return NewNullValue(), fmt.Errorf("função IIF requer 3 argumentos (cond, verdadeiro, falso)")
		}
		condVal, err := n.Args[0].Eval(ctx)
		if err != nil {
			return NewNullValue(), err
		}
		if condVal.AsBool() {
			return n.Args[1].Eval(ctx)
		}
		return n.Args[2].Eval(ctx)

	default:
		return NewNullValue(), fmt.Errorf("função desconhecida '%s'", fnName)
	}
}
