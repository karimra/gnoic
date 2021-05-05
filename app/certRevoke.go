package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/openconfig/gnoi/cert"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
)

func (a *App) InitCertRevokeCertificatesFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringSliceVar(&a.Config.CertRevokeCertificatesCertificateID, "id", []string{}, "Certificate ID")
	cmd.Flags().BoolVar(&a.Config.CertRevokeCertificatesAll, "all", false, "revoke all certificates")
	//
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunECertRevokeCertificates(cmd *cobra.Command, args []string) error {
	if len(a.Config.CertRevokeCertificatesCertificateID) == 0 && !a.Config.CertRevokeCertificatesAll {
		return errors.New("missing certificate ID `--id`")
	}
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *TargetError, numTargets)

	a.wg.Add(numTargets)
	for _, t := range targets {
		go func(t *Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)
			err = a.CreateGrpcClient(ctx, t, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &TargetError{
					TargetName: t.Config.Address,
					Err:        err,
				}
				return
			}
			err = a.Revoke(ctx, t)
			responseChan <- &TargetError{
				TargetName: t.Config.Address,
				Err:        err,
			}
		}(t)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, len(targets))
	for rsp := range responseChan {
		if rsp.Err != nil {
			a.Logger.Errorf("%q cert revoke failed: %v", rsp.TargetName, rsp.Err)
			errs = append(errs, rsp.Err)
			continue
		}
	}

	for _, err := range errs {
		a.Logger.Errorf("err: %v", err)
	}
	if len(errs) > 0 {
		return fmt.Errorf("there was %d error(s)", len(errs))
	}
	a.Logger.Debug("done...")
	return nil
}

func (a *App) Revoke(ctx context.Context, t *Target) error {
	certClient := cert.NewCertificateManagementClient(t.client)
	//
	certificatesID := a.Config.CertRevokeCertificatesCertificateID
	if len(certificatesID) == 0 && a.Config.CertRevokeCertificatesAll {
		certResponse, err := certClient.GetCertificates(ctx, &cert.GetCertificatesRequest{})
		if err != nil {
			return err
		}
		for _, certinfo := range certResponse.CertificateInfo {
			certificatesID = append(certificatesID, certinfo.CertificateId)
		}
	}
	//
	rsp, err := certClient.RevokeCertificates(ctx, &cert.RevokeCertificatesRequest{
		CertificateId: certificatesID,
	})
	if err != nil {
		return err
	}
	for _, revokeErr := range rsp.CertificateRevocationError {
		fmt.Printf("%q certificateID=%s revoke failed: %v\n", t.Config.Address, revokeErr.GetCertificateId(), revokeErr.GetErrorMessage())
	}
	for _, revoked := range rsp.RevokedCertificateId {
		fmt.Printf("%q certificateID=%s revoked successfully\n", t.Config.Address, revoked)
	}

	return nil
}
