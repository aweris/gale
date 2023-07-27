package expression

import (
	"fmt"
	"reflect"

	"github.com/rhysd/actionlint"
)

var (
	_ Interpreter = new(NotOpNode)
	_ Interpreter = new(CompareOpNode)
	_ Interpreter = new(LogicalOpNode)
)

// NotOpNode is a wrapper of actionlint.NotOpNode
type NotOpNode actionlint.NotOpNode

func (n NotOpNode) Evaluate(provider VariableProvider) (interface{}, error) {
	// convert operand to wrapper type and call Evaluate() to get the receiver value
	operand, err := getInterpreterFromNode(n.Operand).Evaluate(provider)
	if err != nil {
		return nil, err
	}

	return !isTruthy(operand), nil
}

// CompareOpNode is a wrapper of actionlint.CompareOpNode
type CompareOpNode actionlint.CompareOpNode

func (n CompareOpNode) Evaluate(provider VariableProvider) (interface{}, error) {
	// convert left operand to wrapper type and call Evaluate() to get the receiver value
	left, err := getInterpreterFromNode(n.Left).Evaluate(provider)
	if err != nil {
		return nil, err
	}

	// convert right operand to wrapper type and call Evaluate() to get the receiver value
	right, err := getInterpreterFromNode(n.Right).Evaluate(provider)
	if err != nil {
		return nil, err
	}

	leftValue := reflect.ValueOf(left)
	rightValue := reflect.ValueOf(right)

	return compareValues(leftValue, rightValue, n.Kind)
}

// LogicalOpNode is a wrapper of actionlint.LogicalOpNode
type LogicalOpNode actionlint.LogicalOpNode

func (n LogicalOpNode) Evaluate(provider VariableProvider) (interface{}, error) {
	// convert left operand to wrapper type and call Evaluate() to get the receiver value
	left, err := getInterpreterFromNode(n.Left).Evaluate(provider)
	if err != nil {
		return nil, err
	}

	leftValue := reflect.ValueOf(left)

	// convert right operand to wrapper type and call Evaluate() to get the receiver value
	right, err := getInterpreterFromNode(n.Right).Evaluate(provider)
	if err != nil {
		return nil, err
	}

	rightValue := reflect.ValueOf(right)

	// look at the left operand first to determine the result of the logical operator. No need to evaluate the right
	// operand if the result is already determined.
	switch n.Kind {
	case actionlint.LogicalOpNodeKindAnd:
		if isTruthy(left) {
			return getSafeValue(rightValue), nil
		}

		return getSafeValue(leftValue), nil

	case actionlint.LogicalOpNodeKindOr:
		if isTruthy(left) {
			return getSafeValue(leftValue), nil
		}

		return getSafeValue(rightValue), nil
	}

	return nil, fmt.Errorf("invalid logical operator node")
}
