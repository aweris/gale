package expression

import (
	"fmt"
	"math"
	"os"
	"path/filepath"
	"testing"
)

func TestNewExpression(t *testing.T) {
	tests := []struct {
		name     string
		value    string
		expected *Expression
	}{
		{
			name:     "expression omits syntax",
			value:    "foobar",
			expected: &Expression{Value: "${{foobar}}", StartIndex: 0, EndIndex: 10},
		},
		{
			name:     "expression with syntax",
			value:    "${{foobar}}",
			expected: &Expression{Value: "${{foobar}}", StartIndex: 0, EndIndex: 10},
		},
		{
			name:     "invalid expression",
			value:    "${{ invalid expression }",
			expected: nil, // Error expected
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := NewExpression(tt.value)

			if tt.expected == nil && err == nil {
				t.Errorf("Expected error, but got nil for input: %s", tt.value)
			}

			if tt.expected != nil && err != nil {
				t.Errorf("Expected no error, but got %s for input: %s", err.Error(), tt.value)
			}

			if tt.expected != nil {
				if tt.expected.Value != expr.Value {
					t.Errorf("Expected %s, but got %s for input: %s", tt.expected.Value, expr.Value, tt.value)
				}

				if tt.expected.StartIndex != expr.StartIndex || tt.expected.EndIndex != expr.EndIndex {
					t.Errorf(
						"Expected start and end indexes %d:%d, but got %d:%d for input: %s",
						tt.expected.StartIndex, tt.expected.EndIndex, expr.StartIndex, expr.EndIndex, tt.value,
					)
				}
			}
		})
	}
}

func TestParseExpressions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected []*Expression
	}{
		{
			name:     "no expressions",
			input:    "Hello, world!",
			expected: []*Expression{},
		},
		{
			name:  "single valid expression",
			input: "This is a ${{ valid }} expression",
			expected: []*Expression{
				{Value: "${{ valid }}", StartIndex: 10, EndIndex: 21},
			},
		},
		{
			name:     "invalid expression",
			input:    "This is an ${{ invalid expression",
			expected: nil, // Error expected
		},
		{
			name:  "multiple expressions",
			input: "${{ first }} and ${{ second }} expressions",
			expected: []*Expression{
				{Value: "${{ first }}", StartIndex: 0, EndIndex: 11},
				{Value: "${{ second }}", StartIndex: 17, EndIndex: 29},
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			exprs, err := ParseExpressions(tt.input)

			if err != nil {
				t.Errorf("Expected no error, but got %s for input: %s", err.Error(), tt.input)
			}

			if tt.expected != nil {
				if len(tt.expected) != len(exprs) {
					t.Errorf("Expected %d expressions, but got %d for input: %s", len(tt.expected), len(exprs), tt.input)
				}

				for i, expr := range exprs {
					if tt.expected[i].Value != expr.Value {
						t.Errorf("Expected %s, but got %s for input: %s", tt.expected[i].Value, expr.Value, expr.Value)
					}

					if tt.expected[i].StartIndex != expr.StartIndex || tt.expected[i].EndIndex != expr.EndIndex {
						t.Errorf(
							"Expected start and end indexes %d:%d, but got %d:%d for input: %s",
							tt.expected[i].StartIndex, tt.expected[i].EndIndex, expr.StartIndex, expr.EndIndex, expr.Value,
						)
					}
				}
			}
		})
	}
}

func TestExpression_EvaluateLiterals(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"integer literal", "123", 123},
		{"float literal", "123.456", 123.456},
		{"boolean literal true", "true", true},
		{"boolean literal false", "false", false},
		{"null literal", "null", nil},
		{"hexadecimal literal", "0xff", 255},
		{"exponential literal", "1e3", float64(1000)},
		{"exponential literal with negative sign", "1e-3", 0.001},
		{"string literal", "'foobar'", "foobar"},
		{"string literal with escaped single quote", "'foo''s bar'", "foo's bar"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := NewExpression(tt.input)
			if err != nil {
				t.Errorf("Expected no error, but got %s for input: %s", err.Error(), tt.input)
			}

			result, err := expr.Evaluate(nil)
			if err != nil {
				t.Errorf("Expected no error, but got %s for input: %s", err.Error(), tt.input)
			}

			if result != tt.expected {
				t.Errorf("Expected %v, but got %v for input: %s", tt.expected, result, tt.input)
			}
		})
	}
}

func TestExpression_EvaluateCompareOp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"equal", "1 == 1", true},
		{"not equal", "1 != 1", false},
		{"greater than", "1 > 1", false},
		{"greater than or equal", "1 >= 1", true},
		{"less than", "1 < 1", false},
		{"less than or equal", "1 <= 1", true},
		{"and", "true && false", false},
		{"or", "true || false", true},
		{"not", "!true", false},
		{"parentheses", "(1 == 1)", true},
		{"complex", "1 == 1 && 2 == 2", true},
		{"complex with parentheses", "(1 == 1) && (2 == 2)", true},
		{"complex with parentheses and not", "(1 == 1) && !(2 == 2)", false},
		{"complex with parentheses and or", "(1 == 1) || (2 == 2)", true},
		{"complex with parentheses and or and not", "(1 == 1) || !(2 == 2)", true},
		{"not null", "!null", true},
		{"not empty", "!''", true},
		{"coercion null to boolean", "null == 0", true},
		{"coercion null to string", "null == ''", true},
		{"coercion string to boolean", "'' == false", true},
		{"coercion string to number", "'' == 0", true},
		{"coercion boolean to number", "true == 1", true},
		{"coercion boolean to string", "true == '1'", true},
		{"coercion number to string", "1 == '1'", true},
		{"coercion number to boolean", "1 == true", true},
		{"coercion number to boolean", "0 == false", true},
		{"case insensitive string", "'TesT' == 'test'", true},
		// TODO: add function calls as well
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := NewExpression(tt.input)
			if err != nil {
				t.Errorf("Expected no error, but got %s for input: %s", err.Error(), tt.input)
			}

			result, err := expr.Evaluate(nil)
			if err != nil {
				t.Errorf("Expected no error, but got %s for input: %s", err.Error(), tt.input)
			}

			if result != tt.expected {
				t.Errorf("Expected %v, but got %v for input: %s", tt.expected, result, tt.input)
			}
		})
	}
}

func TestExpression_EvaluateLogicalOp(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		// Logical AND - true
		{"true && true", "true && true", true},
		{"true && false", "true && false", false},
		{"true && null", "true && null", nil},
		{"true && 1", "true && 1", 1},
		{"true && 0", "true && 0", 0},
		{"true && ''", "true && ''", ""},
		{"true && 'foo'", "true && 'foo'", "foo"},
		{"true && '0'", "true && '0'", "0"},
		{"true && '1'", "true && '1'", "1"},
		{"true && 'true'", "true && 'true'", "true"},
		{"true && 10", "true && 10", 10},
		{"true && 0.1", "true && 0.1", 0.1},
		{"true && -1", "true && -1", -1},
		{"true && 0.0", "true && 0.0", 0},
		{"true && Infinity", "true && Infinity", math.Inf(1)},
		{"true && NaN", "true && NaN", math.NaN()},

		// Logical OR - true
		{"true || true", "true || true", true},
		{"true || false", "true || false", true},
		{"true || null", "true || null", true},
		{"true || 1", "true || 1", true},
		{"true || 0", "true || 0", true},
		{"true || ''", "true || ''", true},
		{"true || 'foo'", "true || 'foo'", true},
		{"true || '0'", "true || '0'", true},
		{"true || '1'", "true || '1'", true},
		{"true || 'true'", "true || 'true'", true},
		{"true || 10", "true || 10", true},
		{"true || 0.1", "true || 0.1", true},
		{"true || -1", "true || -1", true},
		{"true || 0.0", "true || 0.0", true},
		{"true || Infinity", "true || Infinity", true},
		{"true || NaN", "true || NaN", true},

		// Logical AND - false
		{"false && true", "false && true", false},
		{"false && false", "false && false", false},
		{"false && null", "false && null", false},
		{"false && 1", "false && 1", false},
		{"false && 0", "false && 0", false},
		{"false && ''", "false && ''", false},
		{"false && 'foo'", "false && 'foo'", false},
		{"false && '0'", "false && '0'", false},
		{"false && '1'", "false && '1'", false},
		{"false && 'true'", "false && 'true'", false},
		{"false && 10", "false && 10", false},
		{"false && 0.1", "false && 0.1", false},
		{"false && -1", "false && -1", false},
		{"false && 0.0", "false && 0.0", false},
		{"false && Infinity", "false && Infinity", false},
		{"false && NaN", "false && NaN", false},

		// Logical OR - false
		{"false || true", "false || true", true},
		{"false || false", "false || false", false},
		{"false || null", "false || null", nil},
		{"false || 1", "false || 1", 1},
		{"false || 0", "false || 0", 0},
		{"false || ''", "false || ''", ""},
		{"false || 'foo'", "false || 'foo'", "foo"},
		{"false || '0'", "false || '0'", "0"},
		{"false || '1'", "false || '1'", "1"},
		{"false || 'true'", "false || 'true'", "true"},
		{"false || 10", "false || 10", 10},
		{"false || 0.1", "false || 0.1", 0.1},
		{"false || -1", "false || -1", -1},
		{"false || 0.0", "false || 0.0", 0},
		{"false || Infinity", "false || Infinity", math.Inf(1)},
		{"false || NaN", "false || NaN", math.NaN()},

		// Logical AND - null
		{"null && true", "null && true", nil},
		{"null && false", "null && false", nil},
		{"null && null", "null && null", nil},
		{"null && 1", "null && 1", nil},
		{"null && 0", "null && 0", nil},
		{"null && ''", "null && ''", nil},
		{"null && 'foo'", "null && 'foo'", nil},
		{"null && '0'", "null && '0'", nil},
		{"null && '1'", "null && '1'", nil},
		{"null && 'true'", "null && 'true'", nil},
		{"null && 10", "null && 10", nil},
		{"null && 0.1", "null && 0.1", nil},
		{"null && -1", "null && -1", nil},
		{"null && 0.0", "null && 0.0", nil},
		{"null && Infinity", "null && Infinity", nil},
		{"null && NaN", "null && NaN", nil},

		// Logical OR - null
		{"null || true", "null || true", true},
		{"null || false", "null || false", false},
		{"null || null", "null || null", nil},
		{"null || 1", "null || 1", 1},
		{"null || 0", "null || 0", 0},
		{"null || ''", "null || ''", ""},
		{"null || 'foo'", "null || 'foo'", "foo"},
		{"null || '0'", "null || '0'", "0"},
		{"null || '1'", "null || '1'", "1"},
		{"null || 'true'", "null || 'true'", "true"},
		{"null || 10", "null || 10", 10},
		{"null || 0.1", "null || 0.1", 0.1},
		{"null || -1", "null || -1", -1},
		{"null || 0.0", "null || 0.0", 0},
		{"null || Infinity", "null || Infinity", math.Inf(1)},
		{"null || NaN", "null || NaN", math.NaN()},

		// Logical AND - numeric

		{"1 && true", "1 && true", true},
		{"1 && false", "1 && false", false},
		{"1 && null", "1 && null", nil},
		{"1 && 1", "1 && 1", 1},
		{"1 && 0", "1 && 0", 0},
		{"1 && ''", "1 && ''", ""},
		{"1 && 'foo'", "1 && 'foo'", "foo"},
		{"1 && '0'", "1 && '0'", "0"},
		{"1 && '1'", "1 && '1'", "1"},
		{"1 && 'true'", "1 && 'true'", "true"},
		{"1 && 10", "1 && 10", 10},
		{"1 && 0.1", "1 && 0.1", 0.1},
		{"1 && -1", "1 && -1", -1},
		{"1 && 0.0", "1 && 0.0", 0},
		{"1 && Infinity", "1 && Infinity", math.Inf(1)},
		{"1 && NaN", "1 && NaN", math.NaN()},

		// Logical OR - numeric
		{"1 || true", "1 || true", 1},
		{"1 || false", "1 || false", 1},
		{"1 || null", "1 || null", 1},
		{"1 || 1", "1 || 1", 1},
		{"1 || 0", "1 || 0", 1},
		{"1 || ''", "1 || ''", 1},
		{"1 || 'foo'", "1 || 'foo'", 1},
		{"1 || '0'", "1 || '0'", 1},
		{"1 || '1'", "1 || '1'", 1},
		{"1 || 'true'", "1 || 'true'", 1},
		{"1 || 10", "1 || 10", 1},
		{"1 || 0.1", "1 || 0.1", 1},
		{"1 || -1", "1 || -1", 1},

		// Logical AND - string
		{"'foo' && true", "'foo' && true", true},
		{"'foo' && false", "'foo' && false", false},
		{"'foo' && null", "'foo' && null", nil},
		{"'foo' && 1", "'foo' && 1", 1},
		{"'foo' && 0", "'foo' && 0", 0},
		{"'foo' && ''", "'foo' && ''", ""},
		{"'foo' && 'foo'", "'foo' && 'foo'", "foo"},
		{"'foo' && '0'", "'foo' && '0'", "0"},
		{"'foo' && '1'", "'foo' && '1'", "1"},
		{"'foo' && 'true'", "'foo' && 'true'", "true"},
		{"'foo' && 10", "'foo' && 10", 10},
		{"'foo' && 0.1", "'foo' && 0.1", 0.1},
		{"'foo' && -1", "'foo' && -1", -1},
		{"'foo' && 0.0", "'foo' && 0.0", 0},
		{"'foo' && Infinity", "'foo' && Infinity", math.Inf(1)},
		{"'foo' && NaN", "'foo' && NaN", math.NaN()},

		// Logical OR - string
		{"'foo' || true", "'foo' || true", "foo"},
		{"'foo' || false", "'foo' || false", "foo"},
		{"'foo' || null", "'foo' || null", "foo"},
		{"'foo' || 1", "'foo' || 1", "foo"},
		{"'foo' || 0", "'foo' || 0", "foo"},
		{"'foo' || ''", "'foo' || ''", "foo"},
		{"'foo' || 'foo'", "'foo' || 'foo'", "foo"},
		{"'foo' || '0'", "'foo' || '0'", "foo"},
		{"'foo' || '1'", "'foo' || '1'", "foo"},
		{"'foo' || 'true'", "'foo' || 'true'", "foo"},
		{"'foo' || 10", "'foo' || 10", "foo"},
		{"'foo' || 0.1", "'foo' || 0.1", "foo"},
		{"'foo' || -1", "'foo' || -1", "foo"},
		{"'foo' || 0.0", "'foo' || 0.0", "foo"},
		{"'foo' || Infinity", "'foo' || Infinity", "foo"},
		{"'foo' || NaN", "'foo' || NaN", "foo"},

		// Logical AND - float

		{"0.1 && true", "0.1 && true", true},
		{"0.1 && false", "0.1 && false", false},
		{"0.1 && null", "0.1 && null", nil},
		{"0.1 && 1", "0.1 && 1", 1},
		{"0.1 && 0", "0.1 && 0", 0},
		{"0.1 && ''", "0.1 && ''", ""},
		{"0.1 && 'foo'", "0.1 && 'foo'", "foo"},
		{"0.1 && '0'", "0.1 && '0'", "0"},
		{"0.1 && '1'", "0.1 && '1'", "1"},
		{"0.1 && 'true'", "0.1 && 'true'", "true"},
		{"0.1 && 10", "0.1 && 10", 10},
		{"0.1 && 0.1", "0.1 && 0.1", 0.1},
		{"0.1 && -1", "0.1 && -1", -1},
		{"0.1 && 0.0", "0.1 && 0.0", 0},
		{"0.1 && Infinity", "0.1 && Infinity", math.Inf(1)},
		{"0.1 && NaN", "0.1 && NaN", math.NaN()},

		// Logical OR - float

		{"0.1 || true", "0.1 || true", 0.1},
		{"0.1 || false", "0.1 || false", 0.1},
		{"0.1 || null", "0.1 || null", 0.1},
		{"0.1 || 1", "0.1 || 1", 0.1},
		{"0.1 || 0", "0.1 || 0", 0.1},
		{"0.1 || ''", "0.1 || ''", 0.1},
		{"0.1 || 'foo'", "0.1 || 'foo'", 0.1},
		{"0.1 || '0'", "0.1 || '0'", 0.1},
		{"0.1 || '1'", "0.1 || '1'", 0.1},
		{"0.1 || 'true'", "0.1 || 'true'", 0.1},
		{"0.1 || 10", "0.1 || 10", 0.1},
		{"0.1 || 0.1", "0.1 || 0.1", 0.1},
		{"0.1 || -1", "0.1 || -1", 0.1},
		{"0.1 || 0.0", "0.1 || 0.0", 0.1},
		{"0.1 || Infinity", "0.1 || Infinity", 0.1},
		{"0.1 || NaN", "0.1 || NaN", 0.1},

		// Logical AND - Infinity

		{"Infinity && true", "Infinity && true", true},
		{"Infinity && false", "Infinity && false", false},
		{"Infinity && null", "Infinity && null", nil},
		{"Infinity && 1", "Infinity && 1", 1},
		{"Infinity && 0", "Infinity && 0", 0},
		{"Infinity && ''", "Infinity && ''", ""},
		{"Infinity && 'foo'", "Infinity && 'foo'", "foo"},
		{"Infinity && '0'", "Infinity && '0'", "0"},
		{"Infinity && '1'", "Infinity && '1'", "1"},
		{"Infinity && 'true'", "Infinity && 'true'", "true"},
		{"Infinity && 10", "Infinity && 10", 10},
		{"Infinity && 0.1", "Infinity && 0.1", 0.1},
		{"Infinity && -1", "Infinity && -1", -1},
		{"Infinity && 0.0", "Infinity && 0.0", 0},
		{"Infinity && Infinity", "Infinity && Infinity", math.Inf(1)},
		{"Infinity && NaN", "Infinity && NaN", math.NaN()},

		// Logical OR - Infinity

		{"Infinity || true", "Infinity || true", math.Inf(1)},
		{"Infinity || false", "Infinity || false", math.Inf(1)},
		{"Infinity || null", "Infinity || null", math.Inf(1)},
		{"Infinity || 1", "Infinity || 1", math.Inf(1)},
		{"Infinity || 0", "Infinity || 0", math.Inf(1)},
		{"Infinity || ''", "Infinity || ''", math.Inf(1)},
		{"Infinity || 'foo'", "Infinity || 'foo'", math.Inf(1)},
		{"Infinity || '0'", "Infinity || '0'", math.Inf(1)},
		{"Infinity || '1'", "Infinity || '1'", math.Inf(1)},
		{"Infinity || 'true'", "Infinity || 'true'", math.Inf(1)},
		{"Infinity || 10", "Infinity || 10", math.Inf(1)},
		{"Infinity || 0.1", "Infinity || 0.1", math.Inf(1)},
		{"Infinity || -1", "Infinity || -1", math.Inf(1)},
		{"Infinity || 0.0", "Infinity || 0.0", math.Inf(1)},
		{"Infinity || Infinity", "Infinity || Infinity", math.Inf(1)},
		{"Infinity || NaN", "Infinity || NaN", math.Inf(1)},

		// Logical AND - NaN

		{"NaN && true", "NaN && true", math.NaN()},
		{"NaN && false", "NaN && false", math.NaN()},
		{"NaN && null", "NaN && null", math.NaN()},
		{"NaN && 1", "NaN && 1", math.NaN()},
		{"NaN && 0", "NaN && 0", math.NaN()},
		{"NaN && ''", "NaN && ''", math.NaN()},
		{"NaN && 'foo'", "NaN && 'foo'", math.NaN()},
		{"NaN && '0'", "NaN && '0'", math.NaN()},
		{"NaN && '1'", "NaN && '1'", math.NaN()},
		{"NaN && 'true'", "NaN && 'true'", math.NaN()},
		{"NaN && 10", "NaN && 10", math.NaN()},
		{"NaN && 0.1", "NaN && 0.1", math.NaN()},
		{"NaN && -1", "NaN && -1", math.NaN()},
		{"NaN && 0.0", "NaN && 0.0", math.NaN()},
		{"NaN && Infinity", "NaN && Infinity", math.NaN()},
		{"NaN && NaN", "NaN && NaN", math.NaN()},

		// Logical OR - NaN

		{"NaN || true", "NaN || true", true},
		{"NaN || false", "NaN || false", false},
		{"NaN || null", "NaN || null", nil},
		{"NaN || 1", "NaN || 1", 1},
		{"NaN || 0", "NaN || 0", 0},
		{"NaN || ''", "NaN || ''", ""},
		{"NaN || 'foo'", "NaN || 'foo'", "foo"},
		{"NaN || '0'", "NaN || '0'", "0"},
		{"NaN || '1'", "NaN || '1'", "1"},
		{"NaN || 10", "NaN || 10", 10},
		{"NaN || 0.1", "NaN || 0.1", 0.1},
		{"NaN || -1", "NaN || -1", -1},
		{"NaN || 0.0", "NaN || 0.0", 0},
		{"NaN || Infinity", "NaN || Infinity", math.Inf(1)},
		{"NaN || NaN", "NaN || NaN", math.NaN()},

		// Logical AND - empty string

		{"'' && true", "'' && true", ""},
		{"'' && false", "'' && false", ""},
		{"'' && null", "'' && null", ""},
		{"'' && 1", "'' && 1", ""},
		{"'' && 0", "'' && 0", ""},
		{"'' && ''", "'' && ''", ""},
		{"'' && 'foo'", "'' && 'foo'", ""},
		{"'' && '0'", "'' && '0'", ""},
		{"'' && '1'", "'' && '1'", ""},
		{"'' && 'true'", "'' && 'true'", ""},
		{"'' && 10", "'' && 10", ""},
		{"'' && 0.1", "'' && 0.1", ""},
		{"'' && -1", "'' && -1", ""},
		{"'' && 0.0", "'' && 0.0", ""},
		{"'' && Infinity", "'' && Infinity", ""},
		{"'' && NaN", "'' && NaN", ""},

		// Logical OR - empty string

		{"'' || true", "'' || true", true},
		{"'' || false", "'' || false", false},
		{"'' || null", "'' || null", nil},
		{"'' || 1", "'' || 1", 1},
		{"'' || 0", "'' || 0", 0},
		{"'' || ''", "'' || ''", ""},
		{"'' || 'foo'", "'' || 'foo'", "foo"},
		{"'' || '0'", "'' || '0'", "0"},
		{"'' || '1'", "'' || '1'", "1"},
		{"'' || 'true'", "'' || 'true'", "true"},
		{"'' || 10", "'' || 10", 10},
		{"'' || 0.1", "'' || 0.1", 0.1},
		{"'' || -1", "'' || -1", -1},
		{"'' || 0.0", "'' || 0.0", 0},
		{"'' || Infinity", "'' || Infinity", math.Inf(1)},
		{"'' || NaN", "'' || NaN", math.NaN()},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := NewExpression(tt.input)
			if err != nil {
				t.Errorf("Expected no error, but got %s for input: %s", err.Error(), tt.input)
			}

			result, err := expr.Evaluate(&TestVariableProvider{})
			if err != nil {
				t.Errorf("Expected no error, but got %s for input: %s", err.Error(), tt.input)
			}

			if expected, ok := tt.expected.(float64); ok && math.IsNaN(expected) {
				if !math.IsNaN(result.(float64)) {
					t.Errorf("Expected %v, but got %v for input: %s", tt.expected, result, tt.input)
				}
			} else if result != tt.expected {
				t.Errorf("Expected %v, but got %v for input: %s", tt.expected, result, tt.input)
			}
		})
	}
}

func TestExpression_EvaluateContexts(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"foo.bar", "foo.bar", "baz"},
		{"foo.nested.some", "foo.nested.some", "value"},
		{"foo.nested.slice[0].foo", "foo.nested.slice[0].foo", "bar"},
		{"foo.nested.slice[1].foo", "foo.nested.slice[1].foo", "baz"},
		{"foo.nested.slice[2].foo", "foo.nested.slice[2].foo", "qux"},
		{"foo.nested.slice[3].foo", "foo.nested.slice[3].foo", nil},
		{"foo.nested.slice[0].bar", "foo.nested.slice[0].bar", nil},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := NewExpression(tt.input)
			if err != nil {
				t.Errorf("Expected no error, but got %s for input: %s", err.Error(), tt.input)
			}

			result, err := expr.Evaluate(&TestVariableProvider{})
			if err != nil {
				t.Errorf("Expected no error, but got %s for input: %s", err.Error(), tt.input)
			}

			if tt.expected != result {
				t.Errorf("Expected %v, but got %v for input: %s", tt.expected, result, tt.input)
			}
		})
	}
}

func TestExpression_EvaluateFunctions(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected interface{}
	}{
		{"contains string", "contains('foo', 'bar')", false},
		{"not contains string", "contains('foo', 'bar')", false},
		{"contains in slice", "contains(foo.nested.slice.*.foo, 'bar')", true},
		{"not contains in slice", "contains(foo.nested.slice.*.foo, 'tux')", false},
		{"contains with fromJson", "contains(fromJson('[\"foo\", \"bar\"]'), 'bar')", true},
		{"not contains with fromJson", "contains(fromJson('[\"foo\", \"bar\"]'), 'tux')", false},
		{"startsWith string", "startsWith('Hello World', 'Hell')", true},
		{"not startsWith string", "startsWith('foo', 'bar')", false},
		{"endsWith string", "endsWith('Hello World', 'World')", true},
		{"not endsWith string", "endsWith('foo', 'bar')", false},
		{"format string", "format('Hello {0}', 'World')", "Hello World"},
		{"format with escaped string", "format('{{ Hello {0}{1} }}', 'World', '!')", "{ Hello World! }"},
		{"join string", "join(foo.nested.slice.*.foo, '|')", "bar|baz|qux"},
		{"join with fromJson", "join(fromJson('[\"foo\", \"bar\"]'), ',')", "foo,bar"},
		{"join without separator", "join(foo.nested.slice.*.foo)", "bar,baz,qux"},
		{"toJson string", "toJson('foo')", "\"foo\""},
		{"toJson number", "toJson(1)", "1"},
		{"toJson boolean", "toJson(true)", "true"},
		{"toJson slice", "toJson(foo.nested.slice.*.foo)", "[\"bar\",\"baz\",\"qux\"]"},
		{"toJson map", "toJson(foo.nested)", "{\"slice\":[{\"foo\":\"bar\"},{\"foo\":\"baz\"},{\"foo\":\"qux\"}],\"some\":\"value\"}"},
		{"fromJson number", "fromJson('1')", 1.0},
		{"fromJson boolean", "fromJson('true')", true},
		{"always()", "always()", true},
		{"success()", "success()", true},
		{"cancelled()", "cancelled()", false},
		{"failure()", "failure()", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			expr, err := NewExpression(tt.input)
			if err != nil {
				t.Errorf("Expected no error, but got %s for input: %s", err.Error(), tt.input)
			}

			result, err := expr.Evaluate(&TestVariableProvider{})
			if err != nil {
				t.Errorf("Expected no error, but got %s for input: %s", err.Error(), tt.input)
			}

			if tt.expected != result {
				t.Errorf("Expected %v, but got %v for input: %s", tt.expected, result, tt.input)
			}
		})
	}
}

func TestExpression_EvaluateHashFunc(t *testing.T) {
	testCases := []struct {
		name     string
		input    string
		expected string
		files    []struct {
			path    string
			name    string
			content []byte
		}
	}{
		{
			name:     "Hash of existing files matching pattern",
			input:    `hashFiles('./testdata/file-*.txt')`,
			expected: "7f83b1657ff1fc53b92dc18148a1d65dfc2d4b1fa3d677284addd200126d9069",
			files: []struct {
				path    string
				name    string
				content []byte
			}{
				{name: "file-1.txt", content: []byte("Hello World!")},
			},
		},
		{
			name:     "multiple files matching",
			input:    `hashFiles('./testdata/file-*.txt')`,
			expected: "7f83b1657ff1fc53b92dc18148a1d65dfc2d4b1fa3d677284addd200126d9069b86691d0513ea10d252be103dba3459b45f4d4bb3aa5b07bd5460f5ef0357fc8",
			files: []struct {
				path    string
				name    string
				content []byte
			}{
				{name: "file-1.txt", content: []byte("Hello World!")},
				{name: "file-2.txt", content: []byte("Foo Bar!")},
			},
		},
		{
			name:     "Hash of nested directory",
			input:    `hashFiles('./testdata/nested/**/file-*.txt')`,
			expected: "14d53b936bfc2c45b26e84f74742f32f1449bab0ee83f6bb3410ef87b3a4d4beb86691d0513ea10d252be103dba3459b45f4d4bb3aa5b07bd5460f5ef0357fc8",
			files: []struct {
				path    string
				name    string
				content []byte
			}{
				{path: "nested", name: "file-1.txt", content: []byte("Hello World!")},
				{path: "nested/foo", name: "file-2.txt", content: []byte("Foo Bar!")},
				{path: "nested/bar", name: "file-3.txt", content: []byte("Bar Foo!")},
			},
		},
		{
			name:     "Hash of non-existing file",
			input:    `hashFiles('./testdata/non-extant-file.txt')`,
			expected: "",
		},
		{
			name:     "Hash of multiple non-existing files",
			input:    `hashFiles('**/non-extant-files', '**/more-non-extant-files')`,
			expected: "",
		},
	}

	for _, tt := range testCases {
		t.Run(tt.name, func(t *testing.T) {
			// remove directory first to ensure clean state
			err := os.RemoveAll("./testdata")
			if err != nil && !os.IsNotExist(err) {
				t.Errorf("Expected no error, but got %s for input: %s", err.Error(), tt.input)
			}

			for _, file := range tt.files {
				path := filepath.Join("./testdata", file.path)

				// re-create directory
				err = os.MkdirAll(path, 0755)
				if err != nil {
					t.Errorf("Expected no error, but got %s for input: %s", err.Error(), tt.input)
				}

				// create file
				err = os.WriteFile(filepath.Join(path, file.name), file.content, 0600)
				if err != nil {
					t.Errorf("Expected no error, but got %s for input: %s", err.Error(), tt.input)
				}
			}

			expr, err := NewExpression(tt.input)
			if err != nil {
				t.Errorf("Expected no error, but got %s for input: %s", err.Error(), tt.input)
			}

			result, err := expr.Evaluate(&TestVariableProvider{})
			if err != nil {
				t.Errorf("Expected no error, but got %s for input: %s", err.Error(), tt.input)
			}

			if tt.expected != result {
				t.Errorf("Expected %v, but got %v for input: %s", tt.expected, result, tt.input)
			}
		})
	}
}

// TODO: find a better way. Currently tests are relying on static values. It's not maintainable for long term.

type TestVariableProvider struct{}

func (p *TestVariableProvider) GetVariable(name string) (interface{}, error) {
	switch name {
	case "job":
		return map[string]interface{}{"status": "success"}, nil
	case "foo":
		return map[string]interface{}{
			"bar": "baz",
			"nested": map[string]interface{}{
				"some": "value",
				"slice": []map[string]interface{}{
					{"foo": "bar"},
					{"foo": "baz"},
					{"foo": "qux"},
				},
			},
		}, nil
	case "infinity":
		return math.Inf(1), nil
	case "nan":
		return math.NaN(), nil
	}

	return nil, fmt.Errorf("variable %s not found", name)
}
