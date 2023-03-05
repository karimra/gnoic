package cmd

import (
	"github.com/spf13/cobra"
)

// upgradeCmd represents the version upgrade command
func newVersionUpgradeCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "upgrade",
		Aliases: []string{"up"},
		Short:   "upgrade gnoic to the latest available version",
		PreRun: func(cmd *cobra.Command, _ []string) {
			gApp.Config.SetLocalFlagsFromFile(cmd)
		},
		RunE: gApp.VersionUpgradeRun,
	}
	gApp.InitVersionUpgradeFlags(cmd)
	return cmd
}
