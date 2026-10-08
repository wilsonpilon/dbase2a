package eval

// Context provê o ambiente de avaliação de variáveis, campos de registros e funções de sistema.
type Context struct {
	GetFieldFunc func(name string) (Value, bool)
	GetVarFunc   func(name string) (Value, bool)
	DeletedFunc  func() bool
	EOFFunc      func() bool
	RecNoFunc    func() int
	CustomFuncs  map[string]func(args []Value) (Value, error)
}

// Node é a interface base para qualquer nó da árvore sintática de expressões.
type Node interface {
	Eval(ctx *Context) (Value, error)
}

// LiteralNode representa um valor constante (string, número ou booleano).
type LiteralNode struct {
	Val Value
}

func (n *LiteralNode) Eval(ctx *Context) (Value, error) {
	return n.Val, nil
}

// IdentifierNode representa uma variável em memória ou campo da tabela ativa.
type IdentifierNode struct {
	Name string
}

func (n *IdentifierNode) Eval(ctx *Context) (Value, error) {
	if ctx != nil {
		if ctx.GetFieldFunc != nil {
			if val, ok := ctx.GetFieldFunc(n.Name); ok {
				return val, nil
			}
		}
		if ctx.GetVarFunc != nil {
			if val, ok := ctx.GetVarFunc(n.Name); ok {
				return val, nil
			}
		}
	}
	// Se não encontrado em campos ou variáveis, retorna string literal como fallback permissivo do dBASE II
	return NewStringValue(n.Name), nil
}

// UnaryOpNode representa uma operação unária (-num, +num, .NOT. bool, !bool).
type UnaryOpNode struct {
	Op   string
	Expr Node
}

// BinaryOpNode representa uma operação binária (+, -, *, /, =, <>, <, >, <=, >=, $, .AND., .OR.).
type BinaryOpNode struct {
	Op    string
	Left  Node
	Right Node
}

// FunctionCallNode representa uma chamada de função (ex: TRIM(NOME), UPPER(STR), DELETED(), RECNO()).
type FunctionCallNode struct {
	Name string
	Args []Node
}
