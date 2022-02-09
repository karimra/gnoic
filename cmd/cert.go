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
	"github.com/spf13/cobra"
)

// certCmd represents the cert command
func newCertCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "cert",
		Short: "run Certificate Management gNOI RPCs",
		PreRun: func(cmd *cobra.Command, args []string) {
			gApp.Config.SetPersistantFlagsFromFile(cmd)
		},
		SilenceUsage: true,
	}
	gApp.InitCertFlags(cmd)
	return cmd
}

// newCertRotateCmd represents the cert rotate command
func newCertRotateCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "rotate",
		Short: "run certificate Rotate gNOI RPC",
		PreRun: func(cmd *cobra.Command, args []string) {
			gApp.Config.SetLocalFlagsFromFile(cmd)
		},
		RunE:         gApp.RunECertRotate,
		SilenceUsage: true,
	}
	gApp.InitCertRotateFlags(cmd)
	return cmd
}

// newCertInstallCmd represents the cert install command
func newCertInstallCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "install",
		Short: "run certificate Install gNOI RPC",
		PreRun: func(cmd *cobra.Command, args []string) {
			gApp.Config.SetLocalFlagsFromFile(cmd)
		},
		RunE:         gApp.RunECertInstall,
		SilenceUsage: true,
	}
	gApp.InitCertInstallFlags(cmd)
	return cmd
}

// newCertGenCSRCmd represents the cert generate csr command
func newCertGenCSRCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "generate-csr",
		Aliases: []string{"gcsr", "gc"},
		Short:   "run certificate GenerateCSR gNOI RPC",
		PreRun: func(cmd *cobra.Command, args []string) {
			gApp.Config.SetLocalFlagsFromFile(cmd)
		},
		RunE:         gApp.RunEGenerateCSR,
		SilenceUsage: true,
	}
	gApp.InitCertGenerateCSRFlags(cmd)
	return cmd
}

// newCertLoadCertificatesCmd represents the cert load-certificates command
func newCertLoadCertificatesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "load",
		Aliases: []string{"load-certs"},
		Short:   "run certificate LoadCertificates gNOI RPC",
		PreRun: func(cmd *cobra.Command, args []string) {
			gApp.Config.SetLocalFlagsFromFile(cmd)
		},
		RunE:         gApp.RunELoadCerts,
		SilenceUsage: true,
	}
	gApp.InitCertLoadCertsFlags(cmd)
	return cmd
}

// newCertLoadCertificateAuthorityBundleCmd represents the cert load-certificates-ca-bundle command
func newCertLoadCertificateAuthorityBundleCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "load-ca",
		Aliases: []string{"load-certs-ca-bundle"},
		Short:   "run certificate LoadCertificateAuthorityBundle gNOI RPC",
		PreRun: func(cmd *cobra.Command, args []string) {
			gApp.Config.SetLocalFlagsFromFile(cmd)
		},
		RunE:         gApp.RunELoadCertsCaBundle,
		SilenceUsage: true,
	}
	gApp.InitCertLoadCertsCaBundleFlags(cmd)
	return cmd
}

// newCertGetCertificatesCmd represents the cert GetCertificates command
func newCertGetCertificatesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "get-certs",
		Aliases: []string{"get"},
		Short:   "run certificate GetCertificates gNOI RPC",
		PreRun: func(cmd *cobra.Command, args []string) {
			gApp.Config.SetLocalFlagsFromFile(cmd)
		},
		RunE:         gApp.RunECertGetCertificates,
		SilenceUsage: true,
	}
	gApp.InitCertGetCertificatesFlags(cmd)
	return cmd
}

// newCertRevokeCertificatesCmd represents the cert RevokeCertificates command
func newCertRevokeCertificatesCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "revoke",
		Aliases: []string{"revoke-certs", "rev"},
		Short:   "run certificate RevokeCertificates gNOI RPC",
		PreRun: func(cmd *cobra.Command, args []string) {
			gApp.Config.SetLocalFlagsFromFile(cmd)
		},
		RunE:         gApp.RunECertRevokeCertificates,
		SilenceUsage: true,
	}
	gApp.InitCertRevokeCertificatesFlags(cmd)
	return cmd
}

// newCertCanGenerateCSRCmd represents the cert CanGenerateCSR command
func newCertCanGenerateCSRCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:     "can-generate-csr",
		Aliases: []string{"cgc"},
		Short:   "run certificate CanGenerateCSR gNOI RPC",
		PreRun: func(cmd *cobra.Command, args []string) {
			gApp.Config.SetLocalFlagsFromFile(cmd)
		},
		RunE:         gApp.RunECertCanGenerateCSR,
		SilenceUsage: true,
	}
	gApp.InitCertCanGenerateCSRFlags(cmd)
	return cmd
}

// newCertGenCSRCmd represents the cert create-ca command
func newCertCreateCaCmd() *cobra.Command {
	cmd := &cobra.Command{
		Use:   "create-ca",
		Short: "create a CA Certificate and Key",
		PreRun: func(cmd *cobra.Command, args []string) {
			gApp.Config.SetLocalFlagsFromFile(cmd)
		},
		RunE:         gApp.RunECertCreateCa,
		SilenceUsage: true,
	}
	gApp.InitCertCreateCaFlags(cmd)
	return cmd
}
