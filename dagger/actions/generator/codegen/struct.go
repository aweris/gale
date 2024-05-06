package codegen

import "github.com/dave/jennifer/jen"

type StructBuilder struct {
	Name        string
	Description string
	Fields      []*FieldBuilder
	Funcs       []*FuncBuilder
}

func Struct(name string) *StructBuilder {
	return &StructBuilder{Name: name}
}

func (sb *StructBuilder) WithDescription(description string) *StructBuilder {
	sb.Description = description
	return sb
}

func (sb *StructBuilder) WithFields(fields ...*FieldBuilder) *StructBuilder {
	sb.Fields = fields
	return sb
}

func (sb *StructBuilder) WithFuncs(funcs ...*FuncBuilder) *StructBuilder {
	sb.Funcs = funcs
	return sb
}

func (sb *StructBuilder) Build() *jen.Statement {
	// Initialize a new statement
	s := &jen.Statement{}

	// Add a comment if provided
	if sb.Description != "" {
		s = s.Comment(sb.Description).Line()
	}

	// Add the fields
	var fields []jen.Code

	for _, f := range sb.Fields {
		fields = append(fields, f.Build())
	}

	// Add the functions
	var funcs []jen.Code

	for _, f := range sb.Funcs {
		funcs = append(funcs, f.WithReceiver(&Receiver{Name: "r", Type: sb.Name}).Build().Line())
	}

	// Add the struct declaration
	return s.Type().Id(sb.Name).Struct(fields...).Line().Add(funcs...)
}
