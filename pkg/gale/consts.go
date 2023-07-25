package gale

import "errors"

var (
	ErrWorkflowNotFound = errors.New("workflow not found")
	ErrJobNotFound      = errors.New("job not found")
)
