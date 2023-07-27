package actions

import (
	"encoding/json"
	"reflect"
	"testing"
)

func TestMarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		value    json.Marshaler
		expected string
	}{
		{
			name:     "bool with value",
			value:    NewBool(true),
			expected: "true",
		},
		{
			name:     "bool with expression",
			value:    NewBoolExpr("${{ example }}"),
			expected: `"${{ example }}"`,
		},
		{
			name:     "int with value",
			value:    NewInt(42),
			expected: "42",
		},
		{
			name:     "int with expression",
			value:    NewIntExpr("${{ example }}"),
			expected: `"${{ example }}"`,
		},
		{
			name:     "float with value",
			value:    NewFloat(3.14),
			expected: "3.14",
		},
		{
			name:     "float with expression",
			value:    NewFloatExpr("${{ example }}"),
			expected: `"${{ example }}"`,
		},
		{
			name:     "string with value",
			value:    NewString("foobar"),
			expected: `"foobar"`,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := tt.value.MarshalJSON()

			if err != nil {
				t.Errorf("Expected no error in MarshalJSON, but got %s", err.Error())
			}

			if string(result) != tt.expected {
				t.Errorf("Expected %s, but got %s", tt.expected, string(result))
			}
		})
	}
}

func TestUnmarshalJSON(t *testing.T) {
	tests := []struct {
		name     string
		data     []byte
		expected interface{}
	}{
		{
			name:     "bool with value",
			data:     []byte(`true`),
			expected: NewBool(true),
		},
		{
			name:     "bool with string value",
			data:     []byte(`"True"`),
			expected: NewBool(true),
		},
		{
			name:     "bool with uppercase string value",
			data:     []byte(`"FALSE"`),
			expected: NewBool(false),
		},
		{
			name:     "bool with numeric 1 string value",
			data:     []byte(`"1"`),
			expected: NewBool(true),
		},
		{
			name:     "bool with value numeric 0 string value",
			data:     []byte(`"0"`),
			expected: NewBool(false),
		},
		{
			name:     "bool with expression",
			data:     []byte(`"${{ example }}"`),
			expected: NewBoolExpr("${{ example }}"),
		},
		{
			name:     "int with value",
			data:     []byte(`42`),
			expected: NewInt(42),
		},
		{
			name:     "int with expression",
			data:     []byte(`"${{ example }}"`),
			expected: NewIntExpr("${{ example }}"),
		},
		{
			name:     "float with value",
			data:     []byte(`3.14`),
			expected: NewFloat(3.14),
		},
		{
			name:     "float with expression",
			data:     []byte(`"${{ example }}"`),
			expected: NewFloatExpr("${{ example }}"),
		},
		{
			name:     "string with value",
			data:     []byte("foobar"),
			expected: NewString("foobar"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := reflect.New(reflect.TypeOf(tt.expected).Elem()).Interface().(json.Unmarshaler)
			err := result.UnmarshalJSON(tt.data)

			if err != nil {
				t.Errorf("Expected no error in UnmarshalJSON, but got %s", err.Error())
			}

			if reflect.TypeOf(result) != reflect.TypeOf(tt.expected) {
				t.Errorf("Expected %s in UnmarshalJSON, but got %s", reflect.TypeOf(tt.expected), reflect.TypeOf(result))
			}

			if !reflect.DeepEqual(result, tt.expected) {
				t.Errorf("Expected %s, but got %s", tt.expected, result)
			}
		})
	}
}
