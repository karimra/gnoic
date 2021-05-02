package app

import (
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"

	"github.com/davecgh/go-spew/spew"
	"github.com/openconfig/gnoi/cert"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/prototext"
)

func (a *App) InitCertInstallFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.CertInstallCertificateID, "cert-id", "", "Certificate ID")
	cmd.Flags().StringVar(&a.Config.CertInstallKeyType, "key-type", "KT_RSA", "Key Type")
	cmd.Flags().StringVar(&a.Config.CertInstallCertificateType, "cert-type", "CT_X509", "Certificate Type")
	cmd.Flags().Uint32Var(&a.Config.CertInstallMinKeySize, "min-key-size", 1024, "Minimum Key Size")
	cmd.Flags().StringVar(&a.Config.CertInstallCommonName, "common-name", "", "CSR common name")
	cmd.Flags().StringVar(&a.Config.CertInstallCountry, "country", "", "CSR country")
	cmd.Flags().StringVar(&a.Config.CertInstallState, "state", "", "CSR state")
	cmd.Flags().StringVar(&a.Config.CertInstallCity, "city", "", "CSR city")
	cmd.Flags().StringVar(&a.Config.CertInstallOrg, "org", "", "CSR organization")
	cmd.Flags().StringVar(&a.Config.CertInstallOrgUnit, "org-unit", "", "CSR organization unit")
	cmd.Flags().StringVar(&a.Config.CertInstallIPAddress, "ip-address", "", "CSR IP address")
	cmd.Flags().StringVar(&a.Config.CertInstallEmailID, "email-id", "", "CSR email ID")
	//
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunECertInstall(cmd *cobra.Command, args []string) error {
	targetsConfigs, err := a.Config.GetTargets()
	if err != nil {
		return err
	}
	targets := make(map[string]*Target)
	for n, tc := range targetsConfigs {
		targets[n] = NewTarget(tc)
	}
	errs := make([]error, 0, len(targets))
	for _, t := range targets {
		ctx, cancel := context.WithCancel(a.ctx)
		defer cancel()
		ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)
		err = a.CreateGrpcClient(ctx, t, a.createBaseDialOpts()...)
		if err != nil {
			errs = append(errs, err)
			continue
		}
		err = a.CertInstall(ctx, t)
		if err != nil {
			a.Logger.Errorf("%q genrate CSR failed: %v", t.Config.Address, err)
			errs = append(errs, err)
			continue
		}
	}

	for _, err := range errs {
		a.Logger.Errorf("err: %v", err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("there was %d errors", len(errs))
	}
	a.Logger.Info("done...")
	return nil
}

func (a *App) CertInstall(ctx context.Context, t *Target) error {
	certClient := cert.NewCertificateManagementClient(t.client)
	stream, err := certClient.Install(ctx)
	if err != nil {
		return err
	}
	err = stream.Send(&cert.InstallCertificateRequest{
		InstallRequest: &cert.InstallCertificateRequest_GenerateCsr{
			GenerateCsr: &cert.GenerateCSRRequest{
				CsrParams: &cert.CSRParams{
					Type:               cert.CertificateType(cert.CertificateType_value[a.Config.CertInstallCertificateType]),
					MinKeySize:         a.Config.CertInstallMinKeySize,
					KeyType:            cert.KeyType(cert.KeyType_value[a.Config.CertInstallKeyType]),
					CommonName:         a.Config.CertInstallCommonName,
					Country:            a.Config.CertInstallCountry,
					State:              a.Config.CertInstallState,
					City:               a.Config.CertInstallCity,
					Organization:       a.Config.CertInstallOrg,
					OrganizationalUnit: a.Config.CertInstallOrgUnit,
					IpAddress:          a.Config.CertInstallIPAddress,
					EmailId:            a.Config.CertInstallEmailID,
				},
				CertificateId: a.Config.CertInstallCertificateID,
			},
		},
	})
	if err != nil {
		return err
	}
	resp, err := stream.Recv()
	if err != nil {
		return err
	}
	if resp == nil {
		fmt.Println("nil response")
	} else {
		fmt.Printf("%+v\n", resp)
	}
	fmt.Println("###")
	fmt.Println(prototext.Format(resp))
	p, rest := pem.Decode(resp.GetGeneratedCsr().GetCsr().Csr)
	fmt.Println(p.Type)
	fmt.Println(p.Headers)
	creq, err := x509.ParseCertificateRequest(p.Bytes)
	if err != nil {
		return err
	}
	spew.Dump(creq)
	fmt.Println(string(rest))
	return nil
}
