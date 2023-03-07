package cmd

import "github.com/spf13/cobra"

// newServicesCmd represents the services command
func newServicesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "services",
		Short: "queries the services supported by the target gRPC server",

		RunE:         gApp.RunEServices,
		SilenceUsage: true,
	}
	// gApp.InitServerFlags(cmd)
	return cmd
}
