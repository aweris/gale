package generator

import (
	"github.com/dave/jennifer/jen"

	"github.com/aweris/gale/common/model"
)

type Input struct {
	Name  string
	Model model.CustomActionInput
}

func NewInput(name string, model model.CustomActionInput) *Input {
	return &Input{Name: name, Model: model}
}

func (i *Input) AsFuncParam() *jen.Statement {
	return jen.Line().Comment(i.Model.Description).Line().Id(FormatActionInputName(i.Name)).Id(i.getParamType())
}

func (i *Input) AsActionRuntimeConfig() *jen.Statement {
	return jen.Dot(StartWithNewLine("WithInput")).Call(jen.Lit(i.Name), i.getCallType())
}

func (i *Input) getParamName() string {
	return FormatActionInputName(i.Name)
}

func (i *Input) getParamType() string {
	if !i.Model.Required || i.Model.Default != "" {
		return "Optional[string]"
	}

	return "string"
}

func (i *Input) getCallType() *jen.Statement {
	if !i.Model.Required || i.Model.Default != "" {
		return jen.Id(i.getParamName()).Dot("GetOr").Call(jen.Lit(i.Model.Default))
	}

	return jen.Id(i.getParamName())
}
