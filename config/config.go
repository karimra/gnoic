package config

import (
	"io"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

const (
	configName      = ".gnoic"
	configLogPrefix = "[config] "
	envPrefix       = "GNOIC"
)

type Config struct {
	GlobalFlags `mapstructure:",squash"`
	LocalFlags  `mapstructure:",squash"`
	FileConfig  *viper.Viper `mapstructure:"-" json:"-" yaml:"-" `

	logger *log.Entry
}

type GlobalFlags struct {
	CfgFile       string
	Address       []string      `mapstructure:"address,omitempty" json:"address,omitempty" yaml:"address,omitempty"`
	Username      string        `mapstructure:"username,omitempty" json:"username,omitempty" yaml:"username,omitempty"`
	Password      string        `mapstructure:"password,omitempty" json:"password,omitempty" yaml:"password,omitempty"`
	Port          string        `mapstructure:"port,omitempty" json:"port,omitempty" yaml:"port,omitempty"`
	Encoding      string        `mapstructure:"encoding,omitempty" json:"encoding,omitempty" yaml:"encoding,omitempty"`
	Insecure      bool          `mapstructure:"insecure,omitempty" json:"insecure,omitempty" yaml:"insecure,omitempty"`
	TLSCa         string        `mapstructure:"tls-ca,omitempty" json:"tls-ca,omitempty" yaml:"tls-ca,omitempty"`
	TLSCert       string        `mapstructure:"tls-cert,omitempty" json:"tls-cert,omitempty" yaml:"tls-cert,omitempty"`
	TLSKey        string        `mapstructure:"tls-key,omitempty" json:"tls-key,omitempty" yaml:"tls-key,omitempty"`
	TLSMinVersion string        `mapstructure:"tls-min-version,omitempty" json:"tls-min-version,omitempty" yaml:"tls-min-version,omitempty"`
	TLSMaxVersion string        `mapstructure:"tls-max-version,omitempty" json:"tls-max-version,omitempty" yaml:"tls-max-version,omitempty"`
	TLSVersion    string        `mapstructure:"tls-version,omitempty" json:"tls-version,omitempty" yaml:"tls-version,omitempty"`
	Timeout       time.Duration `mapstructure:"timeout,omitempty" json:"timeout,omitempty" yaml:"timeout,omitempty"`
	Debug         bool          `mapstructure:"debug,omitempty" json:"debug,omitempty" yaml:"debug,omitempty"`
	SkipVerify    bool          `mapstructure:"skip-verify,omitempty" json:"skip-verify,omitempty" yaml:"skip-verify,omitempty"`
	NoPrefix      bool          `mapstructure:"no-prefix,omitempty" json:"no-prefix,omitempty" yaml:"no-prefix,omitempty"`
	ProxyFromEnv  bool          `mapstructure:"proxy-from-env,omitempty" json:"proxy-from-env,omitempty" yaml:"proxy-from-env,omitempty"`
	Format        string        `mapstructure:"format,omitempty" json:"format,omitempty" yaml:"format,omitempty"`
	PrintRequest  bool          `mapstructure:"print-request,omitempty" json:"print-request,omitempty" yaml:"print-request,omitempty"`
	Retry         time.Duration `mapstructure:"retry,omitempty" json:"retry,omitempty" yaml:"retry,omitempty"`
	LogFile       string        `mapstructure:"log-file,omitempty" json:"log-file,omitempty" yaml:"log-file,omitempty"`
	Log           bool          `mapstructure:"log,omitempty" json:"log,omitempty" yaml:"log,omitempty"`
	Gzip          bool          `mapstructure:"gzip,omitempty" json:"gzip,omitempty" yaml:"gzip,omitempty"`
}

type LocalFlags struct {
	// Cert
	CertCA    string
	CertCAKey string
	// Cert CreateCA
	CertCreateCaOrg           string
	CertCreateCaOrgUnit       string
	CertCreateCaCountry       string
	CertCreateCaState         string
	CertCreateCaLocality      string
	CertCreateCaStreetAddress string
	CertCreateCaPostalCode    string
	CertCreateCaValidity      time.Duration
	CertCreateCaKeySize       int
	CertCreateCaEmailID       string
	CertCreateCaCommonName    string
	// Cert Rotate
	CertRotateCertificateID   string
	CertRotateKeyType         string
	CertRotateCertificateType string
	CertRotateMinKeySize      uint32
	CertRotateCommonName      string
	CertRotateCountry         string
	CertRotateState           string
	CertRotateCity            string
	CertRotateOrg             string
	CertRotateOrgUnit         string
	CertRotateIPAddress       string
	CertRotateEmailID         string
	CertRotateValidity        time.Duration
	CertRotatePrintCSR        bool
	// Cert Install
	CertInstallCertificateID   string
	CertInstallKeyType         string
	CertInstallCertificateType string
	CertInstallMinKeySize      uint32
	CertInstallCommonName      string
	CertInstallCountry         string
	CertInstallState           string
	CertInstallCity            string
	CertInstallOrg             string
	CertInstallOrgUnit         string
	CertInstallIPAddress       string
	CertInstallEmailID         string
	CertInstallValidity        time.Duration
	CertInstallPrintCSR        bool
	// Cert GenerateCSR
	CertGenerateCSRCertificateID   string
	CertGenerateCSRKeyType         string
	CertGenerateCSRCertificateType string
	CertGenerateCSRMinKeySize      uint32
	CertGenerateCSRCommonName      string
	CertGenerateCSRCountry         string
	CertGenerateCSRState           string
	CertGenerateCSRCity            string
	CertGenerateCSROrg             string
	CertGenerateCSROrgUnit         string
	CertGenerateCSRIPAddress       string
	CertGenerateCSREmailID         string
	// Cert CanGenerateCSR
	CertCanGenerateCSRKeyType         string
	CertCanGenerateCSRCertificateType string
	CertCanGenerateCSRKeySize         uint32
	// Cert LoadCertificate
	CertLoadCertificateCertificateID   string
	CertLoadCertificateCertificateType string
	CertLoadCertificatePrivateKey      string
	CertLoadCertificatePublicKey       string
	CertLoadCertificateCaCertificates  []string
	// Cert LoadCertificateCanBundle
	CertLoadCertificateCaBundleCaCertificates []string
	// Cert RevokeCertificate
	CertRevokeCertificatesCertificateID []string
	CertRevokeCertificatesAll           bool
	// Cert GetCertificates
	CertGetCertificatesDetails bool
}

func New() *Config {
	return &Config{
		GlobalFlags{},
		LocalFlags{},
		viper.NewWithOptions(viper.KeyDelimiter("/")),
		nil,
	}
}

func (c *Config) Load() error {
	c.FileConfig.SetEnvPrefix(envPrefix)
	c.FileConfig.SetEnvKeyReplacer(strings.NewReplacer("/", "_", "-", "_"))
	c.FileConfig.AutomaticEnv()
	if c.GlobalFlags.CfgFile != "" {
		c.FileConfig.SetConfigFile(c.GlobalFlags.CfgFile)
	} else {
		home, err := homedir.Dir()
		if err != nil {
			return err
		}
		c.FileConfig.AddConfigPath(".")
		c.FileConfig.AddConfigPath(home)
		c.FileConfig.AddConfigPath(xdg.ConfigHome)
		c.FileConfig.AddConfigPath(xdg.ConfigHome + "/gnoic")
		c.FileConfig.SetConfigName(configName)
	}

	err := c.FileConfig.ReadInConfig()
	if err != nil {
		return err
	}

	err = c.FileConfig.Unmarshal(c.FileConfig)
	if err != nil {
		return err
	}
	// c.mergeEnvVars()
	// return c.expandOSPathFlagValues()
	return nil
}

func (c *Config) SetLogger() {
	c.logger = log.NewEntry(log.StandardLogger())
}

func (c *Config) LogOutput() io.Writer {
	return c.logger.Logger.Out
}
