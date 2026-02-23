package cmd

import (
	"fmt"

	"github.com/spf13/cobra"
)

// version is set at build time via -ldflags.
var version = "dev"

func newVersionCmd() *cobra.Command {
	return &cobra.Command{
		Use:   "version",
		Short: "Print the CLI version",
		Args:  cobra.NoArgs,
		Run: func(cmd *cobra.Command, args []string) {
			fmt.Fprintln(cmd.OutOrStdout(), "mvdata version "+version)
		},
	}
}
