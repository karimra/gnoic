package config

import (
	"crypto/tls"
	"crypto/x509"
	"errors"
	"io/ioutil"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type TargetConfig struct {
	Address       string
	Insecure      *bool
	SkipVerify    *bool
	Username      *string
	Password      *string
	Timeout       time.Duration
	TLSCert       *string
	TLSKey        *string
	TLSCA         *string
	TLSMinVersion string
	TLSMaxVersion string
	TLSVersion    string
	//
	Gzip *bool
}

func (c *Config) GetTargets() (map[string]*TargetConfig, error) {
	if len(c.Address) == 0 {
		return nil, errors.New("no targets found")
	}
	targetsConfigs := make(map[string]*TargetConfig)
	for _, addr := range c.Address {
		tc := new(TargetConfig)
		tc.Address = addr
		c.setTargetConfigDefaults(tc)
		targetsConfigs[addr] = tc
	}
	return targetsConfigs, nil
}

func (c *Config) setTargetConfigDefaults(tc *TargetConfig) {
	if c.Insecure {
		tc.Insecure = &c.Insecure
	}
	if tc.Timeout <= 0 {
		tc.Timeout = c.Timeout
	}
	if tc.Username == nil {
		tc.Username = &c.Username
	}
	if tc.Password == nil {
		tc.Password = &c.Password
	}
	if tc.SkipVerify == nil {
		tc.SkipVerify = &c.SkipVerify
	}
	if tc.Timeout == 0 {
		tc.Timeout = c.Timeout
	}
	if tc.Insecure != nil && !*tc.Insecure {
		if tc.TLSCA == nil {
			if c.TLSCa != "" {
				tc.TLSCA = &c.TLSCa
			}
		}
		if tc.TLSCert == nil {
			tc.TLSCert = &c.TLSCert
		}
		if tc.TLSKey == nil {
			tc.TLSKey = &c.TLSKey
		}
	}
	if tc.TLSVersion == "" {
		tc.TLSVersion = c.TLSVersion
	}
	if tc.TLSMinVersion == "" {
		tc.TLSMinVersion = c.TLSMinVersion
	}
	if tc.TLSMaxVersion == "" {
		tc.TLSMaxVersion = c.TLSMaxVersion
	}
	if tc.Gzip == nil {
		tc.Gzip = &c.Gzip
	}
}

func (tc *TargetConfig) DialOpts() ([]grpc.DialOption, error) {
	tOpts := make([]grpc.DialOption, 0)
	if tc.Insecure != nil && *tc.Insecure {
		tOpts = append(tOpts, grpc.WithInsecure())
	} else {
		tlsConfig, err := tc.newTLS()
		if err != nil {
			return nil, err
		}
		tOpts = append(tOpts, grpc.WithTransportCredentials(credentials.NewTLS(tlsConfig)))
	}
	return tOpts, nil
}

func (tc *TargetConfig) newTLS() (*tls.Config, error) {
	tlsConfig := &tls.Config{
		Renegotiation:      tls.RenegotiateNever,
		InsecureSkipVerify: *tc.SkipVerify,
		MaxVersion:         tc.getTLSMaxVersion(),
		MinVersion:         tc.getTLSMinVersion(),
	}
	err := loadCerts(tlsConfig, tc)
	if err != nil {
		return nil, err
	}
	return tlsConfig, nil
}

func (tc *TargetConfig) getTLSMinVersion() uint16 {
	v := tlsVersionStringToUint(tc.TLSVersion)
	if v > 0 {
		return v
	}
	return tlsVersionStringToUint(tc.TLSMinVersion)
}

func (tc *TargetConfig) getTLSMaxVersion() uint16 {
	v := tlsVersionStringToUint(tc.TLSVersion)
	if v > 0 {
		return v
	}
	return tlsVersionStringToUint(tc.TLSMaxVersion)
}

func tlsVersionStringToUint(v string) uint16 {
	switch v {
	default:
		return 0
	case "1.3":
		return tls.VersionTLS13
	case "1.2":
		return tls.VersionTLS12
	case "1.1":
		return tls.VersionTLS11
	case "1.0", "1":
		return tls.VersionTLS10
	}
}

func loadCerts(tlscfg *tls.Config, tc *TargetConfig) error {
	if *tc.TLSCert != "" && *tc.TLSKey != "" {
		certificate, err := tls.LoadX509KeyPair(*tc.TLSCert, *tc.TLSKey)
		if err != nil {
			return err
		}
		tlscfg.Certificates = []tls.Certificate{certificate}
		tlscfg.BuildNameToCertificate()
	}
	if tc.TLSCA != nil && *tc.TLSCA != "" {
		certPool := x509.NewCertPool()
		caFile, err := ioutil.ReadFile(*tc.TLSCA)
		if err != nil {
			return err
		}
		if ok := certPool.AppendCertsFromPEM(caFile); !ok {
			return errors.New("failed to append certificate")
		}
		tlscfg.RootCAs = certPool
	}
	return nil
}
