package config

import (
	"crypto/tls"
	"crypto/x509"
	"encoding/json"
	"errors"
	"fmt"
	"io/ioutil"
	"net"
	"strings"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type TargetConfig struct {
	Address       string        `json:"address,omitempty"`
	Insecure      *bool         `json:"insecure,omitempty"`
	SkipVerify    *bool         `json:"skip-verify,omitempty"`
	Username      *string       `json:"username,omitempty"`
	Password      *string       `json:"password,omitempty"`
	Timeout       time.Duration `json:"timeout,omitempty"`
	TLSCert       *string       `json:"tls-cert,omitempty"`
	TLSKey        *string       `json:"tls-key,omitempty"`
	TLSCA         *string       `json:"tlsca,omitempty"`
	TLSMinVersion string        `json:"tls-min-version,omitempty"`
	TLSMaxVersion string        `json:"tls-max-version,omitempty"`
	TLSVersion    string        `json:"tls-version,omitempty"`
	Gzip          *bool         `json:"gzip,omitempty"`
	//
	CommonName string `json:"common-name,omitempty"`
	ResolvedIP string `json:"resolved-ip,omitempty"`
}

func (c *Config) GetTargets() (map[string]*TargetConfig, error) {
	if len(c.Address) == 0 {
		return nil, errors.New("no targets found")
	}
	targetsConfigs := make(map[string]*TargetConfig)
	var err error
	for _, addr := range c.Address {
		tc := new(TargetConfig)
		err = c.parseAddress(tc, addr)
		if err != nil {
			return nil, fmt.Errorf("%q failed to parse address: %v", addr, err)
		}
		c.setTargetConfigDefaults(tc)
		targetsConfigs[tc.Address] = tc
		c.logger.Debugf("%q target-config: %s", addr, tc)
	}
	return targetsConfigs, nil
}

func (c *Config) parseAddress(tc *TargetConfig, addr string) error {
	h, _, err := net.SplitHostPort(addr)
	if err != nil {
		if strings.Contains(err.Error(), "missing port in address") ||
			strings.Contains(err.Error(), "too many colons in address") {
			tc.Address = net.JoinHostPort(addr, c.Port)
			h = addr
		} else {
			return fmt.Errorf("error parsing address %q: %v", addr, err)
		}
	} else {
		tc.Address = addr
	}
	// parse provided hostname/IPAddress
	ip := net.ParseIP(h)
	if ip == nil {
		// address is a hostname
		tc.CommonName = h
		resolvedIP, err := net.ResolveIPAddr("ip", h)
		if err != nil {
			c.logger.Infof("%q could not resolve %q: %v", addr, h, err)
		} else {
			tc.ResolvedIP = resolvedIP.String()
		}
	} else {
		// address is IPAddress
		tc.ResolvedIP = ip.String()
		names, err := net.LookupAddr(tc.ResolvedIP)
		if err != nil {
			c.logger.Warnf("%q could not lookup hostname: %v", addr, err)
		}
		c.logger.Debugf("%q resolved names: %v", addr, names)
		if len(names) > 0 {
			tc.CommonName = names[0]
		}
	}
	return nil
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
	if tc.Insecure == nil || (tc.Insecure != nil && !*tc.Insecure) {
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

func (tc *TargetConfig) String() string {
	b, err := json.Marshal(tc)
	if err != nil {
		return ""
	}
	return string(b)
}
