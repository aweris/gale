package context

import (
	"os"
	"reflect"
)

// syncWithEnvValues syncs the env values with the given struct recursively. If the env value is not set or different
// from the struct value, the env value will be set to the struct value.
//
// The method uses `env` tag to get the env variable name. if env tag is not set, the field will be skipped.
func syncWithEnvValues(t interface{}) error {
	val := reflect.ValueOf(t).Elem()
	typ := val.Type()

	for i := 0; i < typ.NumField(); i++ {
		field := typ.Field(i)

		var (
			fieldVal = val.Field(i)
			envTag   = field.Tag.Get("env")
		)

		// if evn tag is not set and the field is not a struct, skip
		if envTag == "" && fieldVal.Kind() != reflect.Struct {
			continue
		}

		if fieldVal.Kind() == reflect.Struct {
			if err := syncWithEnvValues(fieldVal.Addr().Interface()); err != nil {
				return err
			}

			continue
		}

		current, ok := os.LookupEnv(envTag)

		// env value is same as current value, skip
		if ok && reflect.DeepEqual(current, fieldVal.Interface()) {
			continue
		}

		// set the env value
		if err := os.Setenv(envTag, fieldVal.String()); err != nil {
			return err
		}
	}

	return nil
}
