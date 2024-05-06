package expression

import (
	"fmt"
	"regexp"
	"strings"

	"github.com/rhysd/actionlint"
)

//goland:noinspection RegExpRedundantEscape // just to make it more readable
var exprRe = regexp.MustCompile(`\$\{\{[^{}]+\}\}`)

// VariableProvider is an interface to provide variable values for expression evaluation.
type VariableProvider interface {
	// GetVariable returns a value of variable with given name. If the variable is not defined, the second return value
	// should be false.
	GetVariable(name string) (interface{}, error)
}

// Interpreter is an interface to evaluate expression.
type Interpreter interface {
	// Evaluate evaluates the expression and returns the result.
	Evaluate(provider VariableProvider) (interface{}, error)
}

// Expression represents a GitHub expression in a string with position.
type Expression struct {
	Value       string      // Value is a raw value of the string.
	StartIndex  int         // StartIndex is a start index of the expression in the source string.
	EndIndex    int         // EndIndex is an end index of the expression in the source string.
	interpreter Interpreter // interpreter is an interpreter for the expression.
}

// NewExpression parses a string and returns an Expression.
//
// The method assumes that the string contains an expression. If the string omits the expression syntax (${{ }}). It
// will be added to the string automatically and parsed.
func NewExpression(value string) (*Expression, error) {
	matches := exprRe.FindStringSubmatch(value)

	// If there are no matches, then the string omits the expression syntax (${{ }}). This is valid for if conditionals.
	// to normalize the string, and make it consistent, we add the expression syntax to the string.
	//
	// source: https://docs.github.com/en/actions/learn-github-actions/expressions#about-expressions
	if len(matches) == 0 {
		value = strings.TrimSpace(value)
		value = fmt.Sprintf("${{%s}}", value)
	}

	lexer := actionlint.NewExprLexer(strings.TrimPrefix(value, "${{"))

	parser := actionlint.NewExprParser()

	node, err := parser.Parse(lexer)
	if err != nil {
		return nil, fmt.Errorf("failed to parse expression: %w", err)
	}

	return &Expression{
		Value:       value,
		StartIndex:  0,
		EndIndex:    len(value) - 1,
		interpreter: getInterpreterFromNode(node),
	}, nil
}

// ParseExpressions parses a string and returns a slice of Expressions with their start and end indexes in input string.
//
// The method strictly checks expressions syntax. If the string omits the expression syntax (${{ }}), it will be
// considered as an regular string.
//
// If the string not contains any expressions, the method will return an empty slice. If the string contains invalid
// expressions, the method will return an error.
func ParseExpressions(input string) ([]*Expression, error) {
	expressions := make([]*Expression, 0)

	parser := actionlint.NewExprParser()

	matches := exprRe.FindAllStringIndex(input, -1)

	for _, match := range matches {
		value := input[match[0]:match[1]]

		lexer := actionlint.NewExprLexer(strings.TrimPrefix(value, "${{"))

		node, expErr := parser.Parse(lexer)
		if expErr != nil {
			return nil, fmt.Errorf("failed to parse expression: %w", expErr)
		}

		expression := Expression{
			Value:       value,
			StartIndex:  match[0],
			EndIndex:    match[1] - 1,
			interpreter: getInterpreterFromNode(node),
		}

		expressions = append(expressions, &expression)
	}

	return expressions, nil
}

// Evaluate evaluates the expression and returns the result.
func (e *Expression) Evaluate(provider VariableProvider) (interface{}, error) {
	return e.interpreter.Evaluate(provider)
}
