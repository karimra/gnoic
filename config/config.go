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
	configName = ".gnoic"
	envPrefix  = "GNOIC"
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
	ProxyFromEnv  bool          `mapstructure:"proxy-from-env,omitempty" json:"proxy-from-env,omitempty" yaml:"proxy-from-env,omitempty"`
	PrintRequest  bool          `mapstructure:"print-request,omitempty" json:"print-request,omitempty" yaml:"print-request,omitempty"`
	Gzip          bool          `mapstructure:"gzip,omitempty" json:"gzip,omitempty" yaml:"gzip,omitempty"`
}

type LocalFlags struct {
	// tree
	TreeFlat    bool `json:"tree-flat,omitempty" mapstructure:"tree-flat,omitempty" yaml:"tree-flat,omitempty"`
	TreeDetails bool `json:"tree-details,omitempty" mapstructure:"tree-details,omitempty" yaml:"tree-details,omitempty"`
	// Cert
	CertCACert string `json:"cert-ca-cert,omitempty" mapstructure:"cert-ca-cert,omitempty" yaml:"cert-ca-cert,omitempty"`
	CertCAKey  string `json:"cert-ca-key,omitempty" mapstructure:"cert-ca-key,omitempty" yaml:"cert-ca-key,omitempty"`
	// Cert CreateCA
	CertCreateCaOrg           string        `json:"cert-create-ca-org,omitempty" mapstructure:"cert-create-ca-org,omitempty" yaml:"cert-create-ca-org,omitempty"`
	CertCreateCaOrgUnit       string        `json:"cert-create-ca-org-unit,omitempty" mapstructure:"cert-create-ca-org-unit,omitempty" yaml:"cert-create-ca-org-unit,omitempty"`
	CertCreateCaCountry       string        `json:"cert-create-ca-country,omitempty" mapstructure:"cert-create-ca-country,omitempty" yaml:"cert-create-ca-country,omitempty"`
	CertCreateCaState         string        `json:"cert-create-ca-state,omitempty" mapstructure:"cert-create-ca-state,omitempty" yaml:"cert-create-ca-state,omitempty"`
	CertCreateCaLocality      string        `json:"cert-create-ca-locality,omitempty" mapstructure:"cert-create-ca-locality,omitempty" yaml:"cert-create-ca-locality,omitempty"`
	CertCreateCaStreetAddress string        `json:"cert-create-ca-street-address,omitempty" mapstructure:"cert-create-ca-street-address,omitempty" yaml:"cert-create-ca-street-address,omitempty"`
	CertCreateCaPostalCode    string        `json:"cert-create-ca-postal-code,omitempty" mapstructure:"cert-create-ca-postal-code,omitempty" yaml:"cert-create-ca-postal-code,omitempty"`
	CertCreateCaValidity      time.Duration `json:"cert-create-ca-validity,omitempty" mapstructure:"cert-create-ca-validity,omitempty" yaml:"cert-create-ca-validity,omitempty"`
	CertCreateCaKeySize       int           `json:"cert-create-ca-key-size,omitempty" mapstructure:"cert-create-ca-key-size,omitempty" yaml:"cert-create-ca-key-size,omitempty"`
	CertCreateCaEmailID       string        `json:"cert-create-ca-email-id,omitempty" mapstructure:"cert-create-ca-email-id,omitempty" yaml:"cert-create-ca-email-id,omitempty"`
	CertCreateCaCommonName    string        `json:"cert-create-ca-common-name,omitempty" mapstructure:"cert-create-ca-common-name,omitempty" yaml:"cert-create-ca-common-name,omitempty"`
	CertCreateCaKeyOut        string        `json:"cert-create-ca-key-out,omitempty" mapstructure:"cert-create-ca-key-out,omitempty" yaml:"cert-create-ca-key-out,omitempty"`
	CertCreateCaCertOut       string        `json:"cert-create-ca-cert-out,omitempty" mapstructure:"cert-create-ca-cert-out,omitempty" yaml:"cert-create-ca-cert-out,omitempty"`
	// Cert Rotate
	CertRotateCertificateID   string        `json:"cert-rotate-certificate-id,omitempty" mapstructure:"cert-rotate-certificate-id,omitempty" yaml:"cert-rotate-certificate-id,omitempty"`
	CertRotateKeyType         string        `json:"cert-rotate-key-type,omitempty" mapstructure:"cert-rotate-key-type,omitempty" yaml:"cert-rotate-key-type,omitempty"`
	CertRotateCertificateType string        `json:"cert-rotate-certificate-type,omitempty" mapstructure:"cert-rotate-certificate-type,omitempty" yaml:"cert-rotate-certificate-type,omitempty"`
	CertRotateMinKeySize      uint32        `json:"cert-rotate-min-key-size,omitempty" mapstructure:"cert-rotate-min-key-size,omitempty" yaml:"cert-rotate-min-key-size,omitempty"`
	CertRotateCommonName      string        `json:"cert-rotate-common-name,omitempty" mapstructure:"cert-rotate-common-name,omitempty" yaml:"cert-rotate-common-name,omitempty"`
	CertRotateCountry         string        `json:"cert-rotate-country,omitempty" mapstructure:"cert-rotate-country,omitempty" yaml:"cert-rotate-country,omitempty"`
	CertRotateState           string        `json:"cert-rotate-state,omitempty" mapstructure:"cert-rotate-state,omitempty" yaml:"cert-rotate-state,omitempty"`
	CertRotateCity            string        `json:"cert-rotate-city,omitempty" mapstructure:"cert-rotate-city,omitempty" yaml:"cert-rotate-city,omitempty"`
	CertRotateOrg             string        `json:"cert-rotate-org,omitempty" mapstructure:"cert-rotate-org,omitempty" yaml:"cert-rotate-org,omitempty"`
	CertRotateOrgUnit         string        `json:"cert-rotate-org-unit,omitempty" mapstructure:"cert-rotate-org-unit,omitempty" yaml:"cert-rotate-org-unit,omitempty"`
	CertRotateIPAddress       string        `json:"cert-rotate-ip-address,omitempty" mapstructure:"cert-rotate-ip-address,omitempty" yaml:"cert-rotate-ip-address,omitempty"`
	CertRotateEmailID         string        `json:"cert-rotate-email-id,omitempty" mapstructure:"cert-rotate-email-id,omitempty" yaml:"cert-rotate-email-id,omitempty"`
	CertRotateValidity        time.Duration `json:"cert-rotate-validity,omitempty" mapstructure:"cert-rotate-validity,omitempty" yaml:"cert-rotate-validity,omitempty"`
	CertRotatePrintCSR        bool          `json:"cert-rotate-print-csr,omitempty" mapstructure:"cert-rotate-print-csr,omitempty" yaml:"cert-rotate-print-csr,omitempty"`
	// Cert Install
	CertInstallCertificateID   string        `json:"cert-install-certificate-id,omitempty" mapstructure:"cert-install-certificate-id,omitempty" yaml:"cert-install-certificate-id,omitempty"`
	CertInstallKeyType         string        `json:"cert-install-key-type,omitempty" mapstructure:"cert-install-key-type,omitempty" yaml:"cert-install-key-type,omitempty"`
	CertInstallCertificateType string        `json:"cert-install-certificate-type,omitempty" mapstructure:"cert-install-certificate-type,omitempty" yaml:"cert-install-certificate-type,omitempty"`
	CertInstallMinKeySize      uint32        `json:"cert-install-min-key-size,omitempty" mapstructure:"cert-install-min-key-size,omitempty" yaml:"cert-install-min-key-size,omitempty"`
	CertInstallCommonName      string        `json:"cert-install-common-name,omitempty" mapstructure:"cert-install-common-name,omitempty" yaml:"cert-install-common-name,omitempty"`
	CertInstallCountry         string        `json:"cert-install-country,omitempty" mapstructure:"cert-install-country,omitempty" yaml:"cert-install-country,omitempty"`
	CertInstallState           string        `json:"cert-install-state,omitempty" mapstructure:"cert-install-state,omitempty" yaml:"cert-install-state,omitempty"`
	CertInstallCity            string        `json:"cert-install-city,omitempty" mapstructure:"cert-install-city,omitempty" yaml:"cert-install-city,omitempty"`
	CertInstallOrg             string        `json:"cert-install-org,omitempty" mapstructure:"cert-install-org,omitempty" yaml:"cert-install-org,omitempty"`
	CertInstallOrgUnit         string        `json:"cert-install-org-unit,omitempty" mapstructure:"cert-install-org-unit,omitempty" yaml:"cert-install-org-unit,omitempty"`
	CertInstallIPAddress       string        `json:"cert-install-ip-address,omitempty" mapstructure:"cert-install-ip-address,omitempty" yaml:"cert-install-ip-address,omitempty"`
	CertInstallEmailID         string        `json:"cert-install-email-id,omitempty" mapstructure:"cert-install-email-id,omitempty" yaml:"cert-install-email-id,omitempty"`
	CertInstallValidity        time.Duration `json:"cert-install-validity,omitempty" mapstructure:"cert-install-validity,omitempty" yaml:"cert-install-validity,omitempty"`
	CertInstallPrintCSR        bool          `json:"cert-install-print-csr,omitempty" mapstructure:"cert-install-print-csr,omitempty" yaml:"cert-install-print-csr,omitempty"`
	// Cert GenerateCSR
	CertGenerateCSRCertificateID   string `json:"cert-generate-csr-certificate-id,omitempty" mapstructure:"cert-generate-csr-certificate-id,omitempty" yaml:"cert-generate-csr-certificate-id,omitempty"`
	CertGenerateCSRKeyType         string `json:"cert-generate-csr-key-type,omitempty" mapstructure:"cert-generate-csr-key-type,omitempty" yaml:"cert-generate-csr-key-type,omitempty"`
	CertGenerateCSRCertificateType string `json:"cert-generate-csr-certificate-type,omitempty" mapstructure:"cert-generate-csr-certificate-type,omitempty" yaml:"cert-generate-csr-certificate-type,omitempty"`
	CertGenerateCSRMinKeySize      uint32 `json:"cert-generate-csr-min-key-size,omitempty" mapstructure:"cert-generate-csr-min-key-size,omitempty" yaml:"cert-generate-csr-min-key-size,omitempty"`
	CertGenerateCSRCommonName      string `json:"cert-generate-csr-common-name,omitempty" mapstructure:"cert-generate-csr-common-name,omitempty" yaml:"cert-generate-csr-common-name,omitempty"`
	CertGenerateCSRCountry         string `json:"cert-generate-csr-country,omitempty" mapstructure:"cert-generate-csr-country,omitempty" yaml:"cert-generate-csr-country,omitempty"`
	CertGenerateCSRState           string `json:"cert-generate-csr-state,omitempty" mapstructure:"cert-generate-csr-state,omitempty" yaml:"cert-generate-csr-state,omitempty"`
	CertGenerateCSRCity            string `json:"cert-generate-csr-city,omitempty" mapstructure:"cert-generate-csr-city,omitempty" yaml:"cert-generate-csr-city,omitempty"`
	CertGenerateCSROrg             string `json:"cert-generate-csr-org,omitempty" mapstructure:"cert-generate-csr-org,omitempty" yaml:"cert-generate-csr-org,omitempty"`
	CertGenerateCSROrgUnit         string `json:"cert-generate-csr-org-unit,omitempty" mapstructure:"cert-generate-csr-org-unit,omitempty" yaml:"cert-generate-csr-org-unit,omitempty"`
	CertGenerateCSRIPAddress       string `json:"cert-generate-csrip-address,omitempty" mapstructure:"cert-generate-csrip-address,omitempty" yaml:"cert-generate-csrip-address,omitempty"`
	CertGenerateCSREmailID         string `json:"cert-generate-csr-email-id,omitempty" mapstructure:"cert-generate-csr-email-id,omitempty" yaml:"cert-generate-csr-email-id,omitempty"`
	// Cert CanGenerateCSR
	CertCanGenerateCSRKeyType         string `json:"cert-can-generate-csr-key-type,omitempty" mapstructure:"cert-can-generate-csr-key-type,omitempty" yaml:"cert-can-generate-csr-key-type,omitempty"`
	CertCanGenerateCSRCertificateType string `json:"cert-can-generate-csr-certificate-type,omitempty" mapstructure:"cert-can-generate-csr-certificate-type,omitempty" yaml:"cert-can-generate-csr-certificate-type,omitempty"`
	CertCanGenerateCSRKeySize         uint32 `json:"cert-can-generate-csr-key-size,omitempty" mapstructure:"cert-can-generate-csr-key-size,omitempty" yaml:"cert-can-generate-csr-key-size,omitempty"`
	// Cert LoadCertificate
	CertLoadCertificateCertificateID   string   `json:"cert-load-certificate-certificate-id,omitempty" mapstructure:"cert-load-certificate-certificate-id,omitempty" yaml:"cert-load-certificate-certificate-id,omitempty"`
	CertLoadCertificateCertificateType string   `json:"cert-load-certificate-certificate-type,omitempty" mapstructure:"cert-load-certificate-certificate-type,omitempty" yaml:"cert-load-certificate-certificate-type,omitempty"`
	CertLoadCertificatePrivateKey      string   `json:"cert-load-certificate-private-key,omitempty" mapstructure:"cert-load-certificate-private-key,omitempty" yaml:"cert-load-certificate-private-key,omitempty"`
	CertLoadCertificatePublicKey       string   `json:"cert-load-certificate-public-key,omitempty" mapstructure:"cert-load-certificate-public-key,omitempty" yaml:"cert-load-certificate-public-key,omitempty"`
	CertLoadCertificateCaCertificates  []string `json:"cert-load-certificate-ca-certificates,omitempty" mapstructure:"cert-load-certificate-ca-certificates,omitempty" yaml:"cert-load-certificate-ca-certificates,omitempty"`
	// Cert LoadCertificateCanBundle
	CertLoadCertificateCaBundleCaCertificates []string `json:"cert-load-certificate-ca-bundle-ca-certificates,omitempty" mapstructure:"cert-load-certificate-ca-bundle-ca-certificates,omitempty" yaml:"cert-load-certificate-ca-bundle-ca-certificates,omitempty"`
	// Cert RevokeCertificate
	CertRevokeCertificatesCertificateID []string `json:"cert-revoke-certificates-certificate-id,omitempty" mapstructure:"cert-revoke-certificates-certificate-id,omitempty" yaml:"cert-revoke-certificates-certificate-id,omitempty"`
	CertRevokeCertificatesAll           bool     `json:"cert-revoke-certificates-all,omitempty" mapstructure:"cert-revoke-certificates-all,omitempty" yaml:"cert-revoke-certificates-all,omitempty"`
	// Cert GetCertificates
	CertGetCertificatesDetails bool     `json:"cert-get-certificates-details,omitempty" mapstructure:"cert-get-certificates-details,omitempty" yaml:"cert-get-certificates-details,omitempty"`
	CertGetCertificatesID      []string `json:"cert-get-certificates-id,omitempty" mapstructure:"cert-get-certificates-id,omitempty" yaml:"cert-get-certificates-id,omitempty"`
	CertGetCertificatesSave    bool
	// File
	// File Get
	FileGetFile         string `json:"file-get-file,omitempty" mapstructure:"file-get-file,omitempty" yaml:"file-get-file,omitempty"`
	FileGetLocalFile    string `json:"file-get-local-file,omitempty" mapstructure:"file-get-local-file,omitempty" yaml:"file-get-local-file,omitempty"`
	FileGetTargetPrefix bool   `json:"file-get-target-prefix,omitempty" mapstructure:"file-get-target-prefix,omitempty" yaml:"file-get-target-prefix,omitempty"`
	// File Stat
	FileStatFile     string `json:"file-stat-file,omitempty" mapstructure:"file-stat-file,omitempty" yaml:"file-stat-file,omitempty"`
	FileStatHumanize bool
	// File Put
	FilePutFile        string `json:"file-put-file,omitempty" mapstructure:"file-put-file,omitempty" yaml:"file-put-file,omitempty"`
	FilePutRemoteFile  string `json:"file-put-remote-file,omitempty" mapstructure:"file-put-remote-file,omitempty" yaml:"file-put-remote-file,omitempty"`
	FilePutPermissions uint32 `json:"file-put-permissions,omitempty" mapstructure:"file-put-permissions,omitempty" yaml:"file-put-permissions,omitempty"`
	FilePutWriteSize   uint64 `json:"file-put-write-size,omitempty" mapstructure:"file-put-write-size,omitempty" yaml:"file-put-write-size,omitempty"`
	FilePutHashMethod  string `json:"file-put-hash-method,omitempty" mapstructure:"file-put-hash-method,omitempty" yaml:"file-put-hash-method,omitempty"`
	// File Remove
	FileRemoveFile string `json:"file-remove-file,omitempty" mapstructure:"file-remove-file,omitempty" yaml:"file-remove-file,omitempty"`
	// System
	// System Ping
	SystemPingDestination   string        `json:"system-ping-destination,omitempty" mapstructure:"system-ping-destination,omitempty" yaml:"system-ping-destination,omitempty"`
	SystemPingSource        string        `json:"system-ping-source,omitempty" mapstructure:"system-ping-source,omitempty" yaml:"system-ping-source,omitempty"`
	SystemPingCount         int32         `json:"system-ping-count,omitempty" mapstructure:"system-ping-count,omitempty" yaml:"system-ping-count,omitempty"`
	SystemPingInterval      time.Duration `json:"system-ping-interval,omitempty" mapstructure:"system-ping-interval,omitempty" yaml:"system-ping-interval,omitempty"`
	SystemPingWait          time.Duration `json:"system-ping-wait,omitempty" mapstructure:"system-ping-wait,omitempty" yaml:"system-ping-wait,omitempty"`
	SystemPingSize          int32         `json:"system-ping-size,omitempty" mapstructure:"system-ping-size,omitempty" yaml:"system-ping-size,omitempty"`
	SystemPingDoNotFragment bool          `json:"system-ping-do-not-fragment,omitempty" mapstructure:"system-ping-do-not-fragment,omitempty" yaml:"system-ping-do-not-fragment,omitempty"`
	SystemPingDoNotResolve  bool          `json:"system-ping-do-not-resolve,omitempty" mapstructure:"system-ping-do-not-resolve,omitempty" yaml:"system-ping-do-not-resolve,omitempty"`
	SystemPingProtocol      string        `json:"system-ping-protocol,omitempty" mapstructure:"system-ping-protocol,omitempty" yaml:"system-ping-protocol,omitempty"`
	// System Traceroute
	SystemTracerouteDestination   string        `json:"system-traceroute-destination,omitempty" mapstructure:"system-traceroute-destination,omitempty" yaml:"system-traceroute-destination,omitempty"`
	SystemTracerouteSource        string        `json:"system-traceroute-source,omitempty" mapstructure:"system-traceroute-source,omitempty" yaml:"system-traceroute-source,omitempty"`
	SystemTracerouteInterval      time.Duration `json:"system-traceroute-interval,omitempty" mapstructure:"system-traceroute-interval,omitempty" yaml:"system-traceroute-interval,omitempty"`
	SystemTracerouteWait          time.Duration `json:"system-traceroute-wait,omitempty" mapstructure:"system-traceroute-wait,omitempty" yaml:"system-traceroute-wait,omitempty"`
	SystemTracerouteInitialTTL    uint32        `json:"system-traceroute-initial-ttl,omitempty" mapstructure:"system-traceroute-initial-ttl,omitempty" yaml:"system-traceroute-initial-ttl,omitempty"`
	SystemTracerouteMaxTTL        int32         `json:"system-traceroute-max-ttl,omitempty" mapstructure:"system-traceroute-max-ttl,omitempty" yaml:"system-traceroute-max-ttl,omitempty"`
	SystemTracerouteSize          int32         `json:"system-traceroute-size,omitempty" mapstructure:"system-traceroute-size,omitempty" yaml:"system-traceroute-size,omitempty"`
	SystemTracerouteDoNotFragment bool          `json:"system-traceroute-do-not-fragment,omitempty" mapstructure:"system-traceroute-do-not-fragment,omitempty" yaml:"system-traceroute-do-not-fragment,omitempty"`
	SystemTracerouteDoNotResolve  bool          `json:"system-traceroute-do-not-resolve,omitempty" mapstructure:"system-traceroute-do-not-resolve,omitempty" yaml:"system-traceroute-do-not-resolve,omitempty"`
	SystemTracerouteL3Protocol    string        `json:"system-traceroute-l-3-protocol,omitempty" mapstructure:"system-traceroute-l-3-protocol,omitempty" yaml:"system-traceroute-l-3-protocol,omitempty"`
	SystemTracerouteL4Protocol    string        `json:"system-traceroute-l-4-protocol,omitempty" mapstructure:"system-traceroute-l-4-protocol,omitempty" yaml:"system-traceroute-l-4-protocol,omitempty"`
	// System Reboot
	SystemRebootMethod         string        `json:"system-reboot-method,omitempty" mapstructure:"system-reboot-method,omitempty" yaml:"system-reboot-method,omitempty"`
	SystemRebootDelay          time.Duration `json:"system-reboot-delay,omitempty" mapstructure:"system-reboot-delay,omitempty" yaml:"system-reboot-delay,omitempty"`
	SystemRebootMessage        string        `json:"system-reboot-message,omitempty" mapstructure:"system-reboot-message,omitempty" yaml:"system-reboot-message,omitempty"`
	SystemRebootSubscomponents []string      `json:"system-reboot-subscomponents,omitempty" mapstructure:"system-reboot-subscomponents,omitempty" yaml:"system-reboot-subscomponents,omitempty"`
	SystemRebootForce          bool          `json:"system-reboot-force,omitempty" mapstructure:"system-reboot-force,omitempty" yaml:"system-reboot-force,omitempty"`
	// System RebootStatus
	SystemRebootStatusSubscomponents []string `json:"system-reboot-status-subscomponents,omitempty" mapstructure:"system-reboot-status-subscomponents,omitempty" yaml:"system-reboot-status-subscomponents,omitempty"`
	// System CancelReboot
	SystemCancelRebootMessage       string   `json:"system-cancel-reboot-message,omitempty" mapstructure:"system-cancel-reboot-message,omitempty" yaml:"system-cancel-reboot-message,omitempty"`
	SystemCancelRebootSubcomponents []string `json:"system-cancel-reboot-subcomponents,omitempty" mapstructure:"system-cancel-reboot-subcomponents,omitempty" yaml:"system-cancel-reboot-subcomponents,omitempty"`
	// System SwitchControlProcessor
	SystemSwitchControlProcessorPath string `json:"system-switch-control-processor-path,omitempty" mapstructure:"system-switch-control-processor-path,omitempty" yaml:"system-switch-control-processor-path,omitempty"`
	// System SetPackage
	SystemSetPackageFile        string
	SystemSetPackageVersion     string
	SystemSetPackageActivate    bool
	SystemSetPackageRemoteFile  string
	SystemSetPackageCredentials string
	SystemSetPackageChunkSize   uint64
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
