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
		go func(t *Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)
			err = a.CreateGrpcClient(ctx, t, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &certLoadCert{
					TargetError: TargetError{
						TargetName: t.Config.Address,
						Err:        err,
					},
				}
				return
			}
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

func (a *App) CertLoadCertificate(ctx context.Context, t *Target) (*cert.LoadCertificateResponse, error) {
	var err error

	certClient := cert.NewCertificateManagementClient(t.client)
	req := cert.LoadCertificateRequest{
		CertificateId: a.Config.CertLoadCertificateCertificateID,
	}

	if a.Config.CertLoadCertificateCertificate != "" {
		b, err := ioutil.ReadFile(a.Config.CertLoadCertificateCertificate)
		if err != nil {
			return nil, fmt.Errorf("error reading certificate from file %q: %v",
				a.Config.CertLoadCertificateCertificate, err)
		}
		req.Certificate = &cert.Certificate{
			Certificate: b,
			Type:        cert.CertificateType(cert.CertificateType_value[a.Config.CertLoadCertificateCertificateType]),
		}
	}

	if a.Config.CertLoadCertificatePublicKey != "" {
		k, err := ioutil.ReadFile(a.Config.CertLoadCertificatePublicKey)
		if err != nil {
			return nil, fmt.Errorf("error reading public key from %q: %v", a.Config.CertLoadCertificatePublicKey, err)
		}
		req.KeyPair.PublicKey = k
	}

	if a.Config.CertLoadCertificatePrivateKey != "" {
		k, err := ioutil.ReadFile(a.Config.CertLoadCertificatePrivateKey)
		if err != nil {
			return nil, fmt.Errorf("error reading private key from %q: %v", a.Config.CertLoadCertificatePrivateKey, err)
		}
		req.KeyPair.PrivateKey = k
	}

	if n := len(a.Config.CertLoadCertificateCaCertificates); n != 0 {
		req.CaCertificates = make([]*cert.Certificate, n)
		for i, certFilename := range a.Config.CertLoadCertificateCaCertificates {
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

	resp, err := certClient.LoadCertificate(ctx, &req)
	if err != nil {
		return nil, err
	}
	fmt.Println(prototext.Format(resp))
	return resp, nil
}
