package gha

// Environment is a custom type representing a map of environment variables
type Environment map[string]string

// Merge merges two environments together. If a key exists in both environments, the value from
// the other environment will be used.
func (e Environment) Merge(other Environment) Environment {
	merged := Environment{}

	for k, v := range e {
		merged[k] = v
	}

	for k, v := range other {
		merged[k] = v
	}

	return merged
}
