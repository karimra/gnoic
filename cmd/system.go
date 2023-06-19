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
	cmd.AddCommand(
		newSystemPingCmd(),
		newSystemTracerouteCmd(),
		newSystemTimeCmd(),
		newSystemSetPackageCmd(),
		newSystemSwitchControlProcessorCmd(),
		newSystemRebootCmd(),
		newSystemRebootStatusCmd(),
		newSystemCancelRebootCmd(),
		newSystemKillProcessCmd(),
	)
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
		PreRun: func(cmd *cobra.Command, _ []string) {
			gApp.Config.SetLocalFlagsFromFile(cmd)
		},
		PreRunE:      gApp.PreRunESystemTraceRoute,
		RunE:         gApp.RunESystemTraceRoute,
		SilenceUsage: true,
	}
	gApp.InitSystemTracerouteFlags(cmd)
	return cmd
}

// newSystemTimeCmd represents the system time command
func newSystemTimeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "time",
		Short: "run System Time gNOI RPC",
		PreRun: func(cmd *cobra.Command, _ []string) {
			gApp.Config.SetLocalFlagsFromFile(cmd)
		},
		RunE:         gApp.RunESystemTime,
		SilenceUsage: true,
	}
	gApp.InitSystemTimeFlags(cmd)
	return cmd
}

// newSystemSetPackageCmd represents the system set-package command
func newSystemSetPackageCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "set-package",
		Short: "run System SetPackage gNOI RPC",
		PreRun: func(cmd *cobra.Command, _ []string) {
			gApp.Config.SetLocalFlagsFromFile(cmd)
		},
		PreRunE:      gApp.PreRunESetPackage,
		RunE:         gApp.RunESystemSetPackage,
		SilenceUsage: true,
	}
	gApp.InitSystemSetPackageFlags(cmd)
	return cmd
}

// newSystemTracerouteCmd represents the system switch-control-processor command
func newSystemSwitchControlProcessorCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "switch-control-processor",
		Aliases: []string{"scp", "switch"},
		Short:   "run System SwitchControlProcessor gNOI RPC",
		PreRun: func(cmd *cobra.Command, _ []string) {
			gApp.Config.SetLocalFlagsFromFile(cmd)
		},
		RunE:         gApp.RunESystemSwitchControlProcessor,
		SilenceUsage: true,
	}
	gApp.InitSystemSwitchControlProcessorFlags(cmd)
	return cmd
}

// newSystemRebootCmd represents the system reboot command
func newSystemRebootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "reboot",
		Short:        "run System Reboot gNOI RPC",
		PreRunE:      gApp.PreRunESystemReboot,
		RunE:         gApp.RunESystemReboot,
		SilenceUsage: true,
	}
	gApp.InitSystemRebootFlags(cmd)
	return cmd
}

// newSystemRebootStatusCmd represents the system reboot-status command
func newSystemRebootStatusCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "reboot-status",
		Short: "run System RebootStatus gNOI RPC",
		PreRun: func(cmd *cobra.Command, _ []string) {
			gApp.Config.SetLocalFlagsFromFile(cmd)
		},
		RunE:         gApp.RunESystemRebootStatus,
		SilenceUsage: true,
	}
	gApp.InitSystemRebootStatusFlags(cmd)
	return cmd
}

// newSystemCancelRebootCmd represents the system cancel-reboot command
func newSystemCancelRebootCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cancel-reboot",
		Short: "run System CancelReboot gNOI RPC",
		PreRun: func(cmd *cobra.Command, _ []string) {
			gApp.Config.SetLocalFlagsFromFile(cmd)
		},
		RunE:         gApp.RunESystemCancelReboot,
		SilenceUsage: true,
	}
	gApp.InitSystemCancelRebootFlags(cmd)
	return cmd
}

// newSystemRebootCmd represents the system reboot command
func newSystemKillProcessCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "kill-process",
		Aliases:      []string{"k"},
		Short:        "run System KillProcess gNOI RPC",
		PreRunE:      gApp.PreRunESystemKillProcess,
		RunE:         gApp.RunESystemKillProcess,
		SilenceUsage: true,
	}
	gApp.InitSystemKillProcessFlags(cmd)
	return cmd
}
