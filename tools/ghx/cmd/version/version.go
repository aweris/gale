package version

import (
	"encoding/json"
	"fmt"

	"github.com/spf13/cobra"

	"github.com/aweris/gale/internal/version"
)

// NewCommand  creates a new root command.
func NewCommand() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "version",
		Short: "Print version information",
		Long:  "Print version information",
		RunE: func(cmd *cobra.Command, args []string) error {
			v := version.GetVersion()

			marshalled, err := json.MarshalIndent(&v, "", "  ")
			if err != nil {
				return err
			}

			fmt.Println(string(marshalled))

			return nil
		},
	}

	return cmd
}
