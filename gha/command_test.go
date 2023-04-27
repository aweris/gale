package gha

import (
	"reflect"
	"testing"
)

func TestParseCommand(t *testing.T) {
	tests := []struct {
		name           string
		input          string
		expectedResult *Command
		expectedMatch  bool
	}{
		{
			name:  "Command with parameters",
			input: "::set-env name=MY_VAR::some value",
			expectedResult: &Command{
				Name: "set-env",
				Parameters: map[string]string{
					"name": "MY_VAR",
				},
				Value: "some value",
			},
			expectedMatch: true,
		},
		{
			name:  "Command with multiple parameters",
			input: "::my-command foo=bar,baz=qux::some value",
			expectedResult: &Command{
				Name: "my-command",
				Parameters: map[string]string{
					"foo": "bar",
					"baz": "qux",
				},
				Value: "some value",
			},
			expectedMatch: true,
		},
		{
			name:  "Command without parameters",
			input: "::my-command::some value",
			expectedResult: &Command{
				Name:       "my-command",
				Parameters: map[string]string{},
				Value:      "some value",
			},
			expectedMatch: true,
		},
		{
			name:  "Command with empty value",
			input: "::my-command::",
			expectedResult: &Command{
				Name:       "my-command",
				Parameters: map[string]string{},
				Value:      "",
			},
			expectedMatch: true,
		},
		{
			name:           "Invalid command format",
			input:          "This is not a valid command",
			expectedResult: nil,
			expectedMatch:  false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			match, result := ParseCommand(tt.input)

			if match != tt.expectedMatch {
				t.Errorf("Expected match %v, but got %v", tt.expectedMatch, match)
			}

			if !reflect.DeepEqual(result, tt.expectedResult) {
				t.Errorf("Expected result %+v, but got %+v", tt.expectedResult, result)
			}
		})
	}
}
