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

func (a *App) InitCertGetCertificatesFlags(cmd *cobra.Command) {
	cmd.ResetFlags()

	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunECertGetCertificates(cmd *cobra.Command, args []string) error {
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
		err = a.CertGetCertificates(ctx, t)
		if err != nil {
			a.Logger.Errorf("%q get certificates failed: %v", t.Config.Address, err)
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

func (a *App) CertGetCertificates(ctx context.Context, t *Target) error {
	certClient := cert.NewCertificateManagementClient(t.client)

	resp, err := certClient.GetCertificates(ctx, new(cert.GetCertificatesRequest))
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
