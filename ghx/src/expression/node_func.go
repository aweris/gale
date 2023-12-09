package expression

import (
	"crypto/sha256"
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
	"reflect"
	"strings"

	"github.com/rhysd/actionlint"
)

var _ Interpreter = new(FuncCallNode)

// FuncCallNode is a wrapper of actionlint.FuncCallNode
type FuncCallNode actionlint.FuncCallNode

func (n FuncCallNode) Evaluate(provider VariableProvider) (interface{}, error) {
	callee := strings.ToLower(n.Callee)

	args := make([]reflect.Value, 0)

	for _, arg := range n.Args {
		val, err := getInterpreterFromNode(arg).Evaluate(provider)
		if err != nil {
			return nil, err
		}

		args = append(args, reflect.ValueOf(val))
	}

	switch callee {
	case "contains":
		return contains(args...)
	case "startswith":
		return startsWith(args...)
	case "endswith":
		return endsWith(args...)
	case "format":
		return format(args...)
	case "join":
		if len(args) == 1 {
			return join(args[0], reflect.ValueOf(","))
		}
		return join(args...)
	case "tojson":
		return toJSON(args...)
	case "fromjson":
		return fromJSON(args...)
	case "hashfiles":
		return hashFiles(args...)
	case "success":
		return success(provider)
	case "failure":
		return failure(provider)
	case "cancelled":
		return cancelled(provider)
	case "always":
		return always(), nil
	}

	return nil, fmt.Errorf("function '%s' not supported", n.Callee)
}

func contains(args ...reflect.Value) (bool, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("contains() requires two arguments")
	}

	searchValue := args[0]
	itemValue := args[1]

	if searchValue.Kind() == reflect.String {
		searchString := searchValue.String()
		if itemValue.Kind() == reflect.String {
			return strings.Contains(searchString, itemValue.String()), nil
		}
	} else if searchValue.Kind() == reflect.Slice || searchValue.Kind() == reflect.Array {
		for i := 0; i < searchValue.Len(); i++ {
			if reflect.DeepEqual(searchValue.Index(i).Interface(), itemValue.Interface()) {
				return true, nil
			}
		}
	}

	return false, nil
}

func startsWith(args ...reflect.Value) (bool, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("startsWith() requires two arguments")
	}

	searchString := args[0].String()
	searchValue := args[1].String()

	return strings.HasPrefix(searchString, searchValue), nil
}

func endsWith(args ...reflect.Value) (bool, error) {
	if len(args) != 2 {
		return false, fmt.Errorf("endsWith() requires two arguments")
	}

	searchString := args[0].String()
	searchValue := args[1].String()

	return strings.HasSuffix(searchString, searchValue), nil
}

func format(args ...reflect.Value) (string, error) {
	if len(args) < 2 {
		return "", fmt.Errorf("format() requires at least two arguments")
	}

	result := args[0].String()

	var values []interface{}
	if len(args) >= 2 {
		for i := 1; i < len(args); i++ {
			values = append(values, args[i].Interface())
		}
	}

	// Replace placeholders {N} with corresponding values
	for i := 0; i < len(values); i++ {
		placeholder := fmt.Sprintf("{%d}", i)
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", values[i]))
	}

	// Replace escaped placeholders {{N}} with {N}
	result = strings.NewReplacer("{{", "{", "}}", "}").Replace(result)

	return result, nil
}

func join(args ...reflect.Value) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("join() requires at least one argument")
	}

	var array reflect.Value
	var separator string

	if len(args) >= 1 {
		array = args[0]
	}

	if len(args) >= 2 {
		separator = args[1].String()
	} else {
		separator = ","
	}

	var values []string

	switch array.Kind() {
	case reflect.String:
		values = []string{array.String()}
	case reflect.Slice, reflect.Array:
		for i := 0; i < array.Len(); i++ {
			value := array.Index(i).Interface()
			values = append(values, fmt.Sprintf("%v", value))
		}
	}

	return strings.Join(values, separator), nil
}

func toJSON(args ...reflect.Value) (string, error) {
	if len(args) != 1 {
		return "", fmt.Errorf("toJSON() requires exactly one argument")
	}

	value := args[0].Interface()
	jsonBytes, err := json.Marshal(value)
	if err != nil {
		panic(fmt.Sprintf("Failed to convert value to JSON: %v", err))
	}

	return string(jsonBytes), nil
}

func fromJSON(args ...reflect.Value) (interface{}, error) {
	if len(args) != 1 {
		return nil, fmt.Errorf("fromJSON() requires exactly one argument")
	}

	jsonString := args[0].String()

	var value interface{}

	err := json.Unmarshal([]byte(jsonString), &value)
	if err != nil {
		panic(fmt.Sprintf("Failed to parse JSON: %v", err))
	}

	return value, nil
}

// TODO: double check this how GHA handles this. This is just a some hacky implementation to get it working

func hashFiles(args ...reflect.Value) (string, error) {
	if len(args) < 1 {
		return "", fmt.Errorf("hashFiles() requires at least one argument")
	}

	paths := make([]string, 0, len(args))

	for _, arg := range args {
		paths = append(paths, arg.String())
	}

	var fileHashes []string

	for _, path := range paths {
		matches, err := filepath.Glob(path)
		if err != nil {
			return "", err
		}

		for _, match := range matches {
			fileBytes, err := os.ReadFile(match)
			if err != nil {
				return "", err
			}

			hash := sha256.Sum256(fileBytes)
			fileHashes = append(fileHashes, fmt.Sprintf("%x", hash))
		}
	}

	return strings.Join(fileHashes, ""), nil
}

func always() bool {
	return true
}

// TODO: double check this how GHA handles this. This is just a some hacky implementation to get it working

func success(provider VariableProvider) (bool, error) {
	return evaluateStatusFunc(provider, "success")
}

func failure(provider VariableProvider) (bool, error) {
	return evaluateStatusFunc(provider, "failure")
}

func cancelled(provider VariableProvider) (bool, error) {
	return evaluateStatusFunc(provider, "cancelled")
}

func evaluateStatusFunc(provider VariableProvider, status string) (bool, error) {
	expr, err := NewExpression(fmt.Sprintf("${{ job.status == '%s' }}", status))
	if err != nil {
		return false, err
	}

	val, err := expr.Evaluate(provider)
	if err != nil {
		return false, err
	}

	if v, ok := val.(bool); ok {
		return v, nil
	}

	return false, fmt.Errorf("cannot evaluate expression: cancelled()")
}
