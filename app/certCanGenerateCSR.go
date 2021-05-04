package app

import (
	"context"
	"fmt"

	"github.com/openconfig/gnoi/cert"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
)

func (a *App) InitCertCanGenerateCSRFlags(cmd *cobra.Command) {
	cmd.ResetFlags()

	cmd.Flags().StringVar(&a.Config.CertCanGenerateCSRKeyType, "key-type", "KT_RSA", "Key Type")
	cmd.Flags().StringVar(&a.Config.CertCanGenerateCSRCertificateType, "cert-type", "CT_X509", "Certificate Type")
	cmd.Flags().Uint32Var(&a.Config.CertCanGenerateCSRKeySize, "key-size", 2048, "Key Size")

	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunECertCanGenerateCSR(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
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
		err = a.CertCanGenerateCSR(ctx, t)
		if err != nil {
			a.Logger.Errorf("%q can generate CSR failed: %v", t.Config.Address, err)
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
	a.Logger.Debug("done...")
	return nil
}

func (a *App) CertCanGenerateCSR(ctx context.Context, t *Target) error {
	certClient := cert.NewCertificateManagementClient(t.client)
	resp, err := certClient.CanGenerateCSR(ctx,
		&cert.CanGenerateCSRRequest{
			KeyType:         cert.KeyType(cert.KeyType_value[a.Config.CertCanGenerateCSRKeyType]),
			CertificateType: cert.CertificateType(cert.CertificateType_value[a.Config.CertCanGenerateCSRCertificateType]),
			KeySize:         a.Config.CertCanGenerateCSRKeySize,
		})
	if err != nil {
		return err
	}
	fmt.Printf("%q key-type=%s, cert-type=%s, key-size=%d: can_generate: %v\n",
		t.Config.Address,
		a.Config.CertCanGenerateCSRKeyType,
		a.Config.CertCanGenerateCSRCertificateType,
		a.Config.CertCanGenerateCSRKeySize,
		resp.GetCanGenerate())
	return nil
}
