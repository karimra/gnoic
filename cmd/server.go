package cmd

import (
	"github.com/spf13/cobra"
)

// newServerCmd represents the server command
func newServerCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "server",
		Short: "starts a gNOI server",
		PreRunE: func(_ *cobra.Command, _ []string) error {
			if len(gApp.Config.Address) == 0 {
				gApp.Config.Address = []string{":9339"}
			}
			return nil
		},
		RunE:         gApp.RunEServer,
		SilenceUsage: true,
	}
	gApp.InitServerFlags(cmd)
	return cmd
}
