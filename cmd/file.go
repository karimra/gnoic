package cmd

import (
	"errors"

	"github.com/spf13/cobra"
)

// newFileCmd represents the file command
func newFileCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "file",
		Short: "run File gNOI RPCs",

		SilenceUsage: true,
	}
	gApp.InitFileFlags(cmd)
	cmd.AddCommand(
		newFileGetCmd(),
		newFileTransferCmd(),
		newFilePutCmd(),
		newFileStatCmd(),
		newFileRemoveCmd(),
	)
	return cmd
}

// newFileGetCmd represents the file get command
func newFileGetCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "get",
		Short: "run file Get gNOI RPC",
		PreRun: func(cmd *cobra.Command, _ []string) {
			gApp.Config.SetLocalFlagsFromFile(cmd)
		},
		RunE:         gApp.RunEFileGet,
		SilenceUsage: true,
	}
	gApp.InitFileGetFlags(cmd)
	return cmd
}

// newFileTransferCmd represents the file transfer command
func newFileTransferCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "transfer",
		Short:   "run file Transfer gNOI RPC",
		Aliases: []string{"trans", "tr"},
		PreRunE: func(cmd *cobra.Command, _ []string) error {
			gApp.Config.SetLocalFlagsFromFile(cmd)
			if gApp.Config.FileTransferLocal == "" {
				return errors.New("missing local file path, set with --local")
			}
			if gApp.Config.FileTransferRemote == "" {
				return errors.New("missing remote file path, set with --remote")
			}
			return nil
		},
		RunE:         gApp.RunEFileTransfer,
		SilenceUsage: true,
	}
	gApp.InitFileTransferFlags(cmd)
	return cmd
}

// newFilePutCmd represents the file put command
func newFilePutCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "put",
		Short:   "run file Put gNOI RPC",
		PreRunE: gApp.PreRunEFilePut,

		RunE:         gApp.RunEFilePut,
		SilenceUsage: true,
	}
	gApp.InitFilePutFlags(cmd)
	return cmd
}

// newFileStatCmd represents the file
func newFileStatCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "stat",
		Aliases: []string{"st"},
		Short:   "run file Stat gNOI RPC",
		PreRun: func(cmd *cobra.Command, _ []string) {
			gApp.Config.SetLocalFlagsFromFile(cmd)
		},
		RunE:         gApp.RunEFileStat,
		SilenceUsage: true,
	}
	gApp.InitFileStatFlags(cmd)
	return cmd
}

// newFileRemoveCmd represents the file remove command
func newFileRemoveCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "remove",
		Aliases: []string{"rm"},
		Short:   "run file Remove gNOI RPC",
		PreRun: func(cmd *cobra.Command, _ []string) {
			gApp.Config.SetLocalFlagsFromFile(cmd)
		},
		RunE:         gApp.RunEFileRemove,
		SilenceUsage: true,
	}
	gApp.InitFileRemoveFlags(cmd)
	return cmd
}
