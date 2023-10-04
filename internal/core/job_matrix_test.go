package core

import (
	"testing"

	"gopkg.in/yaml.v3"

	"github.com/stretchr/testify/assert"
)

func TestMatrixCombination_IsSubsetOf(t *testing.T) {
	combo1 := MatrixCombination{"OS": "ubuntu", "version": "18.04"}
	combo2 := MatrixCombination{"OS": "ubuntu", "version": "18.04", "arch": "x64"}

	assert.True(t, combo1.IsSubsetOf(combo2))
	assert.False(t, combo2.IsSubsetOf(combo1))
}

func TestMatrixCombination_KeysAndValuesMatch(t *testing.T) {
	combo1 := MatrixCombination{"OS": "ubuntu", "version": "18.04"}
	combo2 := MatrixCombination{"OS": "ubuntu", "version": "20.04"}

	assert.True(t, combo1.KeysAndValuesMatch(combo2, []string{"OS"}))
	assert.False(t, combo1.KeysAndValuesMatch(combo2, []string{"version"}))
}

func TestMatrix_GenerateCombinations(t *testing.T) {
	m := &Matrix{
		Dimensions: map[string]MatrixDimension{
			"OS":      {Key: "OS", Values: []interface{}{"ubuntu", "windows"}},
			"version": {Key: "version", Values: []interface{}{"18.04", "20.04"}},
		},
		Include: []MatrixCombination{{"OS": "ubuntu", "version": "20.04", "arch": "x64"}},
		Exclude: []MatrixCombination{{"OS": "windows", "version": "18.04"}},
	}

	expected := []MatrixCombination{
		{"OS": "ubuntu", "version": "18.04"},
		{"OS": "ubuntu", "version": "20.04", "arch": "x64"},
		{"OS": "windows", "version": "20.04"},
	}

	assert.ElementsMatch(t, expected, m.GenerateCombinations())
}

func TestMatrix_UnmarshalYAML(t *testing.T) {
	yamlStr := `
  fruit: [apple, pear]
  animal: [cat, dog]
  include:
    - color: green
    - color: pink
      animal: cat
    - fruit: apple
      shape: circle
    - fruit: banana
    - fruit: banana
      animal: cat
  exclude:
    - animal: cat
`

	var actual Matrix
	if err := yaml.Unmarshal([]byte(yamlStr), &actual); err != nil {
		t.Fatalf("Failed to unmarshal YAML: %v", err)
	}

	expected := Matrix{
		Dimensions: map[string]MatrixDimension{
			"fruit":  {Key: "fruit", Values: []interface{}{"apple", "pear"}},
			"animal": {Key: "animal", Values: []interface{}{"cat", "dog"}},
		},
		Include: []MatrixCombination{
			{"color": "green"},
			{"color": "pink", "animal": "cat"},
			{"fruit": "apple", "shape": "circle"},
			{"fruit": "banana"},
			{"fruit": "banana", "animal": "cat"},
		},
		Exclude: []MatrixCombination{
			{"animal": "cat"},
		},
	}

	assert.Equal(t, expected, actual)
}

func TestMatrix_UnmarshalJSON(t *testing.T) {
	jsonStr := `
{
	  "fruit": ["apple", "pear"],
	  "animal": ["cat", "dog"],
	  "include": [
	    {"color": "green"},
	    {"color": "pink", "animal": "cat"},
	    {"fruit": "apple", "shape": "circle"},
	    {"fruit": "banana"},
	    {"fruit": "banana", "animal": "cat"}    
	  ],
	  "exclude": [
	    {"animal": "cat"}
      ]
}`

	var actual Matrix
	if err := actual.UnmarshalJSON([]byte(jsonStr)); err != nil {
		t.Fatalf("Failed to unmarshal JSON: %v", err)
	}

	expected := Matrix{
		Dimensions: map[string]MatrixDimension{
			"fruit":  {Key: "fruit", Values: []interface{}{"apple", "pear"}},
			"animal": {Key: "animal", Values: []interface{}{"cat", "dog"}},
		},
		Include: []MatrixCombination{
			{"color": "green"},
			{"color": "pink", "animal": "cat"},
			{"fruit": "apple", "shape": "circle"},
			{"fruit": "banana"},
			{"fruit": "banana", "animal": "cat"},
		},
		Exclude: []MatrixCombination{
			{"animal": "cat"},
		},
	}

	assert.Equal(t, expected, actual)
}
