/*
Copyright © 2021 Karim Radhouani <medkarimrdi@gmail.com>

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

    http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/
package cmd

import (
	"fmt"
	"os"

	"github.com/karimra/gnoic/app"
	"github.com/spf13/cobra"
	"github.com/spf13/viper"
)

var gApp = app.New()

func newRootCmd() *cobra.Command {
	gApp.RootCmd = &cobra.Command{
		Use:               "gnoic",
		Short:             "run gNOI RPCs from the terminal",
		PersistentPreRunE: gApp.PreRun,
	}
	gApp.InitGlobalFlags()
	//
	certCmd := newCertCmd()
	certCmd.AddCommand(newCertRotateCmd())
	certCmd.AddCommand(newCertInstallCmd())
	certCmd.AddCommand(newCertGenCSRCmd())
	certCmd.AddCommand(newCertLoadCertificatesCmd())
	certCmd.AddCommand(newCertLoadCertificateAuthorityBundleCmd())
	certCmd.AddCommand(newCertGetCertificatesCmd())
	certCmd.AddCommand(newCertRevokeCertificatesCmd())
	certCmd.AddCommand(newCertCanGenerateCSRCmd())
	certCmd.AddCommand(newCertCreateCaCmd())
	//
	fileCmd := newFileCmd()
	//
	systemCmd := newSystemCmd()
	//
	gApp.RootCmd.AddCommand(certCmd)
	gApp.RootCmd.AddCommand(fileCmd)
	gApp.RootCmd.AddCommand(systemCmd)
	//
	return gApp.RootCmd
}

// Execute adds all child commands to the root command and sets flags appropriately.
// This is called by main.main(). It only needs to happen once to the rootCmd.
func Execute() {
	err := newRootCmd().Execute()
	if err != nil {
		os.Exit(1)
	}
}

func init() {
	cobra.OnInitialize(initConfig)
}

// initConfig reads in config file and ENV variables if set.
func initConfig() {
	err := gApp.Config.Load()
	if err == nil {
		return
	}
	if _, ok := err.(viper.ConfigFileNotFoundError); !ok {
		fmt.Fprintf(os.Stderr, "failed loading config file: %v\n", err)
	}
}
