package codegen

import (
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/assert"
)

func TestStructBuilder_Build(t *testing.T) {
	tests := []struct {
		name string
		sb   *StructBuilder
		want string
	}{
		{
			name: "empty",
			sb:   Struct("test"),
			want: "type test struct{} \n",
		},
		{
			name: "with description",
			sb:   Struct("test").WithDescription("test description"),
			want: "// test description\ntype test struct{} \n",
		},
		{
			name: "with fields",
			sb:   Struct("test").WithFields(Field("test", "string")),
			want: "type test struct {\n\ttest string\n} \n",
		},
		{
			name: "with funcs",
			sb:   Struct("test").WithFuncs(Func("test", "string").WithDescription("test description").WithBody(jen.Return(jen.Lit("test")))),
			want: "type test struct{}\n\n// test description\nfunc (r test) test() string {\n\treturn \"test\"\n} \n",
		},
		{
			name: "with fields and funcs",
			sb:   Struct("test").WithFields(Field("test", "string")).WithFuncs(Func("test", "string").WithDescription("test description").WithBody(jen.Return(jen.Lit("test")))),
			want: "type test struct {\n\ttest string\n}\n\n// test description\nfunc (r test) test() string {\n\treturn \"test\"\n} \n",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.sb.Build().GoString())
		})
	}
}
