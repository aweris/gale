package expression

import "github.com/rhysd/actionlint"

var (
	_ Interpreter = new(NullNode)
	_ Interpreter = new(BoolNode)
	_ Interpreter = new(IntNode)
	_ Interpreter = new(FloatNode)
	_ Interpreter = new(StringNode)
)

// NullNode is a wrapper of actionlint.NullNode
type NullNode actionlint.NullNode

func (n NullNode) Evaluate(_ VariableProvider) (interface{}, error) {
	return nil, nil
}

// BoolNode is a wrapper of actionlint.BoolNode
type BoolNode actionlint.BoolNode

func (n BoolNode) Evaluate(_ VariableProvider) (interface{}, error) {
	return n.Value, nil
}

// IntNode is a wrapper of actionlint.IntNode
type IntNode actionlint.IntNode

func (n IntNode) Evaluate(_ VariableProvider) (interface{}, error) {
	return n.Value, nil
}

// FloatNode is a wrapper of actionlint.FloatNode
type FloatNode actionlint.FloatNode

func (n FloatNode) Evaluate(_ VariableProvider) (interface{}, error) {
	return n.Value, nil
}

// StringNode is a wrapper of actionlint.StringNode
type StringNode actionlint.StringNode

func (n StringNode) Evaluate(_ VariableProvider) (interface{}, error) {
	return n.Value, nil
}
