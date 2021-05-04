package cmd

import (
	"github.com/spf13/cobra"
)

// newTreeCmd represents the tree command
func newTreeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "tree",
		Short:        "print the commands tree",
		RunE:         gApp.RunETree,
		SilenceUsage: true,
	}
	gApp.InitTreeFlags(cmd)
	return cmd
}
