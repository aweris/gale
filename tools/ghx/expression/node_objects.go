package expression

import (
	"reflect"

	"github.com/rhysd/actionlint"
)

var (
	_ Interpreter = new(VariableNode)
	_ Interpreter = new(ObjectDerefNode)
	_ Interpreter = new(ArrayDerefNode)
	_ Interpreter = new(IndexAccessNode)
)

// VariableNode is a wrapper of actionlint.VariableNode
type VariableNode actionlint.VariableNode

func (n VariableNode) Evaluate(p VariableProvider) (interface{}, error) {
	return p.GetVariable(n.Name)
}

// ObjectDerefNode is a wrapper of actionlint.ObjectDerefNode
type ObjectDerefNode actionlint.ObjectDerefNode

func (n ObjectDerefNode) Evaluate(provider VariableProvider) (interface{}, error) {
	// convert wrapper type and call Evaluate() to get the receiver value
	left, err := getInterpreterFromNode(n.Receiver).Evaluate(provider)
	if err != nil {
		return nil, err
	}

	// get property value from the receiver value
	return getPropertyValue(reflect.ValueOf(left), n.Property)
}

// ArrayDerefNode is a wrapper of actionlint.ArrayDerefNode
type ArrayDerefNode actionlint.ArrayDerefNode

func (n ArrayDerefNode) Evaluate(provider VariableProvider) (interface{}, error) {
	// convert wrapper type and call Evaluate() to get the receiver value
	left, err := getInterpreterFromNode(n.Receiver).Evaluate(provider)
	if err != nil {
		return nil, err
	}

	return getSafeValue(reflect.ValueOf(left)), nil
}

// IndexAccessNode is a wrapper of actionlint.IndexAccessNode
type IndexAccessNode actionlint.IndexAccessNode

func (n IndexAccessNode) Evaluate(provider VariableProvider) (interface{}, error) {
	// convert operand to wrapper type and call Evaluate() to get the receiver value
	left, err := getInterpreterFromNode(n.Operand).Evaluate(provider)
	if err != nil {
		return nil, err
	}

	// convert index to wrapper type and call Evaluate() to get the index value
	right, err := getInterpreterFromNode(n.Index).Evaluate(provider)
	if err != nil {
		return nil, err
	}

	// get values from the receiver and the index
	leftValue := reflect.ValueOf(left)
	rightValue := reflect.ValueOf(right)

	// evaluate the index value
	switch rightValue.Kind() {
	case reflect.String:
		return getPropertyValue(leftValue, rightValue.String())
	case reflect.Int:
		if leftValue.Kind() != reflect.Slice {
			return nil, nil
		}

		index := int(rightValue.Int())
		length := leftValue.Len()

		if index >= 0 && index < length {
			return leftValue.Index(index).Interface(), nil
		}
	}

	// fallback to nil
	return nil, nil
}
