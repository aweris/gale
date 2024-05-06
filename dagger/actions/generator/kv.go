package main

import (
	"fmt"
	"strings"
)

// FIXME: This is a temporary solution until dagger supports map types as public types. Remove when
// issue: https://github.com/dagger/dagger/issues/6138 is fixed.

// KV is representing a key=value pair to be used in a list of key=value strings.
type KV struct {
	Key   string
	Value string
}

// NewKV returns a new KV instance.
func NewKV(key, value string) KV {
	return KV{Key: key, Value: value}
}

// ConvertMapToKVSlice converts a map to a list of key=value strings.
func ConvertMapToKVSlice(m map[string]string) []KV {
	// ensure keys are sorted to ensure consistent output
	keys := getSortedKeys(m)

	// initialize slice with capacity to avoid reallocation
	kv := make([]KV, 0, len(m))

	// use sorted keys to convert map to slice
	for _, k := range keys {
		kv = append(kv, NewKV(k, m[k]))
	}

	return kv
}

// ConvertKVSliceToMap converts a list of KV to a map.
func ConvertKVSliceToMap(kv []KV) map[string]string {
	m := make(map[string]string, len(kv))
	for _, v := range kv {
		m[v.Key] = v.Value
	}
	return m
}

// ParseKeyValuePairs converts a list of key=value strings to a list of KV.
func ParseKeyValuePairs(s []string) ([]KV, error) {
	slice := make([]KV, 0, len(s))
	for _, v := range s {
		kv, err := DecodeKeyValue(v)
		if err != nil {
			return nil, err
		}
		slice = append(slice, kv)
	}
	return slice, nil
}

// DecodeKeyValue converts a key=value string to a KV.
func DecodeKeyValue(str string) (KV, error) {
	parts := strings.SplitN(str, "=", 2)
	if len(parts) != 2 {
		return KV{}, fmt.Errorf("invalid key=value pair: %s", str)
	}
	return NewKV(parts[0], parts[1]), nil
}
