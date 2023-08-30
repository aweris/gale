package core

import (
	"fmt"
	"strings"

	"github.com/cli/go-gh/v2"
)

// GetToken returns the auth token gh is configured to use
func GetToken() (string, error) {
	stdout, stderr, err := gh.Exec("auth", "token")
	if err != nil {
		return "", fmt.Errorf("failed to get token: %w stderr: %s", err, stderr.String())
	}

	// sanitize token by trimming spaces and newlines
	token := strings.TrimSpace(strings.TrimSuffix(stdout.String(), "\n"))

	return token, nil
}
