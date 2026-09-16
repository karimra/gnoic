package config

import (
	"crypto/ecdsa"
	"crypto/elliptic"
	"crypto/rand"
	"crypto/tls"
	"crypto/x509"
	"crypto/x509/pkix"
	"encoding/pem"
	"math/big"
	"os"
	"path/filepath"
	"testing"
	"time"
)

func TestParseAddress(t *testing.T) {
	c := newTestConfig()
	c.Port = "57400"

	tests := []struct {
		name        string
		addr        string
		wantAddress string
		wantIP      string
		wantErr     bool
	}{
		{name: "ipv4 without port", addr: "127.0.0.1", wantAddress: "127.0.0.1:57400", wantIP: "127.0.0.1"},
		{name: "ipv4 with port", addr: "127.0.0.1:9339", wantAddress: "127.0.0.1:9339", wantIP: "127.0.0.1"},
		{name: "ipv6 without port", addr: "::1", wantAddress: "[::1]:57400", wantIP: "::1"},
		{name: "ipv6 with port", addr: "[::1]:9339", wantAddress: "[::1]:9339", wantIP: "::1"},
		{name: "hostname with port", addr: "localhost:9339", wantAddress: "localhost:9339"},
		{name: "unbalanced bracket", addr: "[::1", wantErr: true},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			tc := new(TargetConfig)
			err := c.parseAddress(tc, tt.addr)
			if tt.wantErr {
				if err == nil {
					t.Fatalf("parseAddress(%q) expected error", tt.addr)
				}
				return
			}
			if err != nil {
				t.Fatalf("parseAddress(%q) error: %v", tt.addr, err)
			}
			if tc.Address != tt.wantAddress {
				t.Errorf("Address = %q, want %q", tc.Address, tt.wantAddress)
			}
			if tt.wantIP != "" && tc.ResolvedIP != tt.wantIP {
				t.Errorf("ResolvedIP = %q, want %q", tc.ResolvedIP, tt.wantIP)
			}
		})
	}
	// hostnames become the common name
	tc := new(TargetConfig)
	if err := c.parseAddress(tc, "localhost:9339"); err != nil {
		t.Fatal(err)
	}
	if tc.CommonName != "localhost" {
		t.Errorf("CommonName = %q, want localhost", tc.CommonName)
	}
}

func TestSetTargetConfigDefaults(t *testing.T) {
	c := newTestConfig()
	c.Username = "u"
	c.Password = "p"
	c.Timeout = 7 * time.Second
	c.TLSCa = "/ca.pem"
	c.TLSCert = "/cert.pem"
	c.TLSKey = "/key.pem"
	c.TLSVersion = "1.3"
	c.TLSMinVersion = "1.2"
	c.TLSMaxVersion = "1.3"
	c.SkipVerify = true

	t.Run("inherits globals", func(t *testing.T) {
		tc := &TargetConfig{Address: "10.0.0.1:57400"}
		c.setTargetConfigDefaults(tc)
		if tc.Name != tc.Address {
			t.Errorf("Name = %q, want address", tc.Name)
		}
		if tc.Insecure != nil {
			t.Errorf("Insecure = %v, want nil when global insecure is false", *tc.Insecure)
		}
		if tc.Timeout != 7*time.Second {
			t.Errorf("Timeout = %v, want 7s", tc.Timeout)
		}
		if tc.Username == nil || *tc.Username != "u" || tc.Password == nil || *tc.Password != "p" {
			t.Errorf("credentials not inherited: %v %v", tc.Username, tc.Password)
		}
		if tc.SkipVerify == nil || !*tc.SkipVerify {
			t.Errorf("SkipVerify not inherited")
		}
		if tc.TLSCA == nil || *tc.TLSCA != "/ca.pem" || tc.TLSCert == nil || *tc.TLSCert != "/cert.pem" || tc.TLSKey == nil || *tc.TLSKey != "/key.pem" {
			t.Errorf("TLS files not inherited: %v %v %v", tc.TLSCA, tc.TLSCert, tc.TLSKey)
		}
		if tc.TLSVersion != "1.3" || tc.TLSMinVersion != "1.2" || tc.TLSMaxVersion != "1.3" {
			t.Errorf("TLS versions not inherited")
		}
		if tc.Gzip == nil || *tc.Gzip {
			t.Errorf("Gzip = %v, want false", tc.Gzip)
		}
	})

	t.Run("per-target values win", func(t *testing.T) {
		tc := &TargetConfig{
			Name:       "r1",
			Address:    "10.0.0.1:57400",
			Username:   ptr("tu"),
			Timeout:    time.Second,
			TLSCA:      ptr("/other-ca.pem"),
			TLSVersion: "1.2",
		}
		c.setTargetConfigDefaults(tc)
		if tc.Name != "r1" || *tc.Username != "tu" || tc.Timeout != time.Second || *tc.TLSCA != "/other-ca.pem" || tc.TLSVersion != "1.2" {
			t.Errorf("per-target values overwritten: %s", tc)
		}
	})

	t.Run("insecure target skips TLS files", func(t *testing.T) {
		tc := &TargetConfig{Address: "10.0.0.1:57400", Insecure: ptr(true)}
		c.setTargetConfigDefaults(tc)
		if tc.TLSCA != nil || tc.TLSCert != nil || tc.TLSKey != nil {
			t.Errorf("TLS files set on insecure target: %s", tc)
		}
	})

	t.Run("global insecure applies", func(t *testing.T) {
		ic := newTestConfig()
		ic.Insecure = true
		tc := &TargetConfig{Address: "10.0.0.1:57400"}
		ic.setTargetConfigDefaults(tc)
		if tc.Insecure == nil || !*tc.Insecure {
			t.Errorf("Insecure = %v, want true", tc.Insecure)
		}
	})
}

func TestGetTargets_FromAddressFlag(t *testing.T) {
	c := newTestConfig()
	c.Port = "57400"
	c.Address = []string{"127.0.0.1", "127.0.0.2:9339"}
	c.Username = "u"

	targets, err := c.GetTargets()
	if err != nil {
		t.Fatalf("GetTargets() error: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("got %d targets, want 2", len(targets))
	}
	tc, ok := targets["127.0.0.1:57400"]
	if !ok {
		t.Fatalf("target 127.0.0.1:57400 not found in %v", targets)
	}
	if tc.Username == nil || *tc.Username != "u" {
		t.Errorf("username not inherited")
	}
	if _, ok := targets["127.0.0.2:9339"]; !ok {
		t.Errorf("target 127.0.0.2:9339 not found in %v", targets)
	}
}

func TestGetTargets_FromFile(t *testing.T) {
	c := newTestConfig()
	c.Port = "57400"
	c.Username = "global"
	c.FileConfig.Set("targets", map[string]interface{}{
		"127.0.0.1:9339": map[string]interface{}{
			"username": "local",
			"timeout":  "5s",
			"insecure": true,
		},
		"127.0.0.2": nil,
	})

	targets, err := c.GetTargets()
	if err != nil {
		t.Fatalf("GetTargets() error: %v", err)
	}
	if len(targets) != 2 {
		t.Fatalf("got %d targets, want 2: %v", len(targets), targets)
	}
	t1 := targets["127.0.0.1:9339"]
	if t1 == nil {
		t.Fatalf("target 127.0.0.1:9339 missing")
	}
	if *t1.Username != "local" || t1.Timeout != 5*time.Second || t1.Insecure == nil || !*t1.Insecure {
		t.Errorf("per-target config not decoded: %s", t1)
	}
	t2 := targets["127.0.0.2:57400"]
	if t2 == nil {
		t.Fatalf("target 127.0.0.2:57400 missing")
	}
	if *t2.Username != "global" {
		t.Errorf("global username not applied to bare target: %s", t2)
	}
}

func TestGetTargets_Errors(t *testing.T) {
	c := newTestConfig()
	if _, err := c.GetTargets(); err == nil {
		t.Error("expected error when no targets are configured")
	}
	c.FileConfig.Set("targets", map[string]interface{}{"127.0.0.1": "not-a-map"})
	if _, err := c.GetTargets(); err == nil {
		t.Error("expected error for malformed target entry")
	}
	c2 := newTestConfig()
	c2.Address = []string{"[::1"}
	if _, err := c2.GetTargets(); err == nil {
		t.Error("expected error for unparsable address")
	}
}

func TestTLSVersionStringToUint(t *testing.T) {
	tests := map[string]uint16{
		"1.3": tls.VersionTLS13,
		"1.2": tls.VersionTLS12,
		"1.1": tls.VersionTLS11,
		"1.0": tls.VersionTLS10,
		"1":   tls.VersionTLS10,
		"":    0,
		"2.0": 0,
	}
	for in, want := range tests {
		if got := tlsVersionStringToUint(in); got != want {
			t.Errorf("tlsVersionStringToUint(%q) = %d, want %d", in, got, want)
		}
	}
}

func TestGetTLSMinMaxVersion(t *testing.T) {
	tc := &TargetConfig{TLSMinVersion: "1.2", TLSMaxVersion: "1.3"}
	if tc.getTLSMinVersion() != tls.VersionTLS12 || tc.getTLSMaxVersion() != tls.VersionTLS13 {
		t.Errorf("min/max = %d/%d", tc.getTLSMinVersion(), tc.getTLSMaxVersion())
	}
	tc.TLSVersion = "1.2"
	if tc.getTLSMinVersion() != tls.VersionTLS12 || tc.getTLSMaxVersion() != tls.VersionTLS12 {
		t.Errorf("TLSVersion should override min/max, got %d/%d", tc.getTLSMinVersion(), tc.getTLSMaxVersion())
	}
}

func TestDialOpts_Insecure(t *testing.T) {
	tc := &TargetConfig{Insecure: ptr(true)}
	opts, err := tc.DialOpts()
	if err != nil {
		t.Fatal(err)
	}
	if len(opts) != 1 {
		t.Fatalf("got %d dial options, want 1", len(opts))
	}
}

func TestNewTLS(t *testing.T) {
	certFile, keyFile := writeTestCertificate(t)

	t.Run("skip verify, no files", func(t *testing.T) {
		tc := &TargetConfig{SkipVerify: ptr(true)}
		cfg, err := tc.newTLS()
		if err != nil {
			t.Fatal(err)
		}
		if !cfg.InsecureSkipVerify {
			t.Error("InsecureSkipVerify not set")
		}
		if len(cfg.Certificates) != 0 || cfg.RootCAs != nil {
			t.Error("unexpected certificates loaded")
		}
		// TLS 1.3 suites are appended when max version is unset or 1.3
		if len(cfg.CipherSuites) != len(defaultCipherSuitesTLS12)+len(defaultCipherSuitesTLS13) {
			t.Errorf("got %d cipher suites", len(cfg.CipherSuites))
		}
	})

	t.Run("max 1.2 restricts cipher suites", func(t *testing.T) {
		tc := &TargetConfig{SkipVerify: ptr(true), TLSMaxVersion: "1.2"}
		cfg, err := tc.newTLS()
		if err != nil {
			t.Fatal(err)
		}
		if len(cfg.CipherSuites) != len(defaultCipherSuitesTLS12) {
			t.Errorf("got %d cipher suites, want %d", len(cfg.CipherSuites), len(defaultCipherSuitesTLS12))
		}
		if cfg.MaxVersion != tls.VersionTLS12 {
			t.Errorf("MaxVersion = %d", cfg.MaxVersion)
		}
	})

	t.Run("client cert and CA", func(t *testing.T) {
		tc := &TargetConfig{
			SkipVerify: ptr(false),
			TLSCert:    ptr(certFile),
			TLSKey:     ptr(keyFile),
			TLSCA:      ptr(certFile),
		}
		cfg, err := tc.newTLS()
		if err != nil {
			t.Fatal(err)
		}
		if len(cfg.Certificates) != 1 {
			t.Errorf("got %d client certificates, want 1", len(cfg.Certificates))
		}
		if cfg.RootCAs == nil {
			t.Error("RootCAs not set")
		}
		opts, err := tc.DialOpts()
		if err != nil || len(opts) != 1 {
			t.Errorf("DialOpts() = %v, %v", opts, err)
		}
	})

	t.Run("bad key pair", func(t *testing.T) {
		tc := &TargetConfig{SkipVerify: ptr(false), TLSCert: ptr(certFile), TLSKey: ptr(certFile)}
		if _, err := tc.newTLS(); err == nil {
			t.Error("expected error for mismatched cert/key")
		}
	})

	t.Run("missing CA file", func(t *testing.T) {
		tc := &TargetConfig{SkipVerify: ptr(false), TLSCA: ptr(filepath.Join(t.TempDir(), "nope.pem"))}
		if _, err := tc.newTLS(); err == nil {
			t.Error("expected error for missing CA file")
		}
	})

	t.Run("CA file without certificates", func(t *testing.T) {
		bad := filepath.Join(t.TempDir(), "bad.pem")
		if err := os.WriteFile(bad, []byte("not a pem"), 0600); err != nil {
			t.Fatal(err)
		}
		tc := &TargetConfig{SkipVerify: ptr(false), TLSCA: ptr(bad)}
		if _, err := tc.newTLS(); err == nil {
			t.Error("expected error for CA file without certificates")
		}
	})

	t.Run("preset tls config is reused", func(t *testing.T) {
		preset := &tls.Config{ServerName: "preset"}
		tc := &TargetConfig{}
		tc.SetTLSConfig(preset)
		cfg, err := tc.newTLS()
		if err != nil {
			t.Fatal(err)
		}
		if cfg != preset {
			t.Error("preset tls.Config not returned")
		}
	})
}

func TestTargetConfigString(t *testing.T) {
	tc := &TargetConfig{Name: "r1", Address: "10.0.0.1:57400", Password: ptr("secret")}
	s := tc.String()
	if s == "" {
		t.Fatal("String() returned empty")
	}
	// String() is used in debug logs; make sure it is JSON with the address.
	if want := `"address":"10.0.0.1:57400"`; !contains(s, want) {
		t.Errorf("String() = %s, want it to contain %s", s, want)
	}
}

func contains(s, sub string) bool {
	return len(s) >= len(sub) && (func() bool {
		for i := 0; i+len(sub) <= len(s); i++ {
			if s[i:i+len(sub)] == sub {
				return true
			}
		}
		return false
	})()
}

// writeTestCertificate writes a self-signed ECDSA certificate and key to a
// temp dir and returns their paths.
func writeTestCertificate(t *testing.T) (certFile, keyFile string) {
	t.Helper()
	key, err := ecdsa.GenerateKey(elliptic.P256(), rand.Reader)
	if err != nil {
		t.Fatal(err)
	}
	tmpl := &x509.Certificate{
		SerialNumber:          big.NewInt(1),
		Subject:               pkix.Name{CommonName: "gnoic-test"},
		NotBefore:             time.Now().Add(-time.Hour),
		NotAfter:              time.Now().Add(time.Hour),
		IsCA:                  true,
		BasicConstraintsValid: true,
		KeyUsage:              x509.KeyUsageDigitalSignature | x509.KeyUsageCertSign,
	}
	der, err := x509.CreateCertificate(rand.Reader, tmpl, tmpl, &key.PublicKey, key)
	if err != nil {
		t.Fatal(err)
	}
	keyDER, err := x509.MarshalECPrivateKey(key)
	if err != nil {
		t.Fatal(err)
	}
	dir := t.TempDir()
	certFile = filepath.Join(dir, "cert.pem")
	keyFile = filepath.Join(dir, "key.pem")
	if err := os.WriteFile(certFile, pem.EncodeToMemory(&pem.Block{Type: "CERTIFICATE", Bytes: der}), 0600); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(keyFile, pem.EncodeToMemory(&pem.Block{Type: "EC PRIVATE KEY", Bytes: keyDER}), 0600); err != nil {
		t.Fatal(err)
	}
	return certFile, keyFile
}
