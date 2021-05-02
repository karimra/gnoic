package cmd

import "github.com/spf13/cobra"

// newSystemCmd represents the system command
func newSystemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "system",
		Short: "run system gNOI services",

		SilenceUsage: true,
	}
	gApp.InitSystemFlags(cmd)
	return cmd
}
