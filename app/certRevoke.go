package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func (a *App) InitCertRevokeCertificatesFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunECertRevokeCertificates(cmd *cobra.Command, args []string) error { return nil }
