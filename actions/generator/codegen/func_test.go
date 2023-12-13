package codegen

import (
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/assert"
)

func TestFuncBuilder_Build(t *testing.T) {
	tests := []struct {
		name string
		fb   *FuncBuilder
		want string
	}{
		{
			name: "empty",
			fb:   Func("test", "string").WithBody(jen.Return(jen.Lit("test"))),
			want: "func test() string {\n\treturn \"test\"\n}",
		},
		{
			name: "with description",
			fb:   Func("test", "string").WithDescription("test description").WithBody(jen.Return(jen.Lit("test"))),
			want: "// test description\nfunc test() string {\n\treturn \"test\"\n}",
		},
		{
			name: "with single arg",
			fb:   Func("test", "string").WithArgs(FuncArg("test", "string")).WithBody(jen.Return(jen.Lit("test"))),
			want: "func test(\n\ttest string,\n) string {\n\treturn \"test\"\n}",
		},
		{
			name: "with single optional arg with default and description",
			fb:   Func("test", "string").WithArgs(FuncArg("test", "string").WithOptional(true).WithDefault("test").WithDescription("test description")).WithBody(jen.Return(jen.Lit("test"))),
			want: "func test(\n\t// test description\n\t// +optional=true\n\t// +default=test\n\ttest string,\n) string {\n\treturn \"test\"\n}",
		},
		{
			name: "with receiver",
			fb:   Func("test", "string").WithReceiver(&Receiver{Name: "t", Type: "test", UsePointer: true}).WithBody(jen.Return(jen.Lit("test"))),
			want: "func (t *test) test() string {\n\treturn \"test\"\n}",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.fb.Build().GoString())
		})
	}
}
