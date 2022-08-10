package cmd

import "github.com/spf13/cobra"

// newFactoryResetCmd represents the factory-reset command
func newFactoryResetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "factory-reset",
		Aliases:      []string{"fr", "reset"},
		Short:        "run gNOI factory-reset RPCs",
		SilenceUsage: true,
	}
	cmd.AddCommand(newFactoryResetStartCmd())
	return cmd
}

// newFactoryResetStartCmd represents the factory-reset start command
func newFactoryResetStartCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "start",
		Short: "run gNOI factory-reset Start RPC",
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			gApp.Config.SetLocalFlagsFromFile(cmd)
			return nil
		},
		RunE:         gApp.RunEFactoryResetStart,
		SilenceUsage: true,
	}
	gApp.InitFactoryResetStartFlags(cmd)
	return cmd
}
