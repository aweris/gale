package codegen

import (
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/assert"
)

func TestFuncArgsBuilder_Build(t *testing.T) {
	tests := []struct {
		name string
		fb   *FuncArgsBuilder
		want *jen.Statement
	}{
		{
			name: "minimal",
			fb:   FuncArg("test", "string"),
			want: jen.Id("test").String(),
		},
		{
			name: "with description",
			fb:   FuncArg("test", "string").WithDescription("test description"),
			want: jen.Comment("test description").Line().Id("test").String(),
		},
		{
			name: "with optional",
			fb:   FuncArg("test", "string").WithOptional(true),
			want: jen.Comment("+optional=true").Line().Id("test").String(),
		},
		{
			name: "with default",
			fb:   FuncArg("test", "string").WithOptional(true).WithDefault("test"),
			want: jen.Comment("+optional=true").Line().Comment("+default=test").Line().Id("test").String(),
		},
		{
			name: "with description, optional and default",
			fb:   FuncArg("test", "string").WithDescription("test description").WithOptional(true).WithDefault("test"),
			want: jen.Comment("test description").Line().Comment("+optional=true").Line().Comment("+default=test").Line().Id("test").String(),
		},
		{
			name: "skip default if not optional",
			fb:   FuncArg("test", "string").WithDescription("test description").WithDefault("test"),
			want: jen.Comment("test description").Line().Id("test").String(),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.fb.Build())
		})
	}
}
