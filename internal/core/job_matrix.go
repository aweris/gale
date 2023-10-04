package core

import (
	"encoding/json"

	"gopkg.in/yaml.v3"
)

// MatrixDimension represents a single matrix dimension with its key and values
type MatrixDimension struct {
	Key    string
	Values []interface{}
}

// MatrixCombination represents a single combination of matrix dimensions
type MatrixCombination map[string]interface{}

// IsSubsetOf returns true if the given matrix combination is a subset of the other matrix combination
func (mc *MatrixCombination) IsSubsetOf(other MatrixCombination) bool {
	for k, v := range *mc {
		if otherVal, exists := other[k]; !exists || otherVal != v {
			return false
		}
	}

	return true
}

// KeysAndValuesMatch returns true if the given matrix combination has the same values for the given keys as the other
func (mc *MatrixCombination) KeysAndValuesMatch(other MatrixCombination, keys []string) bool {
	for _, key := range keys {
		if v1, ok := (*mc)[key]; ok {
			if v2, ok := other[key]; !ok || v1 != v2 {
				return false
			}
		}
	}
	return true
}

// Matrix represents a job matrix in a GitHub Actions workflow
type Matrix struct {
	Dimensions map[string]MatrixDimension // Dimensions is the list of matrix dimensions given in the workflow.
	Include    []MatrixCombination        // Include is the list of matrix combinations to update or extend the matrix.
	Exclude    []MatrixCombination        // Exclude is the list of matrix combinations to remove from the matrix.
}

// GenerateCombinations generates all possible combinations from the given matrix dimensions, includes and excludes
func (m *Matrix) GenerateCombinations() []MatrixCombination {
	// if there are no dimensions, just return an empty result
	if len(m.Dimensions) == 0 {
		return []MatrixCombination{}
	}

	var keys []string

	for k := range m.Dimensions {
		keys = append(keys, k)
	}

	combinations := generateCombinations(keys, m.Dimensions)

	if len(m.Exclude) > 0 {
		combinations = applyExclusions(combinations, m.Exclude)
	}

	if len(m.Include) > 0 {
		combinations = applyIncludes(combinations, m.Include, keys)
	}

	return combinations
}

// generateCombinations generates all possible combinations of given dimensions and keys
func generateCombinations(keys []string, dimensions map[string]MatrixDimension) []MatrixCombination {
	// get the first key and its dimension
	key := keys[0]
	dimension := dimensions[key]

	// get the rest of the keys
	subKeys := keys[1:]

	// if there are no more keys, just return the dimension values
	if len(subKeys) == 0 {
		// pre-allocate slice to avoid re-allocations
		result := make([]MatrixCombination, 0, len(dimension.Values))

		for _, val := range dimension.Values {
			result = append(result, MatrixCombination{dimension.Key: val})
		}

		return result
	}

	// otherwise, generate combinations for the rest of the keys
	subCombos := generateCombinations(subKeys, dimensions)

	// pre-allocate slice to avoid re-allocations
	result := make([]MatrixCombination, 0, len(dimension.Values)*len(subCombos))

	// combine the dimension values with the sub combinations
	for _, val := range dimension.Values {
		for _, subCombo := range subCombos {
			combo := MatrixCombination{dimension.Key: val}

			for k, v := range subCombo {
				combo[k] = v
			}

			result = append(result, combo)
		}
	}

	return result
}

// isExcluded returns true if the given combination has a match in the exclusions and should be excluded.
func isExcluded(combo MatrixCombination, exclusions []MatrixCombination) bool {
	for _, exclusion := range exclusions {
		// if all keys in the exclusion are present in the combination and their values match, then the combination
		// should be excluded
		if exclusion.IsSubsetOf(combo) {
			return true
		}
	}

	return false
}

// applyExclusions applies the given exclusions to the given combinations and returns the filtered combinations.
func applyExclusions(combinations []MatrixCombination, exclusions []MatrixCombination) []MatrixCombination {
	var result []MatrixCombination

	for _, combo := range combinations {
		if !isExcluded(combo, exclusions) {
			result = append(result, combo)
		}
	}

	return result
}

// applyIncludes applies the given includes to the given combinations and returns the updated combinations.
func applyIncludes(originalCombinations []MatrixCombination, includes []MatrixCombination, dimensions []string) []MatrixCombination {
	// copy original combinations to new combinations to avoid modifying the original combinations
	newCombinations := append([]MatrixCombination{}, originalCombinations...)

	for _, include := range includes {
		// if the include combination has all the keys in the dimensions, then we should update the combinations
		// that match with the include combination
		isUpdated := false

		for _, combination := range newCombinations {
			// if the original dimension keys matches with include and combination, then we should update the combination
			if include.KeysAndValuesMatch(combination, dimensions) {
				for k, v := range include {
					combination[k] = v
				}

				isUpdated = true
			}
		}

		// it means that include combination at least one match with given combinations and updated them with new values,
		// so we don't need to add it to new combinations
		if isUpdated {
			continue
		}

		// otherwise, we should add the include combination to new combinations. To avoid adding duplicate combinations,
		// we should check if the combination already exists in the original combinations.
		alreadyExists := false

		for _, originalCombination := range originalCombinations {
			if include.KeysAndValuesMatch(originalCombination, dimensions) {
				alreadyExists = true
				break
			}
		}

		// if the combination doesn't exist in the original combinations, then we should add it to new combinations
		if !alreadyExists {
			newCombinations = append(newCombinations, include)
		}
	}

	return newCombinations
}

// UnmarshalYAML implements yaml.Unmarshaler interface for Matrix
func (m *Matrix) UnmarshalYAML(node *yaml.Node) error {
	var raw map[string]interface{}

	if err := node.Decode(&raw); err != nil {
		return err
	}

	m.populate(raw)

	return nil
}

// UnmarshalJSON implements json.Unmarshaler interface for Matrix
func (m *Matrix) UnmarshalJSON(data []byte) error {
	var raw map[string]interface{}

	if err := json.Unmarshal(data, &raw); err != nil {
		return err
	}
	m.populate(raw)

	return nil
}

// populate populates the matrix from given raw data
func (m *Matrix) populate(raw map[string]interface{}) {
	for key, value := range raw {
		switch key {
		case "include":
			m.Include = append([]MatrixCombination{}, extractCombinations(value)...)
		case "exclude":
			m.Exclude = append([]MatrixCombination{}, extractCombinations(value)...)
		default:
			if m.Dimensions == nil {
				m.Dimensions = make(map[string]MatrixDimension)
			}

			if val, ok := value.([]interface{}); ok {
				m.Dimensions[key] = MatrixDimension{Key: key, Values: val}
			}
		}
	}
}

// extractCombinations extracts combinations from given value
func extractCombinations(value interface{}) []MatrixCombination {
	var result []MatrixCombination

	if combinations, ok := value.([]interface{}); ok {
		for _, combination := range combinations {
			if val, ok := combination.(map[string]interface{}); ok {
				result = append(result, convertToMatrixCombination(val))
			}
		}
	}

	return result
}

// convertToMatrixCombination converts given raw data to MatrixCombination object
func convertToMatrixCombination(raw map[string]interface{}) MatrixCombination {
	combo := MatrixCombination{}

	for k, v := range raw {
		combo[k] = v
	}

	return combo
}
