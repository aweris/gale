package expression_test

import (
	"fmt"
	"math"
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/aweris/gale/internal/expression"
	"github.com/aweris/gale/internal/gctx"
)

var _ expression.VariableProvider = new(TestContext)

type TestContext struct {
	Github gctx.GithubContext
}

func (c *TestContext) GetVariable(name string) (interface{}, error) {
	switch name {
	case "github":
		return c.Github, nil
	case "infinity":
		return math.Inf(1), nil
	case "nan":
		return math.NaN(), nil
	}

	return nil, fmt.Errorf("unknown variable: %s", name)
}

func TestString_Eval(t *testing.T) {
	ctx := TestContext{
		Github: gctx.GithubContext{
			Token: "1234567890",
		},
	}

	tests := []struct {
		name     string
		value    string
		expected string
	}{
		{"simple string", "foobar", "foobar"},
		{"expression literal string", "${{'foobar'}}", "foobar"},
		{"single expression", "${{ github.token }}", "1234567890"},
		{"inline expression", "foobar-${{ github.token }}-baz", "foobar-1234567890-baz"},
		{"multiple expressions", "foobar-${{ github.token }}-${{ github.token }}-baz", "foobar-1234567890-1234567890-baz"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var str expression.String

			err := yaml.Unmarshal([]byte(tt.value), &str)
			if err != nil {
				t.Errorf("Expected no error, but got %s", err.Error())
			}

			result, err := str.Eval(&ctx)
			if err != nil {
				t.Errorf("Expected no error, but got %s", err.Error())
			}

			if result != tt.expected {
				t.Errorf("Expected %s, but got %s", tt.expected, result)
			}
		})
	}
}

func TestBool_Eval(t *testing.T) {
	ctx := TestContext{
		Github: gctx.GithubContext{
			Token: "1234567890",
		},
	}

	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{"raw boolean value", "true", true},
		{"expression with true value", "false", false},
		{"expression with false value", "${{ true }}", true},
		{"compare with true value", "${{ github.token == '1234567890' }}", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bv expression.Bool

			err := yaml.Unmarshal([]byte(tt.value.(string)), &bv)
			if err != nil {
				t.Errorf("Expected no error, but got %s", err.Error())
			}

			result, err := bv.Eval(&ctx)
			if err != nil {
				t.Errorf("Expected no error, but got %s", err.Error())
			}

			if result != tt.expected {
				t.Errorf("Expected %t, but got %t", tt.expected, result)
			}
		})
	}
}

func TestInt_Eval(t *testing.T) {
	ctx := TestContext{
		Github: gctx.GithubContext{
			Token: "1234567890",
		},
	}

	tests := []struct {
		name     string
		value    interface{}
		expected int
	}{
		{"raw integer value", "123", 123},
		{"expression with integer value", "${{ 123 }}", 123},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bv expression.Int

			err := yaml.Unmarshal([]byte(tt.value.(string)), &bv)
			if err != nil {
				t.Errorf("Expected no error, but got %s", err.Error())
			}

			result, err := bv.Eval(&ctx)
			if err != nil {
				t.Errorf("Expected no error, but got %s", err.Error())
			}

			if result != tt.expected {
				t.Errorf("Expected %d, but got %d", tt.expected, result)
			}
		})
	}
}

func TestFloat_Eval(t *testing.T) {
	ctx := TestContext{
		Github: gctx.GithubContext{
			Token: "1234567890",
		},
	}

	tests := []struct {
		name     string
		value    interface{}
		expected float64
	}{
		{"raw float value", "123.456", 123.456},
		{"expression with float value", "${{ 123.456 }}", 123.456},
		{"expression with infinity value", "${{ infinity }}", math.Inf(1)},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var bv expression.Float

			err := yaml.Unmarshal([]byte(tt.value.(string)), &bv)
			if err != nil {
				t.Errorf("Expected no error, but got %s", err.Error())
			}

			result, err := bv.Eval(&ctx)
			if err != nil {
				t.Errorf("Expected no error, but got %s", err.Error())
			}

			if result != tt.expected {
				t.Errorf("Expected %f, but got %f", tt.expected, result)
			}
		})
	}
}
