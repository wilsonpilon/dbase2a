package eval

import (
	"fmt"
	"strconv"
	"strings"
)

// Parser converte uma sequência de tokens em uma AST de nós Node.
type Parser struct {
	tokens []Token
	pos    int
}

// ParseExpression analisa uma string de expressão e gera a AST de nós executáveis.
func ParseExpression(exprStr string) (Node, error) {
	exprStr = strings.TrimSpace(exprStr)
	if exprStr == "" {
		return nil, nil
	}

	lexer := NewLexer(exprStr)
	var tokens []Token
	for {
		tok := lexer.NextToken()
		if tok.Type == TokEOF {
			break
		}
		tokens = append(tokens, tok)
	}

	if len(tokens) == 0 {
		return nil, nil
	}

	p := &Parser{tokens: tokens, pos: 0}
	node, err := p.parseOr()
	if err != nil {
		return nil, err
	}

	if p.pos < len(p.tokens) {
		return nil, fmt.Errorf("token inesperado após o fim da expressão: %s", p.tokens[p.pos].Value)
	}

	return node, nil
}

func (p *Parser) current() Token {
	if p.pos >= len(p.tokens) {
		return Token{Type: TokEOF, Value: ""}
	}
	return p.tokens[p.pos]
}

func (p *Parser) advance() Token {
	tok := p.current()
	p.pos++
	return tok
}

func (p *Parser) match(types ...TokenType) bool {
	for _, t := range types {
		if p.current().Type == t {
			p.advance()
			return true
		}
	}
	return false
}

// Precedência 1: .OR.
func (p *Parser) parseOr() (Node, error) {
	left, err := p.parseAnd()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokOr {
		opTok := p.advance()
		right, err := p.parseAnd()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Op: opTok.Value, Left: left, Right: right}
	}

	return left, nil
}

// Precedência 2: .AND.
func (p *Parser) parseAnd() (Node, error) {
	left, err := p.parseNot()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokAnd {
		opTok := p.advance()
		right, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Op: opTok.Value, Left: left, Right: right}
	}

	return left, nil
}

// Precedência 3: .NOT., !
func (p *Parser) parseNot() (Node, error) {
	if p.current().Type == TokNot {
		opTok := p.advance()
		expr, err := p.parseNot()
		if err != nil {
			return nil, err
		}
		return &UnaryOpNode{Op: opTok.Value, Expr: expr}, nil
	}
	return p.parseComparison()
}

// Precedência 4: =, <>, !=, #, <, >, <=, >=, $
func (p *Parser) parseComparison() (Node, error) {
	left, err := p.parseAdditive()
	if err != nil {
		return nil, err
	}

	for {
		curr := p.current()
		if curr.Type == TokEqual || curr.Type == TokNotEqual ||
			curr.Type == TokLess || curr.Type == TokGreater ||
			curr.Type == TokLessEqual || curr.Type == TokGreaterEqual ||
			curr.Type == TokSubstring {
			opTok := p.advance()
			right, err := p.parseAdditive()
			if err != nil {
				return nil, err
			}
			left = &BinaryOpNode{Op: opTok.Value, Left: left, Right: right}
		} else {
			break
		}
	}

	return left, nil
}

// Precedência 5: +, -
func (p *Parser) parseAdditive() (Node, error) {
	left, err := p.parseMultiplicative()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokPlus || p.current().Type == TokMinus {
		opTok := p.advance()
		right, err := p.parseMultiplicative()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Op: opTok.Value, Left: left, Right: right}
	}

	return left, nil
}

// Precedência 6: *, /
func (p *Parser) parseMultiplicative() (Node, error) {
	left, err := p.parseUnary()
	if err != nil {
		return nil, err
	}

	for p.current().Type == TokStar || p.current().Type == TokSlash {
		opTok := p.advance()
		right, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		left = &BinaryOpNode{Op: opTok.Value, Left: left, Right: right}
	}

	return left, nil
}

// Precedência 7: Unário +, -
func (p *Parser) parseUnary() (Node, error) {
	if p.current().Type == TokMinus || p.current().Type == TokPlus {
		opTok := p.advance()
		expr, err := p.parseUnary()
		if err != nil {
			return nil, err
		}
		return &UnaryOpNode{Op: opTok.Value, Expr: expr}, nil
	}
	return p.parsePrimary()
}

// Precedência 8: Primários (Literais, Identificadores, Chamadas de Função, Parênteses)
func (p *Parser) parsePrimary() (Node, error) {
	tok := p.current()

	switch tok.Type {
	case TokString:
		p.advance()
		return &LiteralNode{Val: NewStringValue(tok.Value)}, nil

	case TokNumber:
		p.advance()
		num, err := strconv.ParseFloat(tok.Value, 64)
		if err != nil {
			return nil, fmt.Errorf("número inválido '%s': %w", tok.Value, err)
		}
		return &LiteralNode{Val: NewNumberValue(num)}, nil

	case TokLogical:
		p.advance()
		isTrue := strings.HasPrefix(strings.ToUpper(tok.Value), ".T") || strings.HasPrefix(strings.ToUpper(tok.Value), ".Y")
		return &LiteralNode{Val: NewLogicalValue(isTrue)}, nil

	case TokLParen:
		p.advance() // consome '('
		expr, err := p.parseOr()
		if err != nil {
			return nil, err
		}
		if p.current().Type != TokRParen {
			return nil, fmt.Errorf("esperado ')' para fechar expressão")
		}
		p.advance() // consome ')'
		return expr, nil

	case TokIdentifier:
		identTok := p.advance()
		// Verifica se é uma chamada de função (seguido por '(')
		if p.current().Type == TokLParen {
			p.advance() // consome '('
			var args []Node
			if p.current().Type != TokRParen {
				for {
					arg, err := p.parseOr()
					if err != nil {
						return nil, err
					}
					args = append(args, arg)

					if p.current().Type == TokComma {
						p.advance() // consome vírgula
						continue
					}
					break
				}
			}
			if p.current().Type != TokRParen {
				return nil, fmt.Errorf("esperado ')' ao final dos argumentos da função '%s'", identTok.Value)
			}
			p.advance() // consome ')'
			return &FunctionCallNode{
				Name: strings.ToUpper(identTok.Value),
				Args: args,
			}, nil
		}

		return &IdentifierNode{Name: strings.ToUpper(identTok.Value)}, nil

	default:
		return nil, fmt.Errorf("expressão ou token inválido: '%s'", tok.Value)
	}
}
