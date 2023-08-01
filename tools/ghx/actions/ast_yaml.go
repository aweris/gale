package actions

import (
	"fmt"
	"strconv"

	"gopkg.in/yaml.v3"
)

// yaml.Marshaller and yaml.Unmarshaler bindings for ast types.

var (
	_ yaml.Marshaler   = new(String)
	_ yaml.Unmarshaler = new(String)
)

const (
	yamlBoolTag  = "!!bool"
	yamlStrTag   = "!!str"
	yamlIntTag   = "!!int"
	yamlFloatTag = "!!float"
)

func (s *String) MarshalYAML() (interface{}, error) {
	if s.Quoted {
		return strconv.Quote(s.Value), nil
	}

	return s.Value, nil
}

func (s *String) UnmarshalYAML(n *yaml.Node) error {
	// Do not check n.Tag is !!str because we don't need to check the node is string strictly.
	// In almost all cases, other nodes (like 42) are handled as string with its string representation.
	if n.Kind != yaml.ScalarNode {
		return fmt.Errorf("expected scalar node but got %s node with %q tag", nodeKindName(n.Kind), n.Tag)
	}

	s.Value = n.Value
	s.Quoted = isQuotedString(n)

	return nil
}

var (
	_ yaml.Marshaler   = new(Bool)
	_ yaml.Unmarshaler = new(Bool)
)

func (b *Bool) MarshalYAML() (interface{}, error) {
	if b.Expression != nil {
		return b.Expression.Value, nil
	}

	return b.Value, nil
}

func (b *Bool) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.ScalarNode {
		return fmt.Errorf("expected scalar node but got %s node with %q tag", nodeKindName(n.Kind), n.Tag)
	}

	switch n.Tag {
	case yamlBoolTag:
		val, err := strconv.ParseBool(n.Value)
		if err != nil {
			return fmt.Errorf("failed to parse bool value: %w", err)
		}

		b.Value = val
	case yamlStrTag:
		b.Expression = &String{Value: n.Value, Quoted: isQuotedString(n)}
	default:
		return fmt.Errorf("expected !!bool or !!str tag but got %q", n.Tag)
	}

	return nil
}

var (
	_ yaml.Marshaler   = new(Int)
	_ yaml.Unmarshaler = new(Int)
)

func (i *Int) MarshalYAML() (interface{}, error) {
	if i.Expression != nil {
		return i.Expression.Value, nil
	}

	return i.Value, nil
}

func (i *Int) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.ScalarNode {
		return fmt.Errorf("expected scalar node but got %s node with %q tag", nodeKindName(n.Kind), n.Tag)
	}

	switch n.Tag {
	case yamlIntTag:
		val, err := strconv.ParseInt(n.Value, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse int value: %w", err)
		}

		i.Value = int(val)
	case yamlStrTag:
		i.Expression = &String{Value: n.Value, Quoted: isQuotedString(n)}
	default:
		return fmt.Errorf("expected !!int or !!str tag but got %q", n.Tag)
	}

	return nil
}

var (
	_ yaml.Marshaler   = new(Float)
	_ yaml.Unmarshaler = new(Float)
)

func (f *Float) MarshalYAML() (interface{}, error) {
	if f.Expression != nil {
		return f.Expression.Value, nil
	}

	return f.Value, nil
}

func (f *Float) UnmarshalYAML(n *yaml.Node) error {
	if n.Kind != yaml.ScalarNode {
		return fmt.Errorf("expected scalar node but got %s node with %q tag", nodeKindName(n.Kind), n.Tag)
	}

	switch n.Tag {
	case yamlFloatTag:
		val, err := strconv.ParseFloat(n.Value, 64)
		if err != nil {
			return fmt.Errorf("failed to parse float value: %w", err)
		}

		f.Value = val
	case yamlStrTag:
		f.Expression = &String{Value: n.Value, Quoted: isQuotedString(n)}
	default:
		return fmt.Errorf("expected !!float or !!str tag but got %q", n.Tag)
	}

	return nil
}

// nodeKindName returns a human-readable name of the yaml.Kind.
func nodeKindName(k yaml.Kind) string {
	switch k {
	case yaml.DocumentNode:
		return "document"
	case yaml.SequenceNode:
		return "sequence"
	case yaml.MappingNode:
		return "mapping"
	case yaml.ScalarNode:
		return "scalar"
	case yaml.AliasNode:
		return "alias"
	default:
		// all known kinds are handled above. This default case is unreachable.
		return "unknown"
	}
}

// isQuotedString returns true if the node is quoted string.
func isQuotedString(n *yaml.Node) bool {
	return n.Style&(yaml.DoubleQuotedStyle|yaml.SingleQuotedStyle) != 0
}
