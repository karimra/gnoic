package cmd

import "github.com/spf13/cobra"

// newHealthzCmd represents the healthz command
func newHealthzCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "healthz",
		Short:        "run gNOI healthz RPCs",
		SilenceUsage: true,
	}
	gApp.InitHealthzFlags(cmd)
	return cmd
}

// newHealthzGetmd represents the healthz get command
func newHealthzGetmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "run gNOI healthz Get RPC",
		PreRunE: func(cmd *cobra.Command, args []string) error {
			gApp.Config.SetLocalFlagsFromFile(cmd)
			return nil
		},
		RunE:         gApp.RunEHealthzGet,
		SilenceUsage: true,
	}
	gApp.InitHealthzGetFlags(cmd)
	return cmd
}
