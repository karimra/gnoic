package app

import (
	"context"
	"crypto/tls"
	"crypto/x509"
	"encoding/pem"
	"errors"
	"fmt"
	"time"

	"github.com/openconfig/gnoi/cert"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/prototext"
)

func (a *App) InitCertRotateFlags(cmd *cobra.Command) {
	cmd.ResetFlags()

	cmd.Flags().StringVar(&a.Config.CertRotateCertificateID, "id", "", "Certificate ID")
	cmd.Flags().StringVar(&a.Config.CertRotateKeyType, "key-type", "KT_RSA", "Key Type")
	cmd.Flags().StringVar(&a.Config.CertRotateCertificateType, "cert-type", "CT_X509", "Certificate Type")
	cmd.Flags().Uint32Var(&a.Config.CertRotateMinKeySize, "min-key-size", 1024, "Minimum Key Size")
	cmd.Flags().StringVar(&a.Config.CertRotateCommonName, "common-name", "", "CSR common name")
	cmd.Flags().StringVar(&a.Config.CertRotateCountry, "country", "", "CSR country")
	cmd.Flags().StringVar(&a.Config.CertRotateState, "state", "", "CSR state")
	cmd.Flags().StringVar(&a.Config.CertRotateCity, "city", "", "CSR city")
	cmd.Flags().StringVar(&a.Config.CertRotateOrg, "org", "", "CSR organization")
	cmd.Flags().StringVar(&a.Config.CertRotateOrgUnit, "org-unit", "", "CSR organization unit")
	cmd.Flags().StringVar(&a.Config.CertRotateIPAddress, "ip-address", "", "CSR IP address")
	cmd.Flags().StringVar(&a.Config.CertRotateEmailID, "email-id", "", "CSR email ID")
	cmd.Flags().DurationVar(&a.Config.CertRotateValidity, "validity", 87600*time.Hour, "Certificate validity")
	cmd.Flags().BoolVar(&a.Config.CertRotatePrintCSR, "print-csr", false, "print the generated Certificate Signing Request")

	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunECertRotate(cmd *cobra.Command, args []string) error {
	var err error
	if a.Config.CertCACert != "" && a.Config.CertCAKey != "" {
		caCert, err = tls.LoadX509KeyPair(a.Config.CertCACert, a.Config.CertCAKey)
		if err != nil {
			return err
		}
		if len(caCert.Certificate) != 1 {
			return errors.New("CA cert and key contains 0 or more than 1 certificate")
		}
		c, err := x509.ParseCertificate(caCert.Certificate[0])
		if c != nil && err == nil {
			caCert.Leaf = c
		}
		a.Logger.Infof("read local CA certs")
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
			err = a.CertRotate(ctx, t)
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
			wErr := fmt.Errorf("%q Cert Rotate failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
	}
	return a.handleErrs(errs)
}

func (a *App) CertRotate(ctx context.Context, t *Target) error {
	certClient := cert.NewCertificateManagementClient(t.client)
	stream, err := certClient.Rotate(ctx)
	if err != nil {
		return fmt.Errorf("%q failed creating Rotate gRPC stream: %v", t.Config.Address, err)
	}
	var commonName = a.Config.CertInstallCommonName
	var ipAddr = a.Config.CertInstallIPAddress
	if commonName == "" {
		commonName = t.Config.CommonName
	}
	if ipAddr == "" {
		ipAddr = t.Config.ResolvedIP
	}
	err = stream.Send(
		&cert.RotateCertificateRequest{
			RotateRequest: &cert.RotateCertificateRequest_GenerateCsr{
				GenerateCsr: &cert.GenerateCSRRequest{
					CsrParams: &cert.CSRParams{
						Type:               cert.CertificateType(cert.CertificateType_value[a.Config.CertRotateCertificateType]),
						MinKeySize:         a.Config.CertRotateMinKeySize,
						KeyType:            cert.KeyType(cert.KeyType_value[a.Config.CertRotateKeyType]),
						CommonName:         commonName,
						Country:            a.Config.CertRotateCountry,
						State:              a.Config.CertRotateState,
						City:               a.Config.CertRotateCity,
						Organization:       a.Config.CertRotateOrg,
						OrganizationalUnit: a.Config.CertRotateOrgUnit,
						IpAddress:          ipAddr,
						EmailId:            a.Config.CertRotateEmailID,
					},
					CertificateId: a.Config.CertRotateCertificateID,
				},
			},
		})
	if err != nil {
		return fmt.Errorf("%q failed send Rotate RPC: GenCSR: %v", err, t.Config.Address)
	}
	resp, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("%q failed rcv Rotate RPC: GenCSR: %v", err, t.Config.Address)
	}
	if resp == nil {
		return fmt.Errorf("%q returned a <nil> CSR response", t.Config.Address)
	}
	if a.Config.CertRotatePrintCSR {
		fmt.Printf("%q genCSR response:\n %s\n", t.Config.Address, prototext.Format(resp))
	}
	p, rest := pem.Decode(resp.GetGeneratedCsr().GetCsr().Csr)
	if p == nil || len(rest) > 0 {
		return fmt.Errorf("%q failed to decode returned CSR", t.Config.Address)
	}
	creq, err := x509.ParseCertificateRequest(p.Bytes)
	if err != nil {
		return fmt.Errorf("failed parsing certificate request: %v", err)
	}
	if a.Config.CertRotatePrintCSR {
		s, err := CertificateRequestText(creq)
		if err != nil {
			return err
		}
		fmt.Printf("%q generated CSR:\n", t.Config.Address)
		fmt.Printf("%s\n", s)
	}
	certificate, err := certificateFromCSR(creq, a.Config.CertRotateValidity)
	if err != nil {
		return fmt.Errorf("failed certificateFromCSR: %v", err)
	}
	a.Logger.Infof("%q signing certificate %q with the provided CA", t.Config.Address, certificate.Subject.String())
	signedCert, err := a.sign(certificate, &caCert)
	if err != nil {
		return fmt.Errorf("failed signing certificate: %v", err)
	}
	b, err := toPEM(signedCert)
	if err != nil {
		return fmt.Errorf("failed toPEM: %v", err)
	}
	a.Logger.Infof("%q rotating certificate id=%s %q", t.Config.Address, a.Config.CertRotateCertificateID, certificate.Subject.String())
	err = stream.Send(&cert.RotateCertificateRequest{
		RotateRequest: &cert.RotateCertificateRequest_LoadCertificate{
			LoadCertificate: &cert.LoadCertificateRequest{
				Certificate: &cert.Certificate{
					Type:        cert.CertificateType(cert.CertificateType_value[a.Config.CertRotateCertificateType]),
					Certificate: b,
				},
				CertificateId: a.Config.CertRotateCertificateID,
			},
		},
	})
	if err != nil {
		return fmt.Errorf("%q failed sending RotateRequest: %v", t.Config.Address, err)
	}
	_, err = stream.Recv()
	if err != nil {
		return fmt.Errorf("%q RotateRequest LoadCertificate RPC failed: %v", t.Config.Address, err)
	}
	err = stream.Send(&cert.RotateCertificateRequest{
		RotateRequest: &cert.RotateCertificateRequest_FinalizeRotation{
			FinalizeRotation: &cert.FinalizeRequest{},
		},
	})
	if err != nil {
		return fmt.Errorf("%q RotateRequest FinalizeRequest RPC failed: %v", t.Config.Address, err)
	}
	a.Logger.Infof("%q Rotate RPC successful", t.Config.Address)
	return nil
}
