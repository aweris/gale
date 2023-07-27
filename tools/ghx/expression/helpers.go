package expression

import (
	"encoding"
	"fmt"
	"math"
	"reflect"
	"strconv"
	"strings"

	"github.com/rhysd/actionlint"
)

// getPropertyValue retrieves the value of the specified property from the given value.
// The property can be accessed from struct, map, or slice types using dot notation.
// If the property is not found, nil is returned along with no error.
func getPropertyValue(left reflect.Value, property string) (value interface{}, err error) {
	switch left.Kind() {
	case reflect.Ptr:
		return getPropertyValue(left.Elem(), property)
	case reflect.Struct:
		return getPropertyValueFromStruct(left, property)
	case reflect.Map:
		return getPropertyValueFromMap(left, property)
	case reflect.Slice:
		return getPropertyValueFromSlice(left, property)
	}

	return nil, nil
}

// getPropertyValueFromSlice retrieves the values of the specified property from the elements of the given slice.
func getPropertyValueFromSlice(left reflect.Value, property string) (interface{}, error) {
	var values []interface{}

	for i := 0; i < left.Len(); i++ {
		value, err := getPropertyValue(left.Index(i), property)
		if err != nil {
			return nil, err
		}

		if value != nil {
			values = append(values, value)
		}
	}

	return values, nil
}

// getPropertyValueFromMap retrieves the value of the specified property from the given map.
func getPropertyValueFromMap(left reflect.Value, property string) (interface{}, error) {
	iter := left.MapRange()

	for iter.Next() {
		key := iter.Key()

		if key.Kind() != reflect.String {
			return nil, fmt.Errorf("'%s' in map key not implemented", key.Kind())
		}

		if strings.EqualFold(key.String(), property) {
			return unwrapValue(iter.Value())
		}
	}

	return nil, nil
}

// getPropertyValueFromStruct retrieves the value of the specified property from the given struct.
func getPropertyValueFromStruct(left reflect.Value, property string) (interface{}, error) {
	leftType := left.Type()
	fieldIndex := findFieldIndexByJSONTag(leftType, property)

	if fieldIndex < 0 {
		return nil, fmt.Errorf("property '%s' not found in struct", property)
	}

	fieldValue := left.Field(fieldIndex)

	if fieldValue.Kind() == reflect.Invalid {
		return nil, nil
	}

	i := fieldValue.Interface()

	if m, ok := i.(encoding.TextMarshaler); ok {
		text, err := m.MarshalText()
		if err != nil {
			return nil, err
		}
		return string(text), nil
	}

	return i, nil
}

// findFieldIndexByJSONTag finds the index of the field with the specified JSON tag in the given struct type.
func findFieldIndexByJSONTag(structType reflect.Type, jsonTag string) int {
	for i := 0; i < structType.NumField(); i++ {
		if structType.Field(i).Tag.Get("json") == jsonTag {
			return i
		}
	}

	return -1
}

// unwrapValue unwraps the underlying value of the given reflect.Value, de-referencing it if it is a pointer.
func unwrapValue(value reflect.Value) (interface{}, error) {
	if value.Kind() == reflect.Ptr {
		value = value.Elem()
	}

	return value.Interface(), nil
}

// getSafeValue returns the value of the given reflect.Value, or nil if the value is invalid.
func getSafeValue(value reflect.Value) interface{} {
	// If the value is invalid, return nil
	if !value.IsValid() {
		return nil
	}

	// if value is float64 and is 0, return 0
	if value.Kind() == reflect.Float64 {
		if value.Float() == 0 {
			return 0
		}
	}

	// return the value interface
	return value.Interface()
}

// isTruthy returns true if the given value is truthy, false otherwise.
func isTruthy(input interface{}) bool {
	value := reflect.ValueOf(input)
	switch value.Kind() {
	case reflect.Bool:
		return value.Bool()

	case reflect.String:
		return value.String() != ""

	case reflect.Int:
		return value.Int() != 0

	case reflect.Float64:
		if math.IsNaN(value.Float()) {
			return false
		}

		return value.Float() != 0

	case reflect.Map, reflect.Slice:
		return true

	default:
		return false
	}
}

// compareValues compares the given values using the specified comparison operator.
func compareValues(leftValue reflect.Value, rightValue reflect.Value, kind actionlint.CompareOpNodeKind) (interface{}, error) {
	if leftValue.Kind() != rightValue.Kind() {
		if !isNumber(leftValue) {
			leftValue = coerceToNumber(leftValue)
		}
		if !isNumber(rightValue) {
			rightValue = coerceToNumber(rightValue)
		}
	}

	switch leftValue.Kind() {
	case reflect.Bool:
		left := coerceToNumber(leftValue)
		right := coerceToNumber(rightValue)
		return compare(float64(left.Int()), float64(right.Int()), kind)

	case reflect.String:
		left := strings.ToLower(leftValue.String())
		right := strings.ToLower(rightValue.String())
		return compare(left, right, kind)

	case reflect.Int:
		left := float64(leftValue.Int())
		if rightValue.Kind() == reflect.Float64 {
			right := rightValue.Float()
			return compare(left, right, kind)
		}
		right := float64(rightValue.Int())
		return compare(left, right, kind)

	case reflect.Float64:
		left := leftValue.Float()
		if rightValue.Kind() == reflect.Int {
			right := float64(rightValue.Int())
			return compare(left, right, kind)
		}
		right := rightValue.Float()
		return compare(left, right, kind)

	case reflect.Invalid:
		if rightValue.Kind() == reflect.Invalid {
			return true, nil
		}
		return nil, fmt.Errorf("compare params of Invalid type: left: %+v, right: %+v", leftValue.Kind(), rightValue.Kind())

	default:
		return nil, fmt.Errorf("compare not implemented for types: left: %+v, right: %+v", leftValue.Kind(), rightValue.Kind())
	}
}

// isNumber returns true if the given value is a number, false otherwise.
func isNumber(value reflect.Value) bool {
	switch value.Kind() {
	case reflect.Int, reflect.Float64:
		return true
	default:
		return false
	}
}

// coerceToNumber converts the given value to a number if possible, otherwise returns NaN.
func coerceToNumber(value reflect.Value) reflect.Value {
	switch value.Kind() {
	case reflect.Invalid:
		return reflect.ValueOf(0)

	case reflect.Bool:
		switch value.Bool() {
		case true:
			return reflect.ValueOf(1)
		case false:
			return reflect.ValueOf(0)
		}

	case reflect.String:
		if value.String() == "" {
			return reflect.ValueOf(0)
		}

		if number, err := strconv.ParseFloat(value.String(), 64); err == nil {
			return reflect.ValueOf(number)
		}
	}

	return reflect.ValueOf(math.NaN())
}

// compare compares the given values using the specified comparison operator.
func compare[T string | float64](left, right T, kind actionlint.CompareOpNodeKind) (bool, error) {
	switch kind {
	case actionlint.CompareOpNodeKindLess:
		return left < right, nil
	case actionlint.CompareOpNodeKindLessEq:
		return left <= right, nil
	case actionlint.CompareOpNodeKindGreater:
		return left > right, nil
	case actionlint.CompareOpNodeKindGreaterEq:
		return left >= right, nil
	case actionlint.CompareOpNodeKindEq:
		return left == right, nil
	case actionlint.CompareOpNodeKindNotEq:
		return left != right, nil
	default:
		return false, fmt.Errorf("TODO: not implemented to compare '%+v'", kind)
	}
}
