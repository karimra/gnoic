package app

import (
	"context"
	"fmt"
	"io/ioutil"

	"github.com/openconfig/gnoi/cert"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/prototext"
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
		go func(t *Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)
			err = a.CreateGrpcClient(ctx, t, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &certLoadCABundle{
					TargetError: TargetError{
						TargetName: t.Config.Address,
						Err:        err,
					},
				}
				return
			}
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

func (a *App) CertLoadCABundle(ctx context.Context, t *Target) (*cert.LoadCertificateAuthorityBundleResponse, error) {
	var err error

	certClient := cert.NewCertificateManagementClient(t.client)
	req := cert.LoadCertificateAuthorityBundleRequest{}

	if n := len(a.Config.CertLoadCertificateCaBundleCaCertificates); n != 0 {
		req.CaCertificates = make([]*cert.Certificate, n)
		for i, certFilename := range a.Config.CertLoadCertificateCaBundleCaCertificates {
			b, err := ioutil.ReadFile(certFilename)
			if err != nil {
				return nil, fmt.Errorf("error reading certificate from file %q: %v",
					certFilename, err)
			}
			req.CaCertificates[i] = &cert.Certificate{
				Certificate: b,
				Type:        cert.CertificateType(cert.CertificateType_value[a.Config.CertLoadCertificateCertificateType]),
			}
		}
	}

	resp, err := certClient.LoadCertificateAuthorityBundle(ctx, &req)
	if err != nil {
		return nil, err
	}
	fmt.Println(prototext.Format(resp))
	return resp, nil
}
