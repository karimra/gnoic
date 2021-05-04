package cmd

import "github.com/spf13/cobra"

// newVersionCmd represents the version command
func newVersionCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "version",
		Short:        "print gnoic version",
		RunE:         gApp.RunEVersion,
		SilenceUsage: true,
	}
	return cmd
}
