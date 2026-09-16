package app

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/rsa"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"net"
	"strings"
	"testing"
	"time"
)

// newTestCA generates an RSA CA usable with App.sign.
func newTestCA(t *testing.T) *tls.Certificate {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	sn, err := genSerialNumber()
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          sn,
		Subject:               pkix.Name{CommonName: "gnoic test CA", Organization: []string{"gNOIc"}},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(24 * time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	leaf, err := x509.ParseCertificate(der)
	if err != nil {
		t.Fatal(err)
	}
	return &tls.Certificate{Certificate: [][]byte{der}, PrivateKey: key, Leaf: leaf}
}

func newTestCSR(t *testing.T) *x509.CertificateRequest {
	t.Helper()
	key, err := rsa.GenerateKey(rand.Reader, 2048)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.CertificateRequest{
		Subject:            pkix.Name{CommonName: "router1", Country: []string{"US"}},
		DNSNames:           []string{"router1"},
		IPAddresses:        []net.IP{net.ParseIP("10.0.0.1")},
		EmailAddresses:     []string{"noc@example.com"},
		SignatureAlgorithm: x509.SHA256WithRSA,
	}
	der, err := x509.CreateCertificateRequest(rand.Reader, tmpl, key)
	if err != nil {
		t.Fatal(err)
	}
	csr, err := x509.ParseCertificateRequest(der)
	if err != nil {
		t.Fatal(err)
	}
	return csr
}

func TestGenSerialNumber(t *testing.T) {
	a, err := genSerialNumber()
	if err != nil {
		t.Fatal(err)
	}
	b, err := genSerialNumber()
	if err != nil {
		t.Fatal(err)
	}
	if a.Sign() < 0 || b.Sign() < 0 || a.BitLen() > 128 {
		t.Errorf("serial numbers out of range: %v %v", a, b)
	}
	if a.Cmp(b) == 0 {
		t.Errorf("two generated serial numbers are equal: %v", a)
	}
}

func TestKeyID(t *testing.T) {
	rsaKey, err := rsa.GenerateKey(rand.Reader, 1024)
	if err != nil {
		t.Fatal(err)
	}
	id, err := keyID(&rsaKey.PublicKey)
	if err != nil {
		t.Fatal(err)
	}
	if len(id) != 32 {
		t.Errorf("keyID length = %d, want 32 (sha256)", len(id))
	}
	id2, _ := keyID(&rsaKey.PublicKey)
	if string(id) != string(id2) {
		t.Error("keyID is not deterministic")
	}
	ecKey, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := keyID(&ecKey.PublicKey); err == nil {
		t.Error("expected error for non-RSA key")
	}
}

func TestCertificateFromCSRSignAndPEM(t *testing.T) {
	a := newTestApp(t)
	ca := newTestCA(t)
	csr := newTestCSR(t)

	c, err := certificateFromCSR(csr, 48*time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if c.Subject.CommonName != "router1" || len(c.DNSNames) != 1 || len(c.IPAddresses) != 1 || len(c.EmailAddresses) != 1 {
		t.Errorf("CSR fields not copied: %v", c.Subject)
	}
	if len(c.SubjectKeyId) != 32 {
		t.Errorf("SubjectKeyId length = %d", len(c.SubjectKeyId))
	}
	if !c.NotAfter.After(time.Now().Add(47 * time.Hour)) {
		t.Errorf("NotAfter = %v, want ~48h from now", c.NotAfter)
	}

	signed, err := a.sign(c, ca)
	if err != nil {
		t.Fatal(err)
	}
	if signed.Issuer.CommonName != "gnoic test CA" {
		t.Errorf("Issuer = %v", signed.Issuer)
	}
	roots := x509.NewCertPool()
	roots.AddCert(ca.Leaf)
	if _, err := signed.Verify(x509.VerifyOptions{
		Roots:     roots,
		DNSName:   "router1",
		KeyUsages: []x509.ExtKeyUsage{x509.ExtKeyUsageServerAuth},
	}); err != nil {
		t.Errorf("signed certificate does not verify against CA: %v", err)
	}

	pemBytes, err := toPEM(signed)
	if err != nil {
		t.Fatal(err)
	}
	block, rest := pem.Decode(pemBytes)
	if block == nil || block.Type != "CERTIFICATE" || len(rest) != 0 {
		t.Fatalf("toPEM produced invalid PEM: %s", pemBytes)
	}
	if _, err := x509.ParseCertificate(block.Bytes); err != nil {
		t.Errorf("PEM does not contain a parsable certificate: %v", err)
	}
}

func TestSign_MissingCA(t *testing.T) {
	a := newTestApp(t)
	csr := newTestCSR(t)
	c, err := certificateFromCSR(csr, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	if _, err := a.sign(c, &tls.Certificate{}); err == nil {
		t.Error("expected error when signing with an empty CA")
	}
}

func TestCertificateText(t *testing.T) {
	a := newTestApp(t)
	ca := newTestCA(t)
	csr := newTestCSR(t)
	c, err := certificateFromCSR(csr, time.Hour)
	if err != nil {
		t.Fatal(err)
	}
	signed, err := a.sign(c, ca)
	if err != nil {
		t.Fatal(err)
	}

	txt, err := CertificateText(signed, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{
		"Certificate:",
		"Version: 3",
		"Serial Number:",
		"Signature Algorithm: SHA256-RSA",
		"Issuer: CN=gnoic test CA",
		"Subject: CN=router1",
		"Public Key Algorithm: RSA",
		"Public-Key: (2048 bit)",
		"X509v3 Subject Alternative Name",
		"DNS:router1",
		"IP Address:10.0.0.1",
		"X509v3 Basic Constraints",
		"TLS Web Server Authentication",
		"TLS Web Client Authentication",
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("CertificateText missing %q", want)
		}
	}
	if strings.Contains(txt, "-----BEGIN CERTIFICATE-----") {
		t.Error("PEM printed although printPEM=false")
	}
	withPEM, err := CertificateText(signed, true)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(withPEM, "-----BEGIN CERTIFICATE-----") {
		t.Error("PEM missing although printPEM=true")
	}

	// CA certificate: no SANs, CA:TRUE
	caTxt, err := CertificateText(ca.Leaf, false)
	if err != nil {
		t.Fatal(err)
	}
	if !strings.Contains(caTxt, "CA:TRUE") {
		t.Errorf("CA text missing CA:TRUE:\n%s", caTxt)
	}
}

func TestCertificateText_ECDSA(t *testing.T) {
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber: mustSerial(t),
		Subject:      pkix.Name{CommonName: "ec"},
		NotBefore:    time.Now(),
		NotAfter:     time.Now().Add(time.Hour),
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	c, _ := x509.ParseCertificate(der)
	txt, err := CertificateText(c, false)
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Public Key Algorithm: ECDSA", "Curve: P-256", "Signature Algorithm: ECDSA-SHA256"} {
		if !strings.Contains(txt, want) {
			t.Errorf("ECDSA CertificateText missing %q:\n%s", want, txt)
		}
	}
}

func TestCertificateRequestText(t *testing.T) {
	csr := newTestCSR(t)
	txt, err := CertificateRequestText(csr)
	if err != nil {
		t.Fatal(err)
	}
	// NOTE: requested extensions (SANs) are only printed when csr.Version == 3,
	// but PKCS#10 requests are always version 0, so they never appear today
	// (bug, fixed in a later phase).
	for _, want := range []string{
		"Certificate Request:",
		"Version: 0 (0x0)",
		"Subject: CN=router1",
		"Public Key Algorithm: RSA",
		"Public-Key: (2048 bit)",
		"Signature Algorithm: SHA256-RSA",
		"-----BEGIN CERTIFICATE REQUEST-----",
	} {
		if !strings.Contains(txt, want) {
			t.Errorf("CertificateRequestText missing %q:\n%s", want, txt)
		}
	}
}

func TestPrintVersion(t *testing.T) {
	var sb strings.Builder
	printVersion(3, &sb)
	if !strings.Contains(sb.String(), "Version: 3 (0x2)") {
		t.Errorf("printVersion(3) = %q", sb.String())
	}
	sb.Reset()
	printVersion(0, &sb)
	if !strings.Contains(sb.String(), "Version: 0 (0x0)") {
		t.Errorf("printVersion(0) = %q", sb.String())
	}
}

func mustSerial(t *testing.T) *big.Int {
	t.Helper()
	sn, err := genSerialNumber()
	if err != nil {
		t.Fatal(err)
	}
	return sn
}
