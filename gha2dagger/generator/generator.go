package generator

import (
	"fmt"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"

	"github.com/dave/jennifer/jen"

	"github.com/aweris/gale/common/model"
)

// GenModule generates the run function for the custom action module.
func GenRunFn(module, caRepo, caRef string, caInputs map[string]model.CustomActionInput) *jen.Statement {
	params := make([]jen.Code, 0, len(caInputs))
	inputs := make([]jen.Code, 0, len(caInputs))

	for inputName, input := range caInputs {
		input := NewInput(inputName, input)

		params = append(params, input.AsFuncParam())
		inputs = append(inputs, input.AsActionRuntimeConfig())
	}

	params = append(params, getActionRuntimeDefaultRunParams()...)
	params = append(params, jen.Line())

	signature := jen.Params(jen.Id("m").Id(module)).Id("Run").Params(params...).Id("*Container")

	body := jen.Block(
		jen.Comment("initializing runtime options").Line().Id("opts").Op(":=").Add(getActionsRuntimeRunOpts()), // initializing runtime options
		jen.Line(), // empty line
		jen.Return(getActionRuntimeSync(caRepo, caRef, inputs)), // return dag.ActionsRuntime().Run("actions/hello-world-javascript-action@main", opts).Sync()
	)

	return jen.Func().Add(signature, body)
}

func getActionRuntimeDefaultRunParams() []jen.Code {
	return []jen.Code{
		jen.Line().Comment("The directory containing the repository source. If source is provided, rest of the options are ignored.").Line().Id("source").Id("Optional[*Directory]"),
		jen.Line().Comment("The name of the repository. Format: owner/name.").Line().Id("repo").Id("Optional[string]"),
		jen.Line().Comment("Tag name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.").Line().Id("tag").Id("Optional[string]"),
		jen.Line().Comment("Branch name to check out. Only one of branch or tag can be used. Precedence is as follows: tag, branch.").Line().Id("branch").Id("Optional[string]"),
		jen.Line().Comment("Image to use for the runner.").Line().Id("runnerImage").Id("Optional[string]"),
		jen.Line().Comment("Enables debug mode.").Line().Id("runnerDebug").Id("Optional[bool]"),
		jen.Line().Comment("GitHub token to use for authentication.").Line().Id("token").Id("Optional[*Secret]"),
	}
}

func getActionsRuntimeRunOpts() jen.Code {
	return jen.Id("ActionsRuntimeRunOpts").
		Values(
			jen.Dict{
				jen.Id("Source"):      jen.Id("source.GetOr(nil)"),
				jen.Id("Repo"):        jen.Id("repo.GetOr(\"\")"),
				jen.Id("Tag"):         jen.Id("tag.GetOr(\"\")"),
				jen.Id("Branch"):      jen.Id("branch.GetOr(\"\")"),
				jen.Id("RunnerImage"): jen.Id("runnerImage.GetOr(\"\")"),
				jen.Id("RunnerDebug"): jen.Id("runnerDebug.GetOr(false)"),
				jen.Id("Token"):       jen.Id("token.GetOr(nil)"),
			},
		)
}

func getActionRuntimeSync(caRepo string, caRef string, inputs []jen.Code) *jen.Statement {
	return jen.Id("dag").Dot("ActionsRuntime").Call().Dot(StartWithNewLine("Run")).
		Call(jen.Lit(fmt.Sprintf("%s@%s", caRepo, caRef)), jen.Id("opts")).
		Add(inputs...).
		Dot(StartWithNewLine("Sync")).
		Call()
}

// FormatModuleName formats a module name to a valid Dagger module name.
func FormatModuleName(name string) string {
	return toTitleCase(name)
}

// FormatActionInputName formats input name to generate `--with-input-name` format for cli
func FormatActionInputName(name string) string {
	return "with" + toTitleCase(name)
}

// StartWithNewLine prefix string with newline character to allow .Dot generations start in new line
func StartWithNewLine(s string) string {
	return "\n" + s
}

func toTitleCase(str string) string {
	title := cases.Title(language.English)
	words := strings.FieldsFunc(str, func(r rune) bool { return r == '-' || r == '_' })
	for i := 0; i < len(words); i++ {
		words[i] = title.String(words[i])
	}
	return strings.Join(words, "")
}
