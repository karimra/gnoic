package app

import (
	"bytes"
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
	"net"
	"time"

	"github.com/openconfig/gnoi/cert"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/prototext"
)

func (a *App) InitCertInstallFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.CertInstallCertificateID, "id", "", "Certificate ID")
	cmd.Flags().StringVar(&a.Config.CertInstallKeyType, "key-type", "KT_RSA", "Key Type")
	cmd.Flags().StringVar(&a.Config.CertInstallCertificateType, "cert-type", "CT_X509", "Certificate Type")
	cmd.Flags().Uint32Var(&a.Config.CertInstallMinKeySize, "min-key-size", 2048, "Minimum Key Size")
	cmd.Flags().StringVar(&a.Config.CertInstallCommonName, "common-name", "", "CSR common name")
	cmd.Flags().StringVar(&a.Config.CertInstallCountry, "country", "", "CSR country")
	cmd.Flags().StringVar(&a.Config.CertInstallState, "state", "", "CSR state")
	cmd.Flags().StringVar(&a.Config.CertInstallCity, "city", "", "CSR city")
	cmd.Flags().StringVar(&a.Config.CertInstallOrg, "org", "", "CSR organization")
	cmd.Flags().StringVar(&a.Config.CertInstallOrgUnit, "org-unit", "", "CSR organization unit")
	cmd.Flags().StringVar(&a.Config.CertInstallIPAddress, "ip-address", "", "CSR IP address")
	cmd.Flags().StringVar(&a.Config.CertInstallEmailID, "email-id", "", "CSR email ID")
	cmd.Flags().DurationVar(&a.Config.CertInstallValidity, "validity", 10*365*24*time.Hour, "certificate validity")
	cmd.Flags().BoolVar(&a.Config.CertInstallPrintCSR, "print-csr", false, "print the generated Certificate Signing Request")
	cmd.Flags().BoolVar(&a.Config.CertInstallGenCSR, "gen-csr", false, "generate Certificate Signing Request locally")
	//
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunECertInstall(cmd *cobra.Command, args []string) error {
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
			err = a.CertInstall(ctx, t)
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
			wErr := fmt.Errorf("%q Cert Install failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
	}
	return a.handleErrs(errs)
}

func (a *App) CertInstall(ctx context.Context, t *Target) error {
	certClient := cert.NewCertificateManagementClient(t.client)
	stream, err := certClient.Install(ctx)
	if err != nil {
		return fmt.Errorf("%q failed creating Install gRPC stream: %v", t.Config.Address, err)
	}
	var commonName = a.Config.CertInstallCommonName
	var ipAddr = a.Config.CertInstallIPAddress
	if commonName == "" {
		commonName = t.Config.CommonName
	}
	if ipAddr == "" {
		ipAddr = t.Config.ResolvedIP
	}
	// if the flag --gen-csr is not present,
	// check if the target can generateCSR,
	// if it cannot, set --gen-csr to true,
	//  to generate the certificate locally
	if !a.Config.CertInstallGenCSR {
		cgcResp, err := certClient.CanGenerateCSR(ctx, &cert.CanGenerateCSRRequest{
			KeyType:         cert.KeyType(cert.KeyType_value[a.Config.CertInstallKeyType]),
			CertificateType: cert.CertificateType(cert.CertificateType_value[a.Config.CertInstallCertificateType]),
			KeySize:         a.Config.CertInstallMinKeySize,
		})
		if err != nil {
			return err
		}
		if !cgcResp.GetCanGenerate() {
			a.Config.CertInstallGenCSR = true
		}
	}

	keyPair := new(cert.KeyPair)
	var creq *x509.CertificateRequest

	if a.Config.CertInstallGenCSR {
		keyPair, creq, err = a.createLocalCSR(t)
		if err != nil {
			return err
		}
	} else {
		err = stream.Send(&cert.InstallCertificateRequest{
			InstallRequest: &cert.InstallCertificateRequest_GenerateCsr{
				GenerateCsr: &cert.GenerateCSRRequest{
					CsrParams: &cert.CSRParams{
						Type:               cert.CertificateType(cert.CertificateType_value[a.Config.CertInstallCertificateType]),
						MinKeySize:         a.Config.CertInstallMinKeySize,
						KeyType:            cert.KeyType(cert.KeyType_value[a.Config.CertInstallKeyType]),
						CommonName:         commonName,
						Country:            a.Config.CertInstallCountry,
						State:              a.Config.CertInstallState,
						City:               a.Config.CertInstallCity,
						Organization:       a.Config.CertInstallOrg,
						OrganizationalUnit: a.Config.CertInstallOrgUnit,
						IpAddress:          ipAddr,
						EmailId:            a.Config.CertInstallEmailID,
					},
					CertificateId: a.Config.CertInstallCertificateID,
				},
			},
		})
		if err != nil {
			return fmt.Errorf("%q failed send Install RPC: GenCSR: %v", err, t.Config.Address)
		}
		resp, err := stream.Recv()
		if err != nil {
			return fmt.Errorf("%q failed rcv Install RPC: GenCSR: %v", err, t.Config.Address)
		}
		if resp == nil {
			return fmt.Errorf("%q returned a <nil> CSR response", t.Config.Address)
		}
		if a.Config.CertInstallPrintCSR {
			fmt.Printf("%q genCSR response:\n %s\n", t.Config.Address, prototext.Format(resp))
		}
		a.Logger.Debugf("%q genCSR response:\n %s\n", t.Config.Address, prototext.Format(resp))

		p, rest := pem.Decode(resp.GetGeneratedCsr().GetCsr().GetCsr())
		if p == nil || len(rest) > 0 {
			return fmt.Errorf("%q failed to decode returned CSR", t.Config.Address)
		}
		creq, err = x509.ParseCertificateRequest(p.Bytes)
		if err != nil {
			return fmt.Errorf("failed parsing certificate request: %v", err)
		}
	}

	s, err := CertificateRequestText(creq)
	if err != nil {
		return err
	}
	if a.Config.CertInstallPrintCSR {
		fmt.Printf("%q generated CSR:\n%s\n", t.Config.Address, s)
	}
	a.Logger.Debugf("%q generated CSR:\n%s\n", t.Config.Address, s)
	//
	certificate, err := certificateFromCSR(creq, a.Config.CertInstallValidity)
	if err != nil {
		return fmt.Errorf("failed certificateFromCSR: %v", err)
	}
	a.Logger.Infof("%q signing certificate %q with the provided CA", t.Config.Address, certificate.Subject.String())
	signedCert, err := a.sign(certificate, &caCert)
	if err != nil {
		return fmt.Errorf("failed signing certificate: %v", err)
	}
	//
	sCertText, err := CertificateText(signedCert, false)
	if err != nil {
		return err
	}
	a.Logger.Debugf("%q signed certificate:\n%s\n", t.Config.Address, sCertText)
	//
	b, err := toPEM(signedCert)
	if err != nil {
		return fmt.Errorf("failed toPEM: %v", err)
	}
	a.Logger.Infof("%q installing certificate id=%s %q", t.Config.Address, a.Config.CertInstallCertificateID, certificate.Subject.String())
	if a.Config.CertInstallGenCSR {
		err = stream.Send(&cert.InstallCertificateRequest{
			InstallRequest: &cert.InstallCertificateRequest_LoadCertificate{
				LoadCertificate: &cert.LoadCertificateRequest{
					Certificate: &cert.Certificate{
						Type:        cert.CertificateType(cert.CertificateType_value[a.Config.CertInstallCertificateType]),
						Certificate: b,
					},
					KeyPair:       keyPair,
					CertificateId: a.Config.CertInstallCertificateID,
				},
			},
		})
	} else {
		err = stream.Send(&cert.InstallCertificateRequest{
			InstallRequest: &cert.InstallCertificateRequest_LoadCertificate{
				LoadCertificate: &cert.LoadCertificateRequest{
					Certificate: &cert.Certificate{
						Type:        cert.CertificateType(cert.CertificateType_value[a.Config.CertInstallCertificateType]),
						Certificate: b,
					},
					CertificateId: a.Config.CertInstallCertificateID,
				},
			},
		})
	}
	if err != nil {
		return fmt.Errorf("%q failed sending InstallRequest: %v", t.Config.Address, err)
	}
	_, err = stream.Recv()
	if err != nil {
		return fmt.Errorf("%q InstallRequest RPC failed: %v", t.Config.Address, err)
	}
	a.Logger.Infof("%q Install RPC successful", t.Config.Address)
	return nil
}

func (a *App) createLocalCSR(t *Target) (*cert.KeyPair, *x509.CertificateRequest, error) {
	var commonName = a.Config.CertInstallCommonName
	var ipAddr = a.Config.CertInstallIPAddress
	if commonName == "" {
		commonName = t.Config.CommonName
	}
	if ipAddr == "" {
		ipAddr = t.Config.ResolvedIP
	}
	privateKey, err := rsa.GenerateKey(rand.Reader, int(a.Config.CertInstallMinKeySize))
	if err != nil {
		return nil, nil, err
	}
	buf := new(bytes.Buffer)
	err = pem.Encode(buf, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
	})
	if err != nil {
		return nil, nil, err
	}
	subj := pkix.Name{}
	if commonName != "" {
		subj.CommonName = commonName
	}
	if a.Config.CertInstallCountry != "" {
		subj.Country = []string{a.Config.CertInstallCountry}
	}
	if a.Config.CertInstallState != "" {
		subj.Province = []string{a.Config.CertInstallState}
	}
	if a.Config.CertInstallCity != "" {
		subj.Locality = []string{a.Config.CertInstallCity}
	}
	if a.Config.CertInstallOrg != "" {
		subj.Organization = []string{a.Config.CertInstallOrg}
	}
	if a.Config.CertInstallOrgUnit != "" {
		subj.OrganizationalUnit = []string{a.Config.CertInstallOrgUnit}
	}
	if a.Config.CertInstallEmailID != "" {
		subj.ExtraNames = append(subj.ExtraNames, pkix.AttributeTypeAndValue{
			Type: oidEmailAddress,
			Value: asn1.RawValue{
				Tag:   asn1.TagIA5String,
				Bytes: []byte(a.Config.CertInstallEmailID),
			},
		})
	}
	var ipAddrs net.IP
	if a.Config.CertInstallIPAddress != "" {
		ipAddrs = net.ParseIP(ipAddr)
	}
	template := x509.CertificateRequest{
		Subject:            subj,
		EmailAddresses:     []string{a.Config.CertInstallEmailID},
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
			PrivateKey: buf.Bytes(),
			PublicKey:  csrBytes,
		},
		creq, err
}
