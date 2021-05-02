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

func (a *App) InitCertCanGenerateCSRFlags(cmd *cobra.Command) {
	cmd.ResetFlags()

	cmd.Flags().StringVar(&a.Config.CertCanGenerateCSRKeyType, "key-type", "", "Key Type")
	cmd.Flags().StringVar(&a.Config.CertCanGenerateCSRCertificateType, "cert-type", "", "Certificate Type")
	cmd.Flags().Uint32Var(&a.Config.CertCanGenerateCSRKeySize, "key-size", 2048, "Key Size")

	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunECertCanGenerateCSR(cmd *cobra.Command, args []string) error {
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
	a.Logger.Info("done...")
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
	if resp == nil {
		fmt.Println("nil response")
		return nil
	}
	if resp.CanGenerate {
		a.Logger.Infof("target %q can generate CSR", t.Config.Address)
	} else {
		a.Logger.Infof("target %q cannot generate CSR", t.Config.Address)
	}
	fmt.Println("###")
	fmt.Println(prototext.Format(resp))
	return nil
}
