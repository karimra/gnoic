package app

import (
	"crypto/rand"
	"crypto/rsa"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"os"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

func (a *App) InitCertCreateCaFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.CertCreateCaOrg, "org", "", "organization name")
	cmd.Flags().StringVar(&a.Config.CertCreateCaOrgUnit, "org-unit", "", "organizational Unit name")
	cmd.Flags().StringVar(&a.Config.CertCreateCaCountry, "country", "", "country name")
	cmd.Flags().StringVar(&a.Config.CertCreateCaState, "state", "", "state name")
	cmd.Flags().StringVar(&a.Config.CertCreateCaLocality, "locality", "", "locality name")
	cmd.Flags().StringVar(&a.Config.CertCreateCaStreetAddress, "street-address", "", "street-address")
	cmd.Flags().StringVar(&a.Config.CertCreateCaPostalCode, "postal-code", "", "postal-code")
	cmd.Flags().DurationVar(&a.Config.CertCreateCaValidity, "validity", 87600*time.Hour, "certificate validity")
	cmd.Flags().IntVar(&a.Config.CertCreateCaKeySize, "key-size", 2048, "key size")
	cmd.Flags().StringVar(&a.Config.CertCreateCaEmailID, "email", "", "email ID")
	cmd.Flags().StringVar(&a.Config.CertCreateCaCommonName, "common-name", "", "common name")
	cmd.Flags().StringVar(&a.Config.CertCreateCaKeyOut, "key-out", "key.pem", "private key output path")
	cmd.Flags().StringVar(&a.Config.CertCreateCaCertOut, "cert-out", "cert.pem", "CA certificate output path")
	//
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunECertCreateCa(cmd *cobra.Command, args []string) error {
	serialNumber, err := genSerialNumber()
	if err != nil {
		return err
	}
	ca := &x509.Certificate{
		SerialNumber: serialNumber,
		Subject: pkix.Name{
			Country:            []string{a.Config.CertCreateCaCountry},
			Organization:       []string{a.Config.CertCreateCaOrg},
			OrganizationalUnit: []string{a.Config.CertCreateCaOrgUnit},
			Province:           []string{a.Config.CertCreateCaState},
			Locality:           []string{a.Config.CertCreateCaLocality},
			StreetAddress:      []string{a.Config.CertCreateCaStreetAddress},
			PostalCode:         []string{a.Config.CertCreateCaPostalCode},
			CommonName:         a.Config.CertCreateCaCommonName,
		},
		NotBefore:             time.Now(),
		NotAfter:              time.Now().Add(a.Config.CertCreateCaValidity),
		IsCA:                  true,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
		BasicConstraintsValid: true,
	}
	if a.Config.CertCreateCaEmailID != "" {
		ca.Subject.ExtraNames = append(ca.Subject.ExtraNames, pkix.AttributeTypeAndValue{
			Type: oidEmailAddress,
			Value: asn1.RawValue{
				Tag:   asn1.TagIA5String,
				Bytes: []byte(a.Config.CertCreateCaEmailID),
			},
		})
	}

	caPrivKey, err := rsa.GenerateKey(rand.Reader, a.Config.CertCreateCaKeySize)
	if err != nil {
		return err
	}

	caBytes, err := x509.CreateCertificate(rand.Reader, ca, ca, &caPrivKey.PublicKey, caPrivKey)
	if err != nil {
		return err
	}
	if a.Config.Debug {
		// parse for printing
		nca, err := x509.ParseCertificate(caBytes)
		if err != nil {
			return err
		}
		s, err := CertificateText(nca, false)
		if err != nil {
			return err
		}
		fmt.Printf("%s\n", s)
	}
	//
	certOut, err := os.Create(a.Config.CertCreateCaCertOut)
	if err != nil {
		return err
	}
	defer certOut.Close()
	err = pem.Encode(certOut, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: caBytes,
	})
	if err != nil {
		return err
	}

	keyOut, err := os.OpenFile(a.Config.CertCreateCaKeyOut, os.O_WRONLY|os.O_CREATE|os.O_TRUNC, 0600)
	if err != nil {
		return err
	}
	defer keyOut.Close()
	err = pem.Encode(keyOut, &pem.Block{
		Type:  "RSA PRIVATE KEY",
		Bytes: x509.MarshalPKCS1PrivateKey(caPrivKey),
	})
	if err != nil {
		return err
	}
	a.Logger.Infof("CA certificate written to %s", a.Config.CertCreateCaCertOut)
	a.Logger.Infof("CA key written to %s", a.Config.CertCreateCaKeyOut)
	return nil
}
