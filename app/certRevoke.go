package app

import (
	"context"
	"errors"
	"fmt"

	"github.com/karimra/gnoic/api"
	gcert "github.com/karimra/gnoic/api/cert"
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
		go func(t *api.Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)
			err = t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &TargetError{
					TargetName: t.Config.Address,
					Err:        err,
				}
				return
			}
			defer t.Close()
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
			wErr := fmt.Errorf("%q Cert Revoke failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
	}
	return a.handleErrs(errs)
}

func (a *App) Revoke(ctx context.Context, t *api.Target) error {
	certClient := t.CertClient()
	//
	opts := make([]gcert.CertOption, 0, len(a.Config.CertRevokeCertificatesCertificateID))
	for _, cid := range a.Config.CertRevokeCertificatesCertificateID {
		opts = append(opts, gcert.CertificateID(cid))
	}

	if len(opts) == 0 && a.Config.CertRevokeCertificatesAll {
		certResponse, err := certClient.GetCertificates(ctx, &cert.GetCertificatesRequest{})
		if err != nil {
			return err
		}
		opts = make([]gcert.CertOption, 0, len(certResponse.GetCertificateInfo()))
		for _, certinfo := range certResponse.GetCertificateInfo() {
			opts = append(opts, gcert.CertificateID(certinfo.CertificateId))
		}
	}
	//
	req, err := gcert.NewCertRevokeCertificatesRequest(opts...)
	if err != nil {
		return err
	}

	a.printMsg(t.Config.Name, req)
	resp, err := certClient.RevokeCertificates(ctx, req)
	if err != nil {
		return err
	}
	a.printMsg(t.Config.Name, resp)
	for _, revokeErr := range resp.CertificateRevocationError {
		a.Logger.Errorf("%q certificateID=%s revoke failed: %v\n", t.Config.Address, revokeErr.GetCertificateId(), revokeErr.GetErrorMessage())
	}
	for _, revoked := range resp.RevokedCertificateId {
		a.Logger.Infof("%q certificateID=%s revoked successfully\n", t.Config.Address, revoked)
	}

	return nil
}
