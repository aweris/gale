package gctx

import (
	"fmt"
	"reflect"

	"dagger.io/dagger"

	"github.com/caarlos0/env/v9"

	"github.com/aweris/gale/internal/dagger/helpers"
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

const trueStr = "true"

// WithContainerEnv loads context fields into the container as environment variables or secrets.
// It expects a struct with exported fields. Fields tagged with `container_env` or `container_secret` are loaded
// using the `env` tag value as their name.
func WithContainerEnv[T any](client *dagger.Client, t T) dagger.WithContainerFunc {
	return func(container *dagger.Container) *dagger.Container {
		val := reflect.ValueOf(&t).Elem()
		typ := val.Type()

		for i := 0; i < typ.NumField(); i++ {
			field := typ.Field(i)

			containerEnvTag := typ.Field(i).Tag.Get("container_env")
			containerSecretTag := typ.Field(i).Tag.Get("container_secret")

			// skip if the field is not tagged with container_env or container_secret
			if containerEnvTag == "" && containerSecretTag == "" {
				continue
			}

			if containerEnvTag == trueStr && containerSecretTag == trueStr {
				return helpers.FailPipeline(container, fmt.Errorf("field %s is tagged with both container_env and container_secret", field.Name))
			}

			envTag := field.Tag.Get("env")
			if envTag == "" {
				return helpers.FailPipeline(container, fmt.Errorf("field %s is tagged with container_env or container_secret but not tagged with env", field.Name))
			}

			// TODO: handle other types properly
			envVal := val.Field(i).Interface()

			if containerEnvTag == trueStr {
				container = container.WithEnvVariable(envTag, fmt.Sprintf("%v", envVal))
			}

			if containerSecretTag == trueStr {
				container = container.WithSecretVariable(envTag, client.SetSecret(envTag, fmt.Sprintf("%v", envVal)))
			}
		}

		return container
	}
}
