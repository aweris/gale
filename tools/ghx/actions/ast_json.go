package actions

import (
	"encoding/json"
	"fmt"
	"strconv"
)

// json.Marshaller and json.Unmarshaler bindings for ast types.

var (
	_ json.Marshaler   = new(String)
	_ json.Unmarshaler = new(String)
)

func (s *String) MarshalJSON() ([]byte, error) {
	return json.Marshal(s.Value)
}

func (s *String) UnmarshalJSON(data []byte) error {
	s.Value = string(data)
	s.Quoted = false

	return nil
}

var (
	_ json.Marshaler   = new(Bool)
	_ json.Unmarshaler = new(Bool)
)

func (b *Bool) MarshalJSON() ([]byte, error) {
	if b.Expression != nil {
		return json.Marshal(b.Expression.Value)
	}

	return json.Marshal(b.Value)
}

func (b *Bool) UnmarshalJSON(data []byte) error {
	var val interface{}

	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}

	switch v := val.(type) {
	case bool:
		b.Value = v
	case string:
		bv, err := strconv.ParseBool(v)
		// if err is not nil, val is an expression.
		if err != nil {
			b.Expression = &String{Value: v, Quoted: false}
		}

		// even this is an expression, val is false, and we can ignore expression and set val here.
		b.Value = bv
	default:
		return fmt.Errorf("invalid value for bool: %v", val)
	}

	return nil
}

var (
	_ json.Marshaler   = new(Int)
	_ json.Unmarshaler = new(Int)
)

func (i *Int) MarshalJSON() ([]byte, error) {
	if i.Expression != nil {
		return json.Marshal(i.Expression.Value)
	}

	return json.Marshal(i.Value)
}

func (i *Int) UnmarshalJSON(data []byte) error {
	var val interface{}

	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}

	switch v := val.(type) {
	case float64:
		i.Value = int(v)
	case string:
		i.Expression = &String{Value: v, Quoted: false}
	default:
		return fmt.Errorf("invalid value for int: %v", val)
	}

	return nil
}

var (
	_ json.Marshaler   = new(Float)
	_ json.Unmarshaler = new(Float)
)

func (f *Float) MarshalJSON() ([]byte, error) {
	if f.Expression != nil {
		return json.Marshal(f.Expression.Value)
	}

	return json.Marshal(f.Value)
}

func (f *Float) UnmarshalJSON(data []byte) error {
	var val interface{}

	if err := json.Unmarshal(data, &val); err != nil {
		return err
	}

	switch v := val.(type) {
	case float64:
		f.Value = v
	case string:
		f.Expression = &String{Value: v, Quoted: false}
	default:
		return fmt.Errorf("invalid value for float: %v", val)
	}

	return nil
}
