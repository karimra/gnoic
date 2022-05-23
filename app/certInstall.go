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
	"net"
	"time"

	"github.com/karimra/gnoic/api"
	gcert "github.com/karimra/gnoic/api/cert"
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
	cmd.Flags().Uint32Var(&a.Config.CertInstallMinKeySize, "min-key-size", 1024, "Minimum Key Size")
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

func (a *App) CertInstall(ctx context.Context, t *api.Target) error {
	// create cert mgmt install stream RPC
	stream, err := t.CertClient().Install(ctx)
	if err != nil {
		return fmt.Errorf("%q failed creating Install gRPC stream: %v", t.Config.Address, err)
	}
	// if the flag --gen-csr is not present,
	// check if the target can generateCSR,
	// if it cannot, set --gen-csr to true,
	// to generate the certificate locally
	if !a.Config.CertInstallGenCSR {
		cgcReq, err := gcert.NewCertCanGenerateCSRRequest(
			gcert.CertificateType(a.Config.CertInstallCertificateType),
			gcert.KeyType(a.Config.CertInstallKeyType),
			gcert.KeySize(a.Config.CertInstallMinKeySize),
		)
		if err != nil {
			return err
		}
		cgcResp, err := t.CertClient().CanGenerateCSR(ctx, cgcReq)
		if err != nil {
			return fmt.Errorf("%q failed Cert CanGenCSR RPC: %w", t.Config.Name, err)
		}
		if !cgcResp.GetCanGenerate() {
			a.Config.CertInstallGenCSR = true
		}
	}

	keyPair := new(cert.KeyPair)
	var creq *x509.CertificateRequest

	if a.Config.CertInstallGenCSR {
		keyPair, creq, err = a.createLocalCSRInstall(t)
	} else {
		creq, err = a.createRemoteCSRInstall(stream, t)
	}
	if err != nil {
		return err
	}

	s, err := CertificateRequestText(creq)
	if err != nil {
		return err
	}
	if a.Config.CertInstallPrintCSR {
		fmt.Printf("%q generated CSR:\n%s\n", t.Config.Address, s)
	}
	a.Logger.Debugf("%q generated CSR:\n%s\n", t.Config.Address, s)

	// create certificate from CSR
	certificate, err := certificateFromCSR(creq, a.Config.CertInstallValidity)
	if err != nil {
		return fmt.Errorf("failed certificateFromCSR: %v", err)
	}
	// sign certificate
	a.Logger.Infof("%q signing certificate %q with the provided CA", t.Config.Address, certificate.Subject.String())
	signedCert, err := a.sign(certificate, &caCert)
	if err != nil {
		return fmt.Errorf("%q failed signing certificate: %v", t.Config.Address, err)
	}
	//
	sCertText, err := CertificateText(signedCert, false)
	if err != nil {
		return err
	}
	a.Logger.Debugf("%q signed certificate:\n%s\n", t.Config.Address, sCertText)
	// encode signed certifcate in PEM format
	b, err := toPEM(signedCert)
	if err != nil {
		return fmt.Errorf("%q failed to encode as PEM: %v", t.Config.Address, err)
	}
	a.Logger.Infof("%q installing certificate id=%s %q", t.Config.Address, a.Config.CertInstallCertificateID, certificate.Subject.String())

	// install certificate load certificate request options
	opts := []gcert.CertOption{
		gcert.Certificate(
			gcert.CertificateType(a.Config.CertInstallCertificateType),
			gcert.CertificateBytes(b),
		),
	}
	if a.Config.CertInstallGenCSR {
		// if the csr was generated locally, add the key pair and cert ID
		opts = append(opts,
			gcert.KeyPair(keyPair.GetPublicKey(), keyPair.GetPrivateKey()),
			gcert.CertificateID(a.Config.CertInstallCertificateID),
		)
	}
	loadCertReq, err := gcert.NewCertInstallLoadCertificateRequest(opts...)
	if err != nil {
		return err
	}
	err = stream.Send(loadCertReq)
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

func (a *App) createLocalCSRInstall(t *api.Target) (*cert.KeyPair, *x509.CertificateRequest, error) {
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

	var subj pkix.Name
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
	if ipAddr != "" {
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
			PrivateKey: pem.EncodeToMemory(&pem.Block{
				Type:  "RSA PRIVATE KEY",
				Bytes: x509.MarshalPKCS1PrivateKey(privateKey),
			}),
			PublicKey: csrBytes,
		},
		creq, err
}

func (a *App) createRemoteCSRInstall(stream cert.CertificateManagement_InstallClient, t *api.Target) (*x509.CertificateRequest, error) {
	var commonName = a.Config.CertInstallCommonName
	var ipAddr = a.Config.CertInstallIPAddress
	if commonName == "" {
		commonName = t.Config.CommonName
	}
	if ipAddr == "" {
		ipAddr = t.Config.ResolvedIP
	}
	req, err := gcert.NewCertInstallGenerateCSRRequest(
		gcert.CertificateID(a.Config.CertInstallCertificateID),
		gcert.CSRParams(
			gcert.CertificateType(a.Config.CertInstallCertificateType),
			gcert.MinKeySize(a.Config.CertInstallMinKeySize),
			gcert.KeyType(a.Config.CertInstallKeyType),
			gcert.CommonName(commonName),
			gcert.Country(a.Config.CertInstallCountry),
			gcert.State(a.Config.CertInstallState),
			gcert.City(a.Config.CertInstallCity),
			gcert.Org(a.Config.CertInstallOrg),
			gcert.OrgUnit(a.Config.CertInstallOrgUnit),
			gcert.IPAddress(ipAddr),
			gcert.EmailID(a.Config.CertInstallEmailID),
		),
	)
	if err != nil {
		return nil, err
	}
	a.printMsg(t.Config.Name, req)

	err = stream.Send(req)
	if err != nil {
		return nil, fmt.Errorf("%q failed send Install RPC: GenCSR: %v", err, t.Config.Address)
	}
	resp, err := stream.Recv()
	if err != nil {
		return nil, fmt.Errorf("%q failed rcv Install RPC: GenCSR: %v", err, t.Config.Address)
	}
	if resp == nil {
		return nil, fmt.Errorf("%q returned a <nil> CSR response", t.Config.Address)
	}
	if !a.Config.CertInstallPrintCSR {
		a.printMsg(t.Config.Name, resp)
	}
	if a.Config.CertInstallPrintCSR {
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
