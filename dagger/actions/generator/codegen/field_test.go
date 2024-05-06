package codegen

import (
	"testing"

	"github.com/dave/jennifer/jen"
	"github.com/stretchr/testify/assert"
)

func TestFieldBuilder_Build(t *testing.T) {
	tests := []struct {
		name string
		fb   *FieldBuilder
		want *jen.Statement
	}{
		{
			name: "empty",
			fb:   Field("test", "string"),
			want: jen.Id("test").Id("string"),
		},
		{
			name: "with description",
			fb:   Field("test", "string").WithDescription("test description"),
			want: jen.Comment("test description").Line().Id("test").Id("string"),
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			assert.Equal(t, tt.want, tt.fb.Build())
		})
	}
}
