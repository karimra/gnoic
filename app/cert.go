package app

import (
	"bytes"
	"crypto"
	"crypto/rand"
	"crypto/rsa"
	"crypto/sha256"
	"crypto/tls"
	"crypto/x509"
	"encoding/asn1"
	"encoding/pem"
	"fmt"
	"math/big"
	"time"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

var caCert tls.Certificate

func (a *App) InitCertFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.PersistentFlags().StringVar(&a.Config.CertCAKey, "ca-key", "", "CA key")
	cmd.PersistentFlags().StringVar(&a.Config.CertCACert, "ca-cert", "", "CA Certificate")
	//
	cmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func genSerialNumber() (*big.Int, error) {
	serialNumberLimit := new(big.Int).Lsh(big.NewInt(1), 128)
	return rand.Int(rand.Reader, serialNumberLimit)
}

func certificateFromCSR(csr *x509.CertificateRequest, certExpiration time.Duration) (*x509.Certificate, error) {
	sn, err := genSerialNumber()
	if err != nil {
		return nil, err
	}
	certificate := &x509.Certificate{
		SerialNumber:          sn,
		BasicConstraintsValid: true,
		DNSNames:              csr.DNSNames,
		ExtKeyUsage:           []x509.ExtKeyUsage{x509.ExtKeyUsageClientAuth, x509.ExtKeyUsageServerAuth},
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageKeyEncipherment,
		NotAfter:              time.Now().Add(certExpiration),
		NotBefore:             time.Now().Add(-1 * time.Hour),
		SignatureAlgorithm:    csr.SignatureAlgorithm,
		Subject:               csr.Subject,
		Signature:             csr.Signature,
		Extensions:            csr.Extensions,
		Version:               csr.Version,
		ExtraExtensions:       csr.ExtraExtensions,
		EmailAddresses:        csr.EmailAddresses,
		IPAddresses:           csr.IPAddresses,
		URIs:                  csr.URIs,
		PublicKeyAlgorithm:    csr.PublicKeyAlgorithm,
		PublicKey:             csr.PublicKey,
	}
	certificate.SubjectKeyId, err = keyID(csr.PublicKey)
	return certificate, err
}

func keyID(pub crypto.PublicKey) ([]byte, error) {
	pk, ok := pub.(*rsa.PublicKey)
	if !ok {
		return nil, fmt.Errorf("failed to parse public key, not a rsa.PublicKey type")
	}
	pkBytes, err := asn1.Marshal(*pk)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal public key: %v", err)
	}
	subjectKeyID := sha256.Sum256(pkBytes)
	return subjectKeyID[:], nil
}

func (a *App) sign(c *x509.Certificate, ca *tls.Certificate) (*x509.Certificate, error) {
	derCert, err := x509.CreateCertificate(rand.Reader, c, ca.Leaf, c.PublicKey, ca.PrivateKey)
	if err != nil {
		return nil, err
	}
	return x509.ParseCertificate(derCert)
}

func toPEM(c *x509.Certificate) ([]byte, error) {
	b := new(bytes.Buffer)
	err := pem.Encode(b, &pem.Block{
		Type:  "CERTIFICATE",
		Bytes: c.Raw,
	})
	return b.Bytes(), err
}
