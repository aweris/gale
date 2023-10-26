package main

import (
	"strings"

	"gopkg.in/yaml.v3"

	"actions-runtime/model"
)

// marshalWorkflowToFile marshals the given workflow as yaml and writes it to a file.
func marshalWorkflowToFile(workflow *model.Workflow) (*File, error) {
	data, err := yaml.Marshal(workflow)
	if err != nil {
		return nil, err
	}

	return dag.Directory().WithNewFile("workflow.yml", string(data)).File("workflow.yml"), nil
}

func parseKeyValues(kvs []string) map[string]string {
	result := make(map[string]string, len(kvs))

	for _, kv := range kvs {
		parts := strings.Split(kv, "=")
		result[parts[0]] = parts[1]
	}

	return result
}
