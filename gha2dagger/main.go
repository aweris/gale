package main

import (
	"fmt"
	"os"
)

func main() {
	if len(os.Args) < 3 {
		fmt.Fprintf(os.Stderr, "usage: %s <action>@<version> <destination>\n", os.Args[0])
		os.Exit(1)
	}

	var (
		source = os.Args[1]
		dest   = os.Args[2]
	)

	// get custom action info
	action, err := NewCustomAction(source)
	if err != nil {
		fmt.Fprintf(os.Stderr, "failed to get custom action %s: %v\n", source, err)
		os.Exit(1)
	}

	if err := generate(dest, action); err != nil {
		fmt.Fprintf(os.Stderr, "failed to generate custom action %s: %v\n", source, err)
		os.Exit(1)
	}
}
