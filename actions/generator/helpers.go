package main

import (
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// toTitleCase converts a string to title case.
func toTitleCase(str string) string {
	title := cases.Title(language.English)
	words := strings.FieldsFunc(str, func(r rune) bool { return r == '-' || r == '_' })
	for i := 0; i < len(words); i++ {
		words[i] = title.String(words[i])
	}
	return strings.Join(words, "")
}

// toCamelCase converts a string from snake_case or kebab-case to camelCase.
func toCamelCase(s string) string {
	// Split the string into words.
	words := strings.FieldsFunc(s, func(r rune) bool {
		return r == '-' || r == '_'
	})

	// Process each word.
	for i, word := range words {
		// Capitalize the first letter of each word except the first one.
		if i > 0 {
			words[i] = strings.Title(word)
		} else {
			words[i] = strings.ToLower(word)
		}
	}

	// Join the words back into a single string.
	return strings.Join(words, "")
}

// getSortedKeys returns a sorted list of keys from a map.
func getSortedKeys[T any](m map[string]T) []string {
	var keys []string

	for k := range m {
		keys = append(keys, k)
	}

	sort.Strings(keys)

	return keys
}
