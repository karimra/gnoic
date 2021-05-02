package app

import (
	"context"
	"fmt"

	"github.com/openconfig/gnoi/cert"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/prototext"
)

func (a *App) InitCertGenerateCSRFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	//
	cmd.Flags().StringVar(&a.Config.CertGenerateCSRCertificateID, "cert-id", "", "Certificate ID")
	cmd.Flags().StringVar(&a.Config.CertGenerateCSRKeyType, "key-type", "", "Key Type")
	cmd.Flags().StringVar(&a.Config.CertGenerateCSRCertificateType, "cert-type", "", "Certificate Type")
	cmd.Flags().Uint32Var(&a.Config.CertGenerateCSRMinKeySize, "min-key-size", 1024, "Minimum Key Size")
	cmd.Flags().StringVar(&a.Config.CertGenerateCSRCommonName, "common-name", "", "CSR common name")
	cmd.Flags().StringVar(&a.Config.CertGenerateCSRCountry, "country", "", "CSR country")
	cmd.Flags().StringVar(&a.Config.CertGenerateCSRState, "state", "", "CSR state")
	cmd.Flags().StringVar(&a.Config.CertGenerateCSRCity, "city", "", "CSR city")
	cmd.Flags().StringVar(&a.Config.CertGenerateCSROrg, "org", "", "CSR organization")
	cmd.Flags().StringVar(&a.Config.CertGenerateCSROrgUnit, "org-unit", "", "CSR organization unit")
	cmd.Flags().StringVar(&a.Config.CertGenerateCSRIPAddress, "ip-address", "", "CSR IP address")
	cmd.Flags().StringVar(&a.Config.CertGenerateCSREmailID, "email-id", "", "CSR email ID")
	//
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunEGenerateCSR(cmd *cobra.Command, args []string) error {
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
		err = a.CertGenerateCSR(ctx, t)
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

func (a *App) CertGenerateCSR(ctx context.Context, t *Target) error {
	certClient := cert.NewCertificateManagementClient(t.client)
	resp, err := certClient.GenerateCSR(ctx, &cert.GenerateCSRRequest{
		CsrParams: &cert.CSRParams{
			Type:               cert.CertificateType(cert.CertificateType_value[a.Config.CertGenerateCSRCertificateType]),
			MinKeySize:         a.Config.CertGenerateCSRMinKeySize,
			KeyType:            cert.KeyType(cert.KeyType_value[a.Config.CertGenerateCSRKeyType]),
			CommonName:         a.Config.CertGenerateCSRCommonName,
			Country:            a.Config.CertGenerateCSRCountry,
			State:              a.Config.CertGenerateCSRState,
			City:               a.Config.CertGenerateCSRCity,
			Organization:       a.Config.CertGenerateCSROrg,
			OrganizationalUnit: a.Config.CertGenerateCSROrgUnit,
			IpAddress:          a.Config.CertGenerateCSRIPAddress,
			EmailId:            a.Config.CertGenerateCSREmailID,
		},
		CertificateId: a.Config.CertGenerateCSRCertificateID,
	})
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
	return nil
}
