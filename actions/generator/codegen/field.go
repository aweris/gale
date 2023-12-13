package codegen

import "github.com/dave/jennifer/jen"

// FieldBuilder is a builder for a struct field.
type FieldBuilder struct {
	Name        string
	Type        string
	Description string
}

// Field creates a new field builder.
func Field(name, typ string) *FieldBuilder {
	return &FieldBuilder{Name: name, Type: typ}
}

// WithDescription sets the description of the field.
func (fb *FieldBuilder) WithDescription(description string) *FieldBuilder {
	fb.Description = description
	return fb
}

// Build builds the field.
func (fb *FieldBuilder) Build() *jen.Statement {
	// Initialize a new statement
	s := &jen.Statement{}

	// Add a comment if provided
	if fb.Description != "" {
		s = s.Comment(fb.Description).Line()
	}

	// Add the field declaration
	return s.Id(fb.Name).Id(fb.Type)
}
