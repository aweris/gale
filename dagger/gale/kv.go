package main

// FIXME: This is a temporary solution until dagger supports map types as public types. Remove when
// issue: https://github.com/dagger/dagger/issues/6138 is fixed.

type KV []string

func NewKVFromMap(m map[string]string) KV {
	var kv KV
	for k, v := range m {
		kv = append(kv, k, v)
	}
	return kv
}

func (kv KV) ToMap() map[string]string {
	m := make(map[string]string)
	for i := 0; i < len(kv); i += 2 {
		m[kv[i]] = kv[i+1]
	}
	return m
}
