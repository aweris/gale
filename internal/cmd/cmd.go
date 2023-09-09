package cmd

import (
	"log"
	"os"

	"github.com/spf13/pflag"
)

// BindEnv binds the value of a flag to an environment variable.
func BindEnv(fn *pflag.Flag, env string) {
	if fn == nil || fn.Changed {
		return
	}

	val := os.Getenv(env)

	if len(val) > 0 {
		if err := fn.Value.Set(val); err != nil {
			log.Fatalf("failed to bind env: %v\n", err)
		}
	}
}
