package app

import (
	"context"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"errors"
	"fmt"
	"io"
	"net"
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
	cmd.Flags().BoolVar(&a.Config.CertRotateGenCSR, "gen-csr", false, "generate Certificate Signing Request locally")
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
	if !a.Config.CertRotateGenCSR {
		cgcResp, err := certClient.CanGenerateCSR(ctx, &cert.CanGenerateCSRRequest{
			KeyType:         cert.KeyType(cert.KeyType_value[a.Config.CertRotateKeyType]),
			CertificateType: cert.CertificateType(cert.CertificateType_value[a.Config.CertRotateCertificateType]),
			KeySize:         a.Config.CertRotateMinKeySize,
		})
		if err != nil {
			return fmt.Errorf("%q failed CanGenCSR RPC: %v", t.Config.Name, err)
		}
		if !cgcResp.GetCanGenerate() {
			a.Config.CertRotateGenCSR = true
		}
	}
	keyPair := new(cert.KeyPair)
	var creq *x509.CertificateRequest

	if a.Config.CertRotateGenCSR {
		keyPair, creq, err = a.createLocalCSRRotate(t)
	} else {
		creq, err = a.createRemoteCSRRotate(stream, t)
	}
	if err != nil {
		return err
	}

	s, err := CertificateRequestText(creq)
	if err != nil {
		return err
	}
	if a.Config.CertRotatePrintCSR {
		fmt.Printf("%q generated CSR:\n%s\n", t.Config.Address, s)
	}
	a.Logger.Debugf("%q generated CSR:\n%s\n", t.Config.Address, s)

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
	if a.Config.CertRotateGenCSR {
		req := &cert.RotateCertificateRequest{
			RotateRequest: &cert.RotateCertificateRequest_LoadCertificate{
				LoadCertificate: &cert.LoadCertificateRequest{
					Certificate: &cert.Certificate{
						Type:        cert.CertificateType(cert.CertificateType_value[a.Config.CertRotateCertificateType]),
						Certificate: b,
					},
					KeyPair:       keyPair,
					CertificateId: a.Config.CertRotateCertificateID,
				},
			},
		}
		a.printMsg(t.Config.Name, req)
		err = stream.Send(req)
	} else {
		req := &cert.RotateCertificateRequest{
			RotateRequest: &cert.RotateCertificateRequest_LoadCertificate{
				LoadCertificate: &cert.LoadCertificateRequest{
					Certificate: &cert.Certificate{
						Type:        cert.CertificateType(cert.CertificateType_value[a.Config.CertRotateCertificateType]),
						Certificate: b,
					},
					CertificateId: a.Config.CertRotateCertificateID,
				},
			},
		}
		a.printMsg(t.Config.Name, req)
		err = stream.Send(req)
	}
	if err != nil {
		return fmt.Errorf("%q failed sending RotateRequest: %v", t.Config.Address, err)
	}
	resp, err := stream.Recv()
	if err != nil {
		return fmt.Errorf("%q RotateRequest LoadCertificate RPC failed: %v", t.Config.Address, err)
	}
	a.printMsg(t.Config.Name, resp)
	finalizeReq := &cert.RotateCertificateRequest{
		RotateRequest: &cert.RotateCertificateRequest_FinalizeRotation{
			FinalizeRotation: &cert.FinalizeRequest{},
		},
	}
	a.printMsg(t.Config.Name, finalizeReq)
	err = stream.Send(finalizeReq)
	if err != nil {
		return fmt.Errorf("%q RotateRequest FinalizeRequest RPC failed: %v", t.Config.Address, err)
	}
	resp, err = stream.Recv()
	if err != nil && !errors.Is(err, io.EOF) {
		return err
	}
	a.printMsg(t.Config.Name, resp)
	a.Logger.Infof("%q Rotate RPC successful", t.Config.Address)
	return nil
}

func (a *App) createLocalCSRRotate(t *Target) (*cert.KeyPair, *x509.CertificateRequest, error) {
	var commonName = a.Config.CertRotateCommonName
	var ipAddr = a.Config.CertRotateIPAddress
	if commonName == "" {
		commonName = t.Config.CommonName
	}
	if ipAddr == "" {
		ipAddr = t.Config.ResolvedIP
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, int(a.Config.CertRotateMinKeySize))
	if err != nil {
		return nil, nil, err
	}

	var subj pkix.Name
	if commonName != "" {
		subj.CommonName = commonName
	}
	if a.Config.CertRotateCountry != "" {
		subj.Country = []string{a.Config.CertRotateCountry}
	}
	if a.Config.CertRotateState != "" {
		subj.Province = []string{a.Config.CertRotateState}
	}
	if a.Config.CertRotateCity != "" {
		subj.Locality = []string{a.Config.CertRotateCity}
	}
	if a.Config.CertRotateOrg != "" {
		subj.Organization = []string{a.Config.CertRotateOrg}
	}
	if a.Config.CertRotateOrgUnit != "" {
		subj.OrganizationalUnit = []string{a.Config.CertRotateOrgUnit}
	}
	if a.Config.CertRotateEmailID != "" {
		subj.ExtraNames = append(subj.ExtraNames, pkix.AttributeTypeAndValue{
			Type: oidEmailAddress,
			Value: asn1.RawValue{
				Tag:   asn1.TagIA5String,
				Bytes: []byte(a.Config.CertRotateEmailID),
			},
		})
	}

	var ipAddrs net.IP
	if ipAddr != "" {
		ipAddrs = net.ParseIP(ipAddr)
	}
	template := x509.CertificateRequest{
		Subject:            subj,
		EmailAddresses:     []string{a.Config.CertRotateEmailID},
		SignatureAlgorithm: x509.SHA256WithRSA,
		IPAddresses:        make([]net.IP, 0),
	}

	if ipAddrs != nil {
		template.IPAddresses = append(template.IPAddresses, ipAddrs)
	}

	csrBytes, err := x509.CreateCertificateRequest(rand.Reader, &template, privateKey)
	if err != nil {
		return nil, nil, fmt.Errorf("failed to create Certificate Request: %v", err)
	}
	creq, err := x509.ParseCertificateRequest(csrBytes)
	if err != nil {
		return nil, nil, fmt.Errorf("failed parsing certificate request: %v", err)
	}
	return &cert.KeyPair{
			PrivateKey: pem.EncodeToMemory(&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
			}),
			PublicKey: csrBytes,
		},
		creq, err
}

func (a *App) createRemoteCSRRotate(stream cert.CertificateManagement_RotateClient, t *Target) (*x509.CertificateRequest, error) {
	var commonName = a.Config.CertRotateCommonName
	var ipAddr = a.Config.CertRotateIPAddress
	if commonName == "" {
		commonName = t.Config.CommonName
	}
	if ipAddr == "" {
		ipAddr = t.Config.ResolvedIP
	}
	req := &cert.RotateCertificateRequest{
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
	}
	a.printMsg(t.Config.Name, req)
	err := stream.Send(req)
	if err != nil {
		return nil, fmt.Errorf("%q failed send Rotate RPC: GenCSR: %v", err, t.Config.Address)
	}
	resp, err := stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("%q failed rcv Rotate RPC: GenCSR: %v", err, t.Config.Address)
	}
	if resp == nil {
		return nil, fmt.Errorf("%q returned a <nil> CSR response", t.Config.Address)
	}
	a.printMsg(t.Config.Name, resp)
	if a.Config.CertRotatePrintCSR {
		fmt.Printf("%q genCSR response:\n %s\n", t.Config.Address, prototext.Format(resp))
	}

	p, rest := pem.Decode(resp.GetGeneratedCsr().GetCsr().GetCsr())
	if p == nil || len(rest) > 0 {
		return nil, fmt.Errorf("%q failed to decode returned CSR", t.Config.Address)
	}
	creq, err := x509.ParseCertificateRequest(p.Bytes)
	if err != nil {
		return nil, fmt.Errorf("failed parsing certificate request: %v", err)
	}
	return creq, nil
}
