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
	cmd.AddCommand(
		newHealthzGetCmd(),
		newHealthzListCmd(),
		newHealthzAckCmd(),
		newHealthzArtifactCmd(),
		newHealthzCheckCmd(),
	)

	return cmd
}

// newHealthzGetCmd represents the healthz get command
func newHealthzGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "run gNOI healthz Get RPC",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			gApp.Config.SetLocalFlagsFromFile(cmd)
			return nil
		},
		RunE:         gApp.RunEHealthzGet,
		SilenceUsage: true,
	}
	gApp.InitHealthzGetFlags(cmd)
	return cmd
}

// newHealthzListcmd represents the healthz list command
func newHealthzListCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "list",
		Short: "run gNOI healthz List RPC",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			gApp.Config.SetLocalFlagsFromFile(cmd)
			return nil
		},
		RunE:         gApp.RunEHealthzList,
		SilenceUsage: true,
	}
	gApp.InitHealthzListFlags(cmd)
	return cmd
}

// newHealthzAckCmd represents the healthz ack command
func newHealthzAckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "ack",
		Short: "run gNOI healthz Acknowledge RPC",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			gApp.Config.SetLocalFlagsFromFile(cmd)
			return nil
		},
		RunE:         gApp.RunEHealthzAck,
		SilenceUsage: true,
	}
	gApp.InitHealthzAckFlags(cmd)
	return cmd
}

// newHealthzArtifactCmd represents the healthz artifact command
func newHealthzArtifactCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "artifact",
		Aliases: []string{"a"},
		Short:   "run gNOI healthz Artifact RPC",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			gApp.Config.SetLocalFlagsFromFile(cmd)
			return nil
		},
		RunE:         gApp.RunEHealthzArtifact,
		SilenceUsage: true,
	}
	gApp.InitHealthzArtifactFlags(cmd)
	return cmd
}

// newHealthzCheckCmd represents the healthz check command
func newHealthzCheckCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "check",
		Short: "run gNOI healthz Check RPC",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			gApp.Config.SetLocalFlagsFromFile(cmd)
			return nil
		},
		RunE:         gApp.RunEHealthzCheck,
		SilenceUsage: true,
	}
	gApp.InitHealthzCheckFlags(cmd)
	return cmd
}
