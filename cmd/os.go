package cmd

import "github.com/spf13/cobra"

// newOSCmd represents the os command
func newOSCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "os",
		Short:        "run os gNOI RPCs",
		SilenceUsage: true,
	}
	gApp.InitOSFlags(cmd)
	return cmd
}

// newOSInstallCmd represents the os install command
func newOSInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "install",
		Short:        "run OS Install gNOI RPC",
		PreRunE:      gApp.PreRunEOSInstall,
		RunE:         gApp.RunEOSInstall,
		SilenceUsage: true,
	}
	gApp.InitOSInstallFlags(cmd)
	return cmd
}

// newOSInstallCmd represents the os activate command
func newOSActivateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "activate",
		Short:        "run OS Activate gNOI RPC",
		PreRunE:      gApp.PreRunEOSActivate,
		RunE:         gApp.RunEOSActivate,
		SilenceUsage: true,
	}
	gApp.InitOSActivateFlags(cmd)
	return cmd
}

// newOSVerifyCmd represents the os verify command
func newOSVerifyCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:          "verify",
		Short:        "run OS Verify gNOI RPC",
		PreRunE:      gApp.PreRunEOSVerify,
		RunE:         gApp.RunEOSVerify,
		SilenceUsage: true,
	}
	gApp.InitOSVerifyFlags(cmd)
	return cmd
}
