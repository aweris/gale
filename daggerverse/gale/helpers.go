package main

// withEmptyValue returns the given optional value or a new empty value if the given optional is empty.
func withEmptyValue[T any](o Optional[T]) T {
	// define an empty value of the given type to use as default
	var empty T

	return o.GetOr(empty)
}

// mapToKV converts a map to a list of key=value strings. This is a temporary workaround pending a map support in
// Dagger.
func mapToKV(m map[string]string) []string {
	var kv []string
	for k, v := range m {
		kv = append(kv, k+"="+v)
	}
	return kv
}
