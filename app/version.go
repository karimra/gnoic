package app

import (
	"fmt"

	"github.com/spf13/cobra"
)

var (
	version = "dev"
	commit  = "none"
	date    = "unknown"
	gitURL  = ""
)

func (a *App) RunEVersion(cmd *cobra.Command, args []string) error {
	fmt.Printf("version : %s\n", version)
	fmt.Printf(" commit : %s\n", commit)
	fmt.Printf("   date : %s\n", date)
	fmt.Printf(" gitURL : %s\n", gitURL)
	fmt.Printf("   docs : https://gnoic.kmrd.dev\n")
	return nil
}
