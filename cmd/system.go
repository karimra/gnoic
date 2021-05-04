package cmd

import "github.com/spf13/cobra"

// newSystemCmd represents the system command
func newSystemCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "system",
		Aliases: []string{"sys"},
		Short:   "run System gNOI RPCs",

		SilenceUsage: true,
	}
	gApp.InitSystemFlags(cmd)
	return cmd
}

// newSystemPingCmd represents the system ping command
func newSystemPingCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "ping",
		Short:        "run System Ping gNOI RPC",
		PreRunE:      gApp.PreRunESystemPing,
		RunE:         gApp.RunESystemPing,
		SilenceUsage: true,
	}
	gApp.InitSystemPingFlags(cmd)
	return cmd
}

// newSystemTracerouteCmd represents the system traceroute command
func newSystemTracerouteCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "traceroute",
		Short: "run System Traceroute gNOI RPC",
		//PreRunE:      gApp.PreRunESystemPing,
		RunE:         gApp.RunESystemTraceRoute,
		SilenceUsage: true,
	}
	gApp.InitSystemTracerouteFlags(cmd)
	return cmd
}
