package codegen

import (
	"fmt"

	"github.com/dave/jennifer/jen"
)

// FuncArgsBuilder builds function arguments with specific formatting.
type FuncArgsBuilder struct {
	Name        string // Name of the function argument.
	Type        string // Type of the function argument.
	Description string // Description of the function argument.
	Optional    bool   // Whether the argument is optional.
	Default     string // Default value of the argument.
}

// FuncArg creates a new function argument builder.
func FuncArg(name, typ string) *FuncArgsBuilder {
	return &FuncArgsBuilder{Name: name, Type: typ}
}

// WithDescription sets the description of the function argument.
func (fab *FuncArgsBuilder) WithDescription(description string) *FuncArgsBuilder {
	fab.Description = description
	return fab
}

// WithOptional sets whether the argument is optional.
func (fab *FuncArgsBuilder) WithOptional(optional bool) *FuncArgsBuilder {
	fab.Optional = optional
	return fab
}

// WithDefault sets the default value of the argument, applicable if optional is true.
func (fab *FuncArgsBuilder) WithDefault(defaultValue string) *FuncArgsBuilder {
	fab.Default = defaultValue
	return fab
}

// Build builds the function argument.
func (fab *FuncArgsBuilder) Build() *jen.Statement {
	// initialize a new statement
	s := &jen.Statement{}

	// Add description as a comment
	if fab.Description != "" {
		s = jen.Comment(fab.Description).Line()
	}

	// Optional and Default comments
	if fab.Optional {
		s.Comment(fmt.Sprintf("+optional=%t", fab.Optional)).Line()
		if fab.Default != "" {
			s.Comment(fmt.Sprintf("+default=%s", fab.Default)).Line()
		}
	}

	// Argument declaration
	return s.Id(fab.Name).Id(fab.Type)
}
