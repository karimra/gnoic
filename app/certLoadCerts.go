package app

import (
	"context"
	"fmt"
	"os"

	"github.com/openconfig/gnoi/cert"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"

	"github.com/karimra/gnoic/api"
	gcert "github.com/karimra/gnoic/api/cert"
)

type certLoadCert struct {
	TargetError
	rsp *cert.LoadCertificateResponse
}

func (a *App) InitCertLoadCertsFlags(cmd *cobra.Command) {
	cmd.ResetFlags()

	cmd.Flags().StringVar(&a.Config.CertLoadCertificateCertificate, "cert", "", "Certificate")
	cmd.Flags().StringVar(&a.Config.CertLoadCertificateCertificateType, "cert-type", "CT_X509", "Certificate Type")
	cmd.Flags().StringVar(&a.Config.CertLoadCertificateCertificateID, "id", "", "Certificate ID")
	cmd.Flags().StringVar(&a.Config.CertLoadCertificatePrivateKey, "private-key", "", "Private key")
	cmd.Flags().StringVar(&a.Config.CertLoadCertificatePublicKey, "public-key", "", "Public key")
	cmd.Flags().StringSliceVar(&a.Config.CertLoadCertificateCaCertificates, "ca-certs", []string{}, "CA Certificates to load")

	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunELoadCerts(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *certLoadCert, numTargets)
	a.wg.Add(numTargets)
	for _, t := range targets {
		go func(t *api.Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)
			err = t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &certLoadCert{
					TargetError: TargetError{
						TargetName: t.Config.Address,
						Err:        err,
					},
				}
				return
			}
			defer t.Close()
			rsp, err := a.CertLoadCertificate(ctx, t)
			responseChan <- &certLoadCert{
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
	// result := make([]*certLoadCert, 0, numTargets)

	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q Cert LoadCertificate failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
		// result = append(result, rsp)
	}

	return a.handleErrs(errs)
}

func (a *App) CertLoadCertificate(ctx context.Context, t *api.Target) (*cert.LoadCertificateResponse, error) {
	var err error
	opts := []gcert.CertOption{
		gcert.CertificateType(a.Config.CertLoadCertificateCertificateID),
	}
	// certClient := t.CertClient()

	if a.Config.CertLoadCertificateCertificate != "" {
		b, err := os.ReadFile(a.Config.CertLoadCertificateCertificate)
		if err != nil {
			return nil, fmt.Errorf("error reading certificate from file %q: %v",
				a.Config.CertLoadCertificateCertificate, err)
		}
		opts = append(opts,
			gcert.Certificate(
				gcert.CertificateType(a.Config.CertLoadCertificateCertificateType),
				gcert.CertificateBytes(b),
			))
	}

	if a.Config.CertLoadCertificatePublicKey != "" {
		k, err := os.ReadFile(a.Config.CertLoadCertificatePublicKey)
		if err != nil {
			return nil, fmt.Errorf("error reading public key from %q: %v", a.Config.CertLoadCertificatePublicKey, err)
		}
		opts = append(opts, gcert.PublicKey(k))
	}

	if a.Config.CertLoadCertificatePrivateKey != "" {
		k, err := os.ReadFile(a.Config.CertLoadCertificatePrivateKey)
		if err != nil {
			return nil, fmt.Errorf("error reading private key from %q: %v", a.Config.CertLoadCertificatePrivateKey, err)
		}
		opts = append(opts, gcert.PrivateKey(k))
	}

	if n := len(a.Config.CertLoadCertificateCaCertificates); n != 0 {
		for _, certFilename := range a.Config.CertLoadCertificateCaCertificates {
			b, err := os.ReadFile(certFilename)
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
	}

	req, err := gcert.NewCertLoadCertificateRequest(opts...)
	if err != nil {
		return nil, err
	}
	a.printMsg(t.Config.Name, req)

	resp, err := t.CertClient().LoadCertificate(ctx, req)
	if err != nil {
		return nil, err
	}

	a.printMsg(t.Config.Name, resp)
	return resp, nil
}
