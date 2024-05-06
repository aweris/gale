package codegen

import "github.com/dave/jennifer/jen"

type FuncBuilder struct {
	Name        string             // Name of the function.
	Description string             // Description of the function.
	Return      string             // Return type of the function.
	Args        []*FuncArgsBuilder // Arguments of the function.
	Receiver    *Receiver          // Receiver of the function.
	Body        []jen.Code         // Body of the function.
}

// Receiver represents a function receiver.
type Receiver struct {
	Name       string // Name of the receiver variable.
	Type       string // Type of the receiver variable.
	UsePointer bool   // Whether to use a pointer receiver.
}

func (r Receiver) Build() *jen.Statement {
	s := jen.Id(r.Name)

	if r.UsePointer {
		s = s.Op("*")
	}

	return s.Id(r.Type)
}

func Func(name, returnType string) *FuncBuilder {
	return &FuncBuilder{Name: name, Return: returnType}
}

func (fb *FuncBuilder) WithDescription(description string) *FuncBuilder {
	fb.Description = description
	return fb
}

func (fb *FuncBuilder) WithArgs(args ...*FuncArgsBuilder) *FuncBuilder {
	fb.Args = args
	return fb
}

func (fb *FuncBuilder) WithReceiver(receiver *Receiver) *FuncBuilder {
	fb.Receiver = receiver
	return fb
}

func (fb *FuncBuilder) WithBody(body ...jen.Code) *FuncBuilder {
	fb.Body = body
	return fb
}

func (fb *FuncBuilder) Build() *jen.Statement {
	// initialize a new statement
	s := &jen.Statement{}

	// Add a comment if provided
	if fb.Description != "" {
		s = s.Comment(fb.Description).Line()
	}

	// Add the function declaration
	s = s.Func()

	// Add the receiver if provided
	if fb.Receiver != nil {
		s = s.Params(fb.Receiver.Build())
	}

	// Add name
	s = s.Id(fb.Name)

	// Add the args
	var args []jen.Code

	for idx, a := range fb.Args {
		arg := jen.Line().Add(a.Build())

		if idx == len(fb.Args)-1 {
			arg = arg.Id(",").Line()
		}

		args = append(args, arg)
	}

	s = s.Params(args...)

	// Add the return type and return statement
	s = s.Id(fb.Return)

	// Add function body and return statement
	return s.Block(fb.Body...)
}
