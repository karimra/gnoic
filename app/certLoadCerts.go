package app

import (
	"fmt"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func (a *App) InitCertLoadCertsFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	//
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunELoadCerts(cmd *cobra.Command, args []string) error {
	fmt.Println("not implemented")
	return nil
}
