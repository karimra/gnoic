package config

import (
	"crypto/tls"
	"errors"
	"time"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

type TargetConfig struct {
	Address    string
	Insecure   *bool
	SkipVerify *bool
	Username   *string
	Password   *string
	Timeout    time.Duration
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
		targetsConfigs[addr] = tc
	}
	return targetsConfigs, nil
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

func (tc *TargetConfig) newTLS() (*tls.Config, error) { return nil, nil }
