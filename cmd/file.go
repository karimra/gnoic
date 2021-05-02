package cmd

import "github.com/spf13/cobra"

// newFileCmd represents the file command
func newFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "run file gNOI services",

		SilenceUsage: true,
	}
	gApp.InitFileFlags(cmd)
	return cmd
}
