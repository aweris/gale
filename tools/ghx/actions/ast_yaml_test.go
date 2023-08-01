package actions

import (
	"reflect"
	"testing"

	"gopkg.in/yaml.v3"
)

func TestMarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		value    yaml.Marshaler
		expected interface{}
	}{
		{
			name:     "bool with value",
			value:    NewBool(true),
			expected: true,
		},
		{
			name:     "bool with expression",
			value:    NewBoolExpr("${{ example }}"),
			expected: "${{ example }}",
		},
		{
			name:     "int with value",
			value:    NewInt(42),
			expected: 42,
		},
		{
			name:     "int with expression",
			value:    NewIntExpr("${{ example }}"),
			expected: "${{ example }}",
		},
		{
			name:     "float with value",
			value:    NewFloat(3.14),
			expected: 3.14,
		},
		{
			name:     "float with expression",
			value:    NewFloatExpr("${{ example }}"),
			expected: "${{ example }}",
		},
		{
			name:     "string with value",
			value:    NewString("foobar"),
			expected: "foobar",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.value.MarshalYAML()

			if err != nil {
				t.Errorf("Expected no error in MarshalYAML, but got %s", err.Error())
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %s, but got %s", tt.expected, result)
			}
		})
	}
}

func TestUnmarshalYAML(t *testing.T) {
	tests := []struct {
		name     string
		node     *yaml.Node
		expected yaml.Unmarshaler
	}{
		{
			name:     "bool with value",
			node:     &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!bool", Value: "true"},
			expected: NewBool(true),
		},
		{
			name:     "bool with expression",
			node:     &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "${{ example }}"},
			expected: NewBoolExpr("${{ example }}"),
		},
		{
			name:     "int with value",
			node:     &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!int", Value: "42"},
			expected: NewInt(42),
		},
		{
			name:     "int with expression",
			node:     &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "${{ example }}"},
			expected: NewIntExpr("${{ example }}"),
		},
		{
			name:     "float with value",
			node:     &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!float", Value: "3.14"},
			expected: NewFloat(3.14),
		},
		{
			name:     "float with expression",
			node:     &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "${{ example }}"},
			expected: NewFloatExpr("${{ example }}"),
		},
		{
			name:     "string with value",
			node:     &yaml.Node{Kind: yaml.ScalarNode, Tag: "!!str", Value: "foobar"},
			expected: NewString("foobar"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reflect.New(reflect.TypeOf(tt.expected).Elem()).Interface().(yaml.Unmarshaler)
			err := result.UnmarshalYAML(tt.node)

			if err != nil {
				t.Errorf("Expected no error in UnmarshalYAML, but got %s", err.Error())
			}

			if reflect.TypeOf(result) != reflect.TypeOf(tt.expected) {
				t.Errorf("Expected %s in UnmarshalYAML, but got %s", reflect.TypeOf(tt.expected), reflect.TypeOf(result))
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %s, but got %s", tt.expected, result)
			}
		})
	}
}
