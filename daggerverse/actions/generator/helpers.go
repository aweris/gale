package main

import (
	"sort"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// formatModuleName formats a module name to a valid Dagger module name.
func formatModuleName(name string) string {
	return toTitleCase(name)
}

// formatActionInputName formats input name to generate `--with-input-name` format for cli
func formatActionInputName(name string) string {
	return "with" + toTitleCase(name)
}

// startWithNewLine prefix string with newline character to allow .Dot generations start in new line
func startWithNewLine(s string) string {
	return "\n" + s
}

// toTitleCase converts a string to title case.
func toTitleCase(str string) string {
	title := cases.Title(language.English)
	words := strings.FieldsFunc(str, func(r rune) bool { return r == '-' || r == '_' })
	for i := 0; i < len(words); i++ {
		words[i] = title.String(words[i])
	}
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
