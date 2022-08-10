package config

import (
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/adrg/xdg"
	"github.com/mitchellh/go-homedir"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	PrintProto    bool          `mapstructure:"print-proto,omitempty" json:"print-proto,omitempty" yaml:"print-proto,omitempty"`
	Gzip          bool          `mapstructure:"gzip,omitempty" json:"gzip,omitempty" yaml:"gzip,omitempty"`
	Format        string        `mapstructure:"format,omitempty" json:"format,omitempty" yaml:"format,omitempty"`
}

type LocalFlags struct {
	// tree
	TreeFlat    bool `json:"tree-flat,omitempty" mapstructure:"tree-flat,omitempty" yaml:"tree-flat,omitempty"`
	TreeDetails bool `json:"tree-details,omitempty" mapstructure:"tree-details,omitempty" yaml:"tree-details,omitempty"`
	// VersionUpgrade
	UpgradeUsePkg bool `mapstructure:"upgrade-use-pkg" json:"upgrade-use-pkg,omitempty" yaml:"upgrade-use-pkg,omitempty"`
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
	CertRotateGenCSR          bool          `json:"cert-rotate-gen-csr,omitempty" mapstructure:"cert-rotate-gen-csr,omitempty" yaml:"cert-rotate-gen-csr,omitempty"`
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
	CertInstallGenCSR          bool          `json:"cert-install-gen-csr,omitempty" mapstructure:"cert-install-gen-csr,omitempty" yaml:"cert-install-gen-csr,omitempty"`
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
	CertGenerateCSRIPAddress       string `json:"cert-generate-csr-ip-address,omitempty" mapstructure:"cert-generate-csr-ip-address,omitempty" yaml:"cert-generate-cs-rip-address,omitempty"`
	CertGenerateCSREmailID         string `json:"cert-generate-csr-email-id,omitempty" mapstructure:"cert-generate-csr-email-id,omitempty" yaml:"cert-generate-csr-email-id,omitempty"`
	// Cert CanGenerateCSR
	CertCanGenerateCSRKeyType         string `json:"cert-can-generate-csr-key-type,omitempty" mapstructure:"cert-can-generate-csr-key-type,omitempty" yaml:"cert-can-generate-csr-key-type,omitempty"`
	CertCanGenerateCSRCertificateType string `json:"cert-can-generate-csr-certificate-type,omitempty" mapstructure:"cert-can-generate-csr-certificate-type,omitempty" yaml:"cert-can-generate-csr-certificate-type,omitempty"`
	CertCanGenerateCSRKeySize         uint32 `json:"cert-can-generate-csr-key-size,omitempty" mapstructure:"cert-can-generate-csr-key-size,omitempty" yaml:"cert-can-generate-csr-key-size,omitempty"`
	// Cert LoadCertificate
	CertLoadCertificateCertificateID   string   `json:"cert-load-certificate-certificate-id,omitempty" mapstructure:"cert-load-certificate-certificate-id,omitempty" yaml:"cert-load-certificate-certificate-id,omitempty"`
	CertLoadCertificateCertificateType string   `json:"cert-load-certificate-certificate-type,omitempty" mapstructure:"cert-load-certificate-certificate-type,omitempty" yaml:"cert-load-certificate-certificate-type,omitempty"`
	CertLoadCertificateCertificate     string   `json:"cert-load-certificate-certificate,omitempty" mapstructure:"cert-load-certificate-certificate,omitempty" yaml:"cert-load-certificate-certificate,omitempty"`
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
	CertGetCertificatesSave    bool     `json:"cert-get-certificates-save,omitempty" mapstructure:"cert-get-certificates-save,omitempty" yaml:"cert-get-certificates-save,omitempty"`
	// File
	// File Get
	FileGetFile         []string `json:"file-get-file,omitempty" mapstructure:"file-get-file,omitempty" yaml:"file-get-file,omitempty"`
	FileGetDst          string   `json:"file-get-dst,omitempty" mapstructure:"file-get-dst,omitempty" yaml:"file-get-dst,omitempty"`
	FileGetTargetPrefix bool     `json:"file-get-target-prefix,omitempty" mapstructure:"file-get-target-prefix,omitempty" yaml:"file-get-target-prefix,omitempty"`
	// File Stat
	FileStatPath      []string `json:"file-stat-path,omitempty" mapstructure:"file-stat-path,omitempty" yaml:"file-stat-path,omitempty"`
	FileStatHumanize  bool     `json:"file-stat-humanize,omitempty" mapstructure:"file-stat-humanize,omitempty" yaml:"file-stat-humanize,omitempty"`
	FileStatRecursive bool     `json:"file-stat-recursive,omitempty" mapstructure:"file-stat-recursive,omitempty" yaml:"file-stat-recursive,omitempty"`
	// File Put
	FilePutFile        []string `json:"file-put-file,omitempty" mapstructure:"file-put-file,omitempty" yaml:"file-put-file,omitempty"`
	FilePutDst         string   `json:"file-put-dst,omitempty" mapstructure:"file-put-dst,omitempty" yaml:"file-put-dst,omitempty"`
	FilePutPermissions uint32   `json:"file-put-permissions,omitempty" mapstructure:"file-put-permissions,omitempty" yaml:"file-put-permissions,omitempty"`
	FilePutChunkSize   uint64   `json:"file-put-chunk-size,omitempty" mapstructure:"file-put-chunk-size,omitempty" yaml:"file-put-chunk-size,omitempty"`
	FilePutHashMethod  string   `json:"file-put-hash-method,omitempty" mapstructure:"file-put-hash-method,omitempty" yaml:"file-put-hash-method,omitempty"`
	// File Remove
	FileRemovePath []string `json:"file-remove-path,omitempty" mapstructure:"file-remove-path,omitempty" yaml:"file-remove-path,omitempty"`
	// File Transfer
	FileTransferRemote        string `json:"file-transfer-remote,omitempty" mapstructure:"file-transfer-remote,omitempty" yaml:"file-transfer-remote,omitempty"`
	FileTransferLocal         string `json:"file-transfer-local,omitempty" mapstructure:"file-transfer-local,omitempty" yaml:"file-transfer-local,omitempty"`
	FileTransferSourceAddress string `json:"file-transfer-source-address,omitempty" mapstructure:"file-transfer-source-address,omitempty" yaml:"file-transfer-source-address,omitempty"`
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
	SystemTracerouteDestination    string        `json:"system-traceroute-destination,omitempty" mapstructure:"system-traceroute-destination,omitempty" yaml:"system-traceroute-destination,omitempty"`
	SystemTracerouteSource         string        `json:"system-traceroute-source,omitempty" mapstructure:"system-traceroute-source,omitempty" yaml:"system-traceroute-source,omitempty"`
	SystemTracerouteInterval       time.Duration `json:"system-traceroute-interval,omitempty" mapstructure:"system-traceroute-interval,omitempty" yaml:"system-traceroute-interval,omitempty"`
	SystemTracerouteWait           time.Duration `json:"system-traceroute-wait,omitempty" mapstructure:"system-traceroute-wait,omitempty" yaml:"system-traceroute-wait,omitempty"`
	SystemTracerouteInitialTTL     uint32        `json:"system-traceroute-initial-ttl,omitempty" mapstructure:"system-traceroute-initial-ttl,omitempty" yaml:"system-traceroute-initial-ttl,omitempty"`
	SystemTracerouteMaxTTL         int32         `json:"system-traceroute-max-ttl,omitempty" mapstructure:"system-traceroute-max-ttl,omitempty" yaml:"system-traceroute-max-ttl,omitempty"`
	SystemTracerouteSize           int32         `json:"system-traceroute-size,omitempty" mapstructure:"system-traceroute-size,omitempty" yaml:"system-traceroute-size,omitempty"`
	SystemTracerouteDoNotFragment  bool          `json:"system-traceroute-do-not-fragment,omitempty" mapstructure:"system-traceroute-do-not-fragment,omitempty" yaml:"system-traceroute-do-not-fragment,omitempty"`
	SystemTracerouteDoNotResolve   bool          `json:"system-traceroute-do-not-resolve,omitempty" mapstructure:"system-traceroute-do-not-resolve,omitempty" yaml:"system-traceroute-do-not-resolve,omitempty"`
	SystemTracerouteL3Protocol     string        `json:"system-traceroute-l3-protocol,omitempty" mapstructure:"system-traceroute-l3-protocol,omitempty" yaml:"system-traceroute-l3-protocol,omitempty"`
	SystemTracerouteL4Protocol     string        `json:"system-traceroute-l4-protocol,omitempty" mapstructure:"system-traceroute-l4-protocol,omitempty" yaml:"system-traceroute-l4-protocol,omitempty"`
	SystemTracerouteDoNotLookupAsn bool          `json:"system-traceroute-do-not-lookup-asn,omitempty" mapstructure:"system-traceroute-do-not-lookup-asn,omitempty" yaml:"system-traceroute-do-not-lookup-asn,omitempty"`
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
	SystemSetPackageFile        string `json:"system-set-package-file,omitempty" mapstructure:"system-set-package-file,omitempty" yaml:"system-set-package-file,omitempty"`
	SystemSetPackageVersion     string `json:"system-set-package-version,omitempty" mapstructure:"system-set-package-version,omitempty" yaml:"system-set-package-version,omitempty"`
	SystemSetPackageActivate    bool   `json:"system-set-package-activate,omitempty" mapstructure:"system-set-package-activate,omitempty" yaml:"system-set-package-activate,omitempty"`
	SystemSetPackageRemoteFile  string `json:"system-set-package-remote-file,omitempty" mapstructure:"system-set-package-remote-file,omitempty" yaml:"system-set-package-remote-file,omitempty"`
	SystemSetPackageCredentials string `json:"system-set-package-credentials,omitempty" mapstructure:"system-set-package-credentials,omitempty" yaml:"system-set-package-credentials,omitempty"`
	SystemSetPackageChunkSize   uint64 `json:"system-set-package-chunk-size,omitempty" mapstructure:"system-set-package-chunk-size,omitempty" yaml:"system-set-package-chunk-size,omitempty"`
	// Healthz
	// Healthz Get
	HealthzGetPath string `json:"healthz-get-path,omitempty" mapstructure:"healthz-get-path,omitempty" yaml:"healthz-get-path,omitempty"`
	// OS
	// OS Install
	OsInstallVersion           string `json:"os-install-version,omitempty" mapstructure:"os-install-version,omitempty" yaml:"os-install-version,omitempty"`
	OsInstallStandbySupervisor bool   `json:"os-install-standby-supervisor,omitempty" mapstructure:"os-install-standby-supervisor,omitempty" yaml:"os-install-standby-supervisor,omitempty"`
	OsInstallPackage           string `json:"os-install-package,omitempty" mapstructure:"os-install-package,omitempty" yaml:"os-install-package,omitempty"`
	OsInstallContentSize       uint64 `json:"os-install-content-size,omitempty" mapstructure:"os-install-content-size,omitempty" yaml:"os-install-content-size,omitempty"`
	// OS Activate
	OsActivateVersion           string `json:"os-activate-version,omitempty" mapstructure:"os-activate-version,omitempty" yaml:"os-activate-version,omitempty"`
	OsActivateStandbySupervisor bool   `json:"os-activate-standby-supervisor,omitempty" mapstructure:"os-activate-standby-supervisor,omitempty" yaml:"os-activate-standby-supervisor,omitempty"`
	OsActivateNoReboot          bool   `json:"os-activate-no-reboot,omitempty" mapstructure:"os-activate-no-reboot,omitempty" yaml:"os-activate-no-reboot,omitempty"`
	// Server
	ServerFile     bool   `json:"server-file,omitempty" mapstructure:"server-file,omitempty" yaml:"server-file,omitempty"`
	ServerFileHash string `json:"server-file-hash,omitempty" mapstructure:"server-file-hash,omitempty" yaml:"server-file-hash,omitempty"`
	// FactoryReset
	FactoryResetStartFactoryOS bool `json:"factory-reset-start-factory-os,omitempty" mapstructure:"factory-reset-start-factory-os,omitempty" yaml:"factory-reset-start-factory-os,omitempty"`
	FactoryResetStartZeroFill  bool `json:"factory-reset-start-zero-fill,omitempty" mapstructure:"factory-reset-start-zero-fill,omitempty" yaml:"factory-reset-start-zero-fill,omitempty"`
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
	logger := log.StandardLogger()
	if c.Debug {
		logger.SetLevel(log.DebugLevel)
	}
	c.logger = log.NewEntry(logger)
}

func (c *Config) LogOutput() io.Writer {
	return c.logger.Logger.Out
}

func (c *Config) SetPersistantFlagsFromFile(cmd *cobra.Command) {
	cmd.PersistentFlags().VisitAll(func(f *pflag.Flag) {
		flagName := flagFullName(cmd, f.Name)
		c.logger.Debugf("cmd=%s, flagName=%s, changed=%v, isSetInFile=%v",
			cmd.Name(), flagName, f.Changed, c.FileConfig.IsSet(f.Name))
		if !f.Changed && c.FileConfig.IsSet(f.Name) {
			c.setFlagValue(cmd, f.Name, c.FileConfig.Get(flagName))
		}
	})
}

func (c *Config) SetLocalFlagsFromFile(cmd *cobra.Command) {
	cmd.LocalFlags().VisitAll(func(f *pflag.Flag) {
		flagName := flagFullName(cmd, f.Name)
		c.logger.Debugf("cmd=%s, flagName=%s, changed=%v, isSetInFile=%v",
			cmd.Name(), flagName, f.Changed, c.FileConfig.IsSet(flagName))
		if !f.Changed && c.FileConfig.IsSet(flagName) {
			c.setFlagValue(cmd, f.Name, c.FileConfig.Get(flagName))
		}
	})
}

func (c *Config) setFlagValue(cmd *cobra.Command, fName string, val interface{}) {
	switch val := val.(type) {
	case []interface{}:
		c.logger.Debugf("cmd=%s, flagName=%s, valueType=%T, length=%d, value=%#v",
			cmd.Name(), fName, val, len(val), val)

		nVal := make([]string, 0, len(val))
		for _, v := range val {
			nVal = append(nVal, fmt.Sprintf("%v", v))
		}
		cmd.Flags().Set(fName, strings.Join(nVal, ","))
	default:
		c.logger.Debugf("cmd=%s, flagName=%s, valueType=%T, value=%#v",
			cmd.Name(), fName, val, val)
		cmd.Flags().Set(fName, fmt.Sprintf("%v", val))
	}
}

func flagFullName(cmd *cobra.Command, fName string) string {
	if cmd.Name() == "gnoic" {
		return fName
	}
	ls := []string{cmd.Name(), fName}
	for cmd.Parent() != nil && cmd.Parent().Name() != "gnoic" {
		ls = append([]string{cmd.Parent().Name()}, ls...)
		cmd = cmd.Parent()
	}
	return strings.Join(ls, "-")
}
