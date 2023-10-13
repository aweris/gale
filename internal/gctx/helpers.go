package gctx

import (
	"github.com/caarlos0/env/v9"
)

// NewContextFromEnv initializes a new context from environment variables. It works with structs having exported fields
// tagged with the `env` tag.
func NewContextFromEnv[T any]() (T, error) {
	val := new(T)

	if err := env.Parse(val); err != nil {
		return *val, err
	}

	return *val, nil
}
