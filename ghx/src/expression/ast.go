package expression

import (
	"fmt"
	"strings"

	"github.com/aweris/gale/common/log"
)

// The original source code is from https://github.com/rhysd/actionlint/blob/5656337c1ab1c7022a74181428f6ebb4504d2d25/ast.go
// We modified it to make it compilable with our codebase.

// String represents generic string value with Github Actions expression support.
type String struct {
	Value  string // Value is a raw value of the string.
	Quoted bool   // Quoted represents the string is quoted with ' or " in the YAML source.
}

// NewString creates a new String instance.
func NewString(value string) *String {
	return &String{
		Value:  value,
		Quoted: false,
	}
}

// Eval evaluates the expression and returns the string value.
func (s *String) Eval(provider VariableProvider) string {
	if s.Quoted {
		return s.Value
	}

	exprs, err := ParseExpressions(s.Value)
	if err != nil {
		log.Errorf("failed to parse expressions", "expression", s.Value, "error", err)
		return ""
	}

	if len(exprs) == 0 {
		return s.Value
	}

	str := s.Value

	for _, expr := range exprs {
		val, err := expr.Evaluate(provider)
		if err != nil {
			log.Errorf("failed to evaluate expression", "expression", expr, "error", err)
			return ""
		}

		if val == nil {
			str = strings.Replace(str, expr.Value, "", 1)

			continue
		}

		if v, ok := val.(string); ok {
			str = strings.Replace(str, expr.Value, v, 1)
		}
	}

	return str
}

// value represents generic value with Github Actions expression support.
type value[T bool | int | float64] struct {
	Value      T       // Value is a raw value of the T.
	Expression *String // Expression is a string when expression syntax ${{ }} is used for this section.
}

func (v *value[T]) Eval(provider VariableProvider) (T, error) {
	if v.Expression == nil {
		return v.Value, nil
	}

	expr, err := NewExpression(v.Expression.Value)
	if err != nil {
		return *new(T), err
	}

	val, err := expr.Evaluate(provider)
	if err != nil {
		return *new(T), err
	}

	if v, ok := val.(T); ok {
		return v, nil
	}

	return *new(T), fmt.Errorf("cannot convert %v to %T", val, *new(T))
}

// Bool represents generic boolean value with Github Actions expression support.
//
// The value could be a raw boolean value or an expression. If Expression field is not nil, the value is considered
// as an expression, and it should be evaluated to get the boolean value. Otherwise, the value is a raw boolean value.
type Bool struct {
	value[bool]
}

// NewBool creates a new Bool instance with the given boolean value.
func NewBool(v bool) *Bool {
	return &Bool{
		value: value[bool]{
			Value: v,
		},
	}
}

// NewBoolExpr creates a new Bool instance with the given expression.
func NewBoolExpr(expr string) *Bool {
	return &Bool{
		value: value[bool]{
			Expression: NewString(expr),
		},
	}
}

// Int represents generic integer value with Github Actions expression support.
//
// The value could be a raw integer value or an expression. If Expression field is not nil, the value is considered
// as an expression, and it should be evaluated to get the integer value. Otherwise, the value is a raw integer value.
type Int struct {
	value[int]
}

// NewInt creates a new Int instance with the given integer value.
func NewInt(v int) *Int {
	return &Int{
		value: value[int]{
			Value: v,
		},
	}
}

// NewIntExpr creates a new Int instance with the given expression.
func NewIntExpr(expr string) *Int {
	return &Int{
		value: value[int]{
			Expression: NewString(expr),
		},
	}
}

// Float represents generic float value with Github Actions expression support.
//
// The value could be a raw float value or an expression. If Expression field is not nil, the value is considered
// as an expression, and it should be evaluated to get the float value. Otherwise, the value is a raw float value.
type Float struct {
	value[float64]
}

// NewFloat creates a new Float instance with the given float value.
func NewFloat(v float64) *Float {
	return &Float{
		value: value[float64]{
			Value: v,
		},
	}
}

// NewFloatExpr creates a new Float instance with the given expression.
func NewFloatExpr(expr string) *Float {
	return &Float{
		value: value[float64]{
			Expression: NewString(expr),
		},
	}
}
