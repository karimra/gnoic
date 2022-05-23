package app

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/karimra/gnoic/api"
	gcert "github.com/karimra/gnoic/api/cert"
	"github.com/openconfig/gnoi/cert"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
)

type certLoadCABundle struct {
	TargetError
	rsp *cert.LoadCertificateAuthorityBundleResponse
}

func (a *App) InitCertLoadCertsCaBundleFlags(cmd *cobra.Command) {
	cmd.ResetFlags()

	cmd.Flags().StringSliceVar(&a.Config.CertLoadCertificateCaBundleCaCertificates, "ca-certs", []string{}, "CA Certificates to load")

	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunELoadCertsCaBundle(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *certLoadCABundle, numTargets)
	a.wg.Add(numTargets)
	for _, t := range targets {
		go func(t *api.Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)
			err = t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &certLoadCABundle{
					TargetError: TargetError{
						TargetName: t.Config.Address,
						Err:        err,
					},
				}
				return
			}
			defer t.Close()
			rsp, err := a.CertLoadCABundle(ctx, t)
			responseChan <- &certLoadCABundle{
				TargetError: TargetError{
					TargetName: t.Config.Address,
					Err:        err,
				},
				rsp: rsp,
			}
		}(t)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	// result := make([]*certLoadCABundle, 0, numTargets)

	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q Cert LoadCA Bundle failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
		// result = append(result, rsp)
	}
	return a.handleErrs(errs)
}

func (a *App) CertLoadCABundle(ctx context.Context, t *api.Target) (*cert.LoadCertificateAuthorityBundleResponse, error) {
	var err error

	n := len(a.Config.CertLoadCertificateCaBundleCaCertificates)
	opts := make([]gcert.CertOption, 0, n)

	for _, certFilename := range a.Config.CertLoadCertificateCaBundleCaCertificates {
		b, err := ioutil.ReadFile(certFilename)
		if err != nil {
			return nil, fmt.Errorf("error reading certificate from file %q: %v",
				certFilename, err)
		}
		opts = append(opts,
			gcert.CaCertificate(
				gcert.CertificateType(a.Config.CertLoadCertificateCertificateType),
				gcert.CertificateBytes(b),
			),
		)
	}
	req, err := gcert.NewCertLoadCertificateAuthorityBundleRequest(opts...)
	if err != nil {
		return nil, err
	}
	a.printMsg(t.Config.Name, req)
	resp, err := t.CertClient().LoadCertificateAuthorityBundle(ctx, req)
	if err != nil {
		return nil, err
	}
	a.printMsg(t.Config.Name, resp)
	return resp, nil
}
