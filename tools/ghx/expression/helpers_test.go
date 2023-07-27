package expression

import (
	"math"
	"reflect"
	"testing"

	"github.com/rhysd/actionlint"
)

func TestGetPropertySlice(t *testing.T) {
	m1 := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	m2 := map[string]interface{}{
		"key1": "value4",
		"key2": "value5",
		"key3": "value6",
	}
	slice := []map[string]interface{}{m1, m2}
	left := reflect.ValueOf(slice)

	//nolint:goconst // This is a test
	property := "key1"
	expected := []interface{}{"value1", "value4"}

	value, err := getPropertyValue(left, property)
	if err != nil {
		t.Errorf("Error retrieving property: %v", err)
	}

	if !reflect.DeepEqual(value, expected) {
		t.Errorf("Incorrect property value. Expected: %v, Got: %v", expected, value)
	}
}

func TestGetPropertyMap(t *testing.T) {
	m := map[string]interface{}{
		"key1": "value1",
		"key2": "value2",
		"key3": "value3",
	}
	left := reflect.ValueOf(m)

	//nolint:goconst // This is a test
	property := "key1"
	//nolint:goconst // This is a test
	expected := "value1"

	value, err := getPropertyValueFromMap(left, property)
	if err != nil {
		t.Errorf("Error retrieving property: %v", err)
	}

	if value != expected {
		t.Errorf("Incorrect property value. Expected: %v, Got: %v", expected, value)
	}
}

func TestGetPropertyStruct(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:"field_1"`
		Field2 int    `json:"field_2"`
	}
	s := TestStruct{"value1", 42}
	left := reflect.ValueOf(s)

	property := "field_1"
	expected := "value1"

	value, err := getPropertyValueFromStruct(left, property)
	if err != nil {
		t.Errorf("Error retrieving property: %v", err)
	}

	if value != expected {
		t.Errorf("Incorrect property value. Expected: %v, Got: %v", expected, value)
	}
}

func TestFindFieldIndexByJSONTag(t *testing.T) {
	type TestStruct struct {
		Field1 string `json:"field_1"`
		Field2 int    `json:"field_2"`
	}

	structType := reflect.TypeOf(TestStruct{})
	jsonTag := "field_2"
	expected := 1

	index := findFieldIndexByJSONTag(structType, jsonTag)

	if index != expected {
		t.Errorf("Incorrect field index. Expected: %d, Got: %d", expected, index)
	}
}

func TestUnwrapValue(t *testing.T) {
	value := reflect.ValueOf("test")
	expected := "test"

	unwrapped, err := unwrapValue(value)
	if err != nil {
		t.Errorf("Error unwrapping value: %v", err)
	}

	if unwrapped != expected {
		t.Errorf("Incorrect unwrapped value. Expected: %v, Got: %v", expected, unwrapped)
	}
}

func TestGetProperty(t *testing.T) {
	// Test case 1: Pointer
	t.Run("Pointer", func(t *testing.T) {
		m := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}
		left := reflect.ValueOf(&m)

		property := "key3"
		expected := "value3"

		value, err := getPropertyValue(left, property)
		if err != nil {
			t.Errorf("Error retrieving property: %v", err)
		}

		if value != expected {
			t.Errorf("Incorrect property value. Expected: %v, Got: %v", expected, value)
		}
	})

	// Test case 2: Property in slice
	t.Run("Slice", func(t *testing.T) {
		m1 := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}
		m2 := map[string]interface{}{
			"key1": "value4",
			"key2": "value5",
			"key3": "value6",
		}
		slice := []map[string]interface{}{m1, m2}
		left := reflect.ValueOf(slice)

		property := "key1"
		expected := []interface{}{"value1", "value4"}

		value, err := getPropertyValue(left, property)
		if err != nil {
			t.Errorf("Error retrieving property: %v", err)
		}

		if !reflect.DeepEqual(value, expected) {
			t.Errorf("Incorrect property value. Expected: %v, Got: %v", expected, value)
		}
	})

	t.Run("Map", func(t *testing.T) {
		// Test case 3: Property in map
		m := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}
		left := reflect.ValueOf(m)

		property := "key2"
		expected := "value2"

		value, err := getPropertyValue(left, property)
		if err != nil {
			t.Errorf("Error retrieving property: %v", err)
		}

		if value != expected {
			t.Errorf("Incorrect property value. Expected: %v, Got: %v", expected, value)
		}
	})

	t.Run("Struct", func(t *testing.T) {
		// Test case 4: Property in struct
		type TestStruct struct {
			Field1 string `json:"field_1"`
			Field2 int    `json:"field_2"`
		}
		s := TestStruct{"value1", 42}
		left := reflect.ValueOf(s)

		property := "field_1"
		expected := "value1"

		value, err := getPropertyValue(left, property)
		if err != nil {
			t.Errorf("Error retrieving property: %v", err)
		}

		if value != expected {
			t.Errorf("Incorrect property value. Expected: %v, Got: %v", expected, value)
		}
	})

	// Test case 5: Property not found
	t.Run("NotFound", func(t *testing.T) {
		m := map[string]interface{}{
			"key1": "value1",
			"key2": "value2",
			"key3": "value3",
		}
		slice := []map[string]interface{}{m}
		left := reflect.ValueOf(slice)

		property := "key4"
		var expected []interface{}

		value, err := getPropertyValue(left, property)
		if err != nil {
			t.Errorf("Error retrieving property: %v", err)
		}

		if !reflect.DeepEqual(value, expected) {
			t.Errorf("Incorrect property value. Expected: %v, Got: %v", expected, value)
		}
	})
}

func TestGetSafeValue(t *testing.T) {
	testCases := []struct {
		name     string
		value    reflect.Value
		expected interface{}
	}{
		{
			name:     "Valid value",
			value:    reflect.ValueOf("test"),
			expected: "test",
		},
		{
			name:     "Invalid value",
			value:    reflect.ValueOf(nil),
			expected: nil,
		},
		{
			name:     "Float64 value = 0",
			value:    reflect.ValueOf(float64(0)),
			expected: 0,
		},
		{
			name:     "Float64 value != 0",
			value:    reflect.ValueOf(1.23),
			expected: 1.23,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := getSafeValue(tc.value)
			if result != tc.expected {
				t.Errorf("Incorrect result. Expected: %v, Got: %v", tc.expected, result)
			}
		})
	}
}

func TestIsTruthy(t *testing.T) {
	testCases := []struct {
		name     string
		input    interface{}
		expected bool
	}{
		{
			name:     "Bool true",
			input:    true,
			expected: true,
		},
		{
			name:     "Bool false",
			input:    false,
			expected: false,
		},
		{
			name:     "Non-empty string",
			input:    "hello",
			expected: true,
		},
		{
			name:     "Empty string",
			input:    "",
			expected: false,
		},
		{
			name:     "Non-zero integer",
			input:    42,
			expected: true,
		},
		{
			name:     "Zero integer",
			input:    0,
			expected: false,
		},
		{
			name:     "Non-zero float",
			input:    3.14,
			expected: true,
		},
		{
			name:     "Zero float",
			input:    0.0,
			expected: false,
		},
		{
			name:     "NaN float",
			input:    math.NaN(),
			expected: false,
		},
		{
			name:     "Map",
			input:    map[string]int{"a": 1},
			expected: true,
		},
		{
			name:     "Slice",
			input:    []string{"apple", "banana"},
			expected: true,
		},
		{
			name:     "Struct",
			input:    struct{ Name string }{Name: "John"},
			expected: false,
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			result := isTruthy(tc.input)
			if result != tc.expected {
				t.Errorf("Incorrect result. Expected: %v, Got: %v", tc.expected, result)
			}
		})
	}
}

func TestCompareValues(t *testing.T) {
	tests := []struct {
		name           string
		leftValue      interface{}
		rightValue     interface{}
		kind           actionlint.CompareOpNodeKind
		expectedResult interface{}
	}{
		{
			name:           "Bool_Less",
			leftValue:      true,
			rightValue:     false,
			kind:           actionlint.CompareOpNodeKindLess,
			expectedResult: false,
		},
		{
			name:           "String_Eq",
			leftValue:      "abc",
			rightValue:     "abc",
			kind:           actionlint.CompareOpNodeKindEq,
			expectedResult: true,
		},
		{
			name:           "Int_LessEq",
			leftValue:      10,
			rightValue:     20,
			kind:           actionlint.CompareOpNodeKindLessEq,
			expectedResult: true,
		},
		{
			name:           "Float_Less",
			leftValue:      3.14,
			rightValue:     2.718,
			kind:           actionlint.CompareOpNodeKindLess,
			expectedResult: false,
		},
		{
			name:           "Invalid_Invalid",
			leftValue:      nil,
			rightValue:     nil,
			kind:           actionlint.CompareOpNodeKindEq,
			expectedResult: true,
		},
		{
			name:           "Invalid_String",
			leftValue:      "abc",
			rightValue:     nil,
			kind:           actionlint.CompareOpNodeKindEq,
			expectedResult: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			leftValue := reflect.ValueOf(test.leftValue)
			rightValue := reflect.ValueOf(test.rightValue)

			result, err := compareValues(leftValue, rightValue, test.kind)

			if result != test.expectedResult {
				t.Errorf("Unexpected result. Expected: %v, Got: %v", test.expectedResult, result)
			}

			if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

func TestIsNumber(t *testing.T) {
	tests := []struct {
		name     string
		value    interface{}
		expected bool
	}{
		{
			name:     "Int",
			value:    10,
			expected: true,
		},
		{
			name:     "Float64",
			value:    3.14,
			expected: true,
		},
		{
			name:     "String",
			value:    "123",
			expected: false,
		},
		{
			name:     "Bool",
			value:    true,
			expected: false,
		},
		{
			name:     "Invalid",
			value:    nil,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := reflect.ValueOf(test.value)
			result := isNumber(value)

			if result != test.expected {
				t.Errorf("Unexpected result. Expected: %v, Got: %v", test.expected, result)
			}
		})
	}
}

func TestCoerceToNumber(t *testing.T) {
	tests := []struct {
		name         string
		value        interface{}
		expectedKind reflect.Kind
		expectedVal  interface{}
	}{
		{
			name:         "Invalid",
			value:        nil,
			expectedKind: reflect.Int,
			expectedVal:  0,
		},
		{
			name:         "Bool_True",
			value:        true,
			expectedKind: reflect.Int,
			expectedVal:  1,
		},
		{
			name:         "Bool_False",
			value:        false,
			expectedKind: reflect.Int,
			expectedVal:  0,
		},
		{
			name:         "String_Empty",
			value:        "",
			expectedKind: reflect.Int,
			expectedVal:  0,
		},
		{
			name:         "String_Valid",
			value:        "123.45",
			expectedKind: reflect.Float64,
			expectedVal:  123.45,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			value := coerceToNumber(reflect.ValueOf(test.value))

			if value.Kind() != test.expectedKind {
				t.Errorf("Unexpected kind. Expected: %v, Got: %v", test.expectedKind, value.Kind())
			}

			var val interface{}
			if value.Kind() == reflect.Float64 {
				val = value.Float()
			} else {
				val = int(value.Int())
			}

			if val != test.expectedVal {
				t.Errorf("Unexpected value. Expected: %v, Got: %v", test.expectedVal, val)
			}
		})
	}

	t.Run("String_Invalid", func(t *testing.T) {
		value := coerceToNumber(reflect.ValueOf("hello"))
		if !math.IsNaN(value.Float()) {
			t.Errorf("Expected NaN value, Got: %v", value.Float())
		}
	})
}

func TestCompare(t *testing.T) {
	tests := []struct {
		name     string
		left     interface{}
		right    interface{}
		kind     actionlint.CompareOpNodeKind
		expected bool
	}{
		{
			name:     "Less_Int",
			left:     float64(10),
			right:    float64(20),
			kind:     actionlint.CompareOpNodeKindLess,
			expected: true,
		},
		{
			name:     "LessEq_Int",
			left:     float64(10),
			right:    float64(20),
			kind:     actionlint.CompareOpNodeKindLessEq,
			expected: true,
		},
		{
			name:     "Greater_Int",
			left:     float64(20),
			right:    float64(10),
			kind:     actionlint.CompareOpNodeKindGreater,
			expected: true,
		},
		{
			name:     "GreaterEq_Int",
			left:     float64(20),
			right:    float64(10),
			kind:     actionlint.CompareOpNodeKindGreaterEq,
			expected: true,
		},
		{
			name:     "Eq_Int",
			left:     float64(10),
			right:    float64(10),
			kind:     actionlint.CompareOpNodeKindEq,
			expected: true,
		},
		{
			name:     "NotEq_Int",
			left:     float64(10),
			right:    float64(20),
			kind:     actionlint.CompareOpNodeKindNotEq,
			expected: true,
		},
		{
			name:     "Less_String",
			left:     "abc",
			right:    "def",
			kind:     actionlint.CompareOpNodeKindLess,
			expected: true,
		},
		{
			name:     "LessEq_String",
			left:     "abc",
			right:    "def",
			kind:     actionlint.CompareOpNodeKindLessEq,
			expected: true,
		},
		{
			name:     "Greater_String",
			left:     "def",
			right:    "abc",
			kind:     actionlint.CompareOpNodeKindGreater,
			expected: true,
		},
		{
			name:     "GreaterEq_String",
			left:     "def",
			right:    "abc",
			kind:     actionlint.CompareOpNodeKindGreaterEq,
			expected: true,
		},
		{
			name:     "Eq_String",
			left:     "abc",
			right:    "abc",
			kind:     actionlint.CompareOpNodeKindEq,
			expected: true,
		},
		{
			name:     "NotEq_String",
			left:     "abc",
			right:    "def",
			kind:     actionlint.CompareOpNodeKindNotEq,
			expected: true,
		},
		{
			name:     "False_Int",
			left:     float64(20),
			right:    float64(10),
			kind:     actionlint.CompareOpNodeKindLess,
			expected: false,
		},
		{
			name:     "False_String",
			left:     "def",
			right:    "abc",
			kind:     actionlint.CompareOpNodeKindLess,
			expected: false,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			var result bool
			var err error

			if _, ok := test.left.(string); ok {
				result, err = compare(test.left.(string), test.right.(string), test.kind)
				if err != nil {
					t.Errorf("Error comparing values: %v", err)
				}
			} else {
				result, err = compare(test.left.(float64), test.right.(float64), test.kind)
				if err != nil {
					t.Errorf("Error comparing values: %v", err)
				}
			}

			if result != test.expected {
				t.Errorf("Comparison result mismatch. Expected: %v, Got: %v", test.expected, result)
			}
		})
	}
}
