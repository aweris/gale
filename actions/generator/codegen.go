package main

import (
	"github.com/dave/jennifer/jen"

	"actions-generator/codegen"
)

func generate(ca *CustomAction) *File {
	module := codegen.Module(toTitleCase(ca.RepoName)).
		WithFuncs(
			toRunFunc(ca),
			toGaleOptsFunc(),
			toGaleActionRunOptsFunc(),
			toGaleActionOptsFunc(ca.Meta.Inputs),
		)

	src := module.Build().GoString()

	return dag.Directory().WithNewFile("main.go", src).File("main.go")
}

func toRunFunc(ca *CustomAction) *codegen.FuncBuilder {
	args := []*codegen.FuncArgsBuilder{
		codegen.FuncArg("source", "*Directory").WithDescription("Directory containing the repository source. Takes precedence over `--repo`.\"").WithOptional(true),
		codegen.FuncArg("repo", "string").WithDescription("Repository name, format: owner/name.").WithOptional(true),
		codegen.FuncArg("tag", "string").WithDescription("Tag name to check out. Only works with `--repo`. Takes precedence over `--branch`.").WithOptional(true),
		codegen.FuncArg("branch", "string").WithDescription("Branch name to check out. Only works with `--repo`.").WithOptional(true),
		codegen.FuncArg("container", "*Container").WithDescription("Container to use for the runner.").WithOptional(true),
		codegen.FuncArg("runnerDebug", "bool").WithDescription("Enables debug mode.").WithOptional(true).WithDefault("false"),
	}

	call := make([]jen.Code, 0, len(ca.Meta.Inputs))

	for _, input := range ca.Meta.Inputs {
		arg := codegen.FuncArg(toCamelCase(input.Name), "string").
			WithDescription(input.Description).
			WithOptional(!input.Required || input.Default != "").
			WithDefault(input.Default)
		args = append(args, arg)
		call = append(call, jen.Id(toCamelCase(input.Name)))
	}

	return codegen.Func("Run", "*Container").
		WithDescription("Runs the " + ca.Repo + " GitHub Action.").
		WithArgs(args...).
		WithReceiver(&codegen.Receiver{Name: "_", Type: toTitleCase(ca.RepoName), UsePointer: true}).
		WithBody(
			jen.Return(
				jen.Id("dag").
					Dot("Gale").
					Call(jen.Id("toGaleOpts").Call(jen.Id("source"), jen.Id("repo"), jen.Id("tag"), jen.Id("branch"))).
					Dot("\n").Id("Action").
					Call(jen.Lit(ca.Repo+"@"+ca.Ref), jen.Id("toGaleActionOpts").Call(call...)).
					Dot("\n").Id("Run").
					Call(jen.Id("toGaleActionRunOpts").Call(jen.Id("container"), jen.Id("runnerDebug"))).
					Dot("\n").Id("Sync").
					Call(),
			),
		)
}
func toGaleOptsFunc() *codegen.FuncBuilder {
	return codegen.Func("toGaleOpts", "GaleOpts").
		WithDescription("Converts the custom action inputs to Gale options.").
		WithArgs(
			codegen.FuncArg("source", "*Directory"),
			codegen.FuncArg("repo", "string"),
			codegen.FuncArg("tag", "string"),
			codegen.FuncArg("branch", "string"),
		).
		WithBody(
			jen.Return(
				jen.Id("GaleOpts").Values(
					jen.Dict{
						jen.Id("Source"): jen.Id("source"),
						jen.Id("Repo"):   jen.Id("repo"),
						jen.Id("Tag"):    jen.Id("tag"),
						jen.Id("Branch"): jen.Id("branch"),
					},
				),
			),
		)
}

func toGaleActionRunOptsFunc() *codegen.FuncBuilder {
	return codegen.Func("toGaleActionRunOpts", "GaleActionsRunOpts").
		WithDescription("Converts the custom action inputs to Gale action run options.").
		WithArgs(
			codegen.FuncArg("container", "*Container"),
			codegen.FuncArg("debug", "bool"),
		).
		WithBody(
			jen.Return(
				jen.Id("GaleActionsRunOpts").Values(
					jen.Dict{
						jen.Id("Container"):   jen.Id("container"),
						jen.Id("RunnerDebug"): jen.Id("debug"),
					},
				),
			),
		)
}

func toGaleActionOptsFunc(inputs []CustomActionInput) *codegen.FuncBuilder {
	args := make([]*codegen.FuncArgsBuilder, 0, len(inputs))
	items := make([]jen.Code, 0, len(inputs))

	for i, input := range inputs {
		args = append(args, codegen.FuncArg(toCamelCase(input.Name), "string"))

		item := jen.Line().Qual("fmt", "Sprintf").Call(jen.Lit(input.Name+"=%s"), jen.Id(toCamelCase(input.Name)))

		if i == len(inputs)-1 {
			item.Add(jen.Id(",").Line())
		}

		items = append(items, item)
	}

	return codegen.Func("toGaleActionOpts", "GaleActionOpts").
		WithDescription("Converts the custom action inputs to Gale action options.").
		WithArgs(args...).
		WithBody(
			jen.Return(
				jen.Id("GaleActionOpts").Values(
					jen.Line().
						Add(jen.Dict{jen.Id("With"): jen.Index().String().Values(items...)}).
						Id(",").
						Line(),
				),
			),
		)
}
