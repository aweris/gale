package main

import (
	"fmt"
	"strings"
)

type List struct {
	Workflows []Workflow
}

func (l *List) String() string {
	sb := &strings.Builder{}

	var (
		indentation = "  "
		newline     = "\n"
	)
	for _, workflow := range l.Workflows {

		sb.WriteString("- Workflow: ")
		if workflow.Name != "" {
			sb.WriteString(fmt.Sprintf("%s (path: %s)", workflow.Name, workflow.Path))
		} else {
			sb.WriteString(fmt.Sprintf("%s", workflow.Path))
		}
		sb.WriteString(newline)

		sb.WriteString(indentation)
		sb.WriteString("Jobs:")
		sb.WriteString(newline)

		for _, job := range workflow.Jobs {
			sb.WriteString(indentation)
			sb.WriteString(fmt.Sprintf("  - %s", job.JobID))
			sb.WriteString(newline)
		}

		sb.WriteString("\n") // extra empty line
	}

	return sb.String()
}

// Get returns a workflow.
func (l *List) Get(
	// workflow name or path
	workflow string,
) (*Workflow, error) {
	for _, wf := range l.Workflows {
		if wf.Name == workflow || wf.Path == workflow {
			return &wf, nil
		}
	}

	return nil, fmt.Errorf("workflow %s not found", workflow)
}
