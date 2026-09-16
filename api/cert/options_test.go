package cert

import (
	"errors"
	"testing"

	"github.com/openconfig/gnoi/cert"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/karimra/gnoic/api"
)

func TestOptions_InvalidMessage(t *testing.T) {
	opts := map[string]CertOption{
		"CertificateType":     CertificateType("CT_X509"),
		"CertificateTypeX509": CertificateTypeX509(),
		"CertificateBytes":    CertificateBytes([]byte("x")),
		"CertificateInfo":     CertificateInfo(),
		"Certificate":         Certificate(),
		"CertificateID":       CertificateID("id"),
		"CaCertificate":       CaCertificate(),
		"ErrorMsg":            ErrorMsg("e"),
		"CSRParams":           CSRParams(),
		"CSR":                 CSR(),
		"KeySize":             KeySize(2048),
		"MinKeySize":          MinKeySize(2048),
		"KeyType":             KeyType("KT_RSA"),
		"KeyPair":             KeyPair(),
		"PublicKey":           PublicKey([]byte("x")),
		"PrivateKey":          PrivateKey([]byte("x")),
		"CommonName":          CommonName("cn"),
		"Country":             Country("c"),
		"State":               State("s"),
		"City":                City("c"),
		"Org":                 Org("o"),
		"OrgUnit":             OrgUnit("ou"),
		"IPAddress":           IPAddress("1.1.1.1"),
		"EmailID":             EmailID("a@b"),
		"Endpoint":            Endpoint(cert.Endpoint_EP_IPSEC_TUNNEL, "x"),
		"ModificationTime":    ModificationTime(1),
	}
	for name, o := range opts {
		t.Run(name, func(t *testing.T) {
			if err := o(nil); !errors.Is(err, api.ErrInvalidMsgType) {
				t.Errorf("nil message: got %v, want ErrInvalidMsgType", err)
			}
			if err := o(&emptypb.Empty{}); !errors.Is(err, api.ErrInvalidMsgType) {
				t.Errorf("wrong message type: got %v, want ErrInvalidMsgType", err)
			}
		})
	}
}

func TestOptions_InvalidValue(t *testing.T) {
	if err := CertificateType("CT_PEM")(&cert.CSRParams{}); !errors.Is(err, api.ErrInvalidValue) {
		t.Errorf("unknown certificate type: %v", err)
	}
	if err := KeyType("KT_DSA")(&cert.CSRParams{}); !errors.Is(err, api.ErrInvalidValue) {
		t.Errorf("unknown key type: %v", err)
	}
}

func TestNewCertCanGenerateCSRRequest(t *testing.T) {
	req, err := NewCertCanGenerateCSRRequest(CertificateTypeX509(), KeyType("kt_rsa"), KeySize(2048))
	if err != nil {
		t.Fatal(err)
	}
	if req.GetCertificateType() != cert.CertificateType_CT_X509 || req.GetKeyType() != cert.KeyType_KT_RSA || req.GetKeySize() != 2048 {
		t.Errorf("unexpected request: %v", req)
	}
	if _, err := NewCertCanGenerateCSRRequest(CommonName("x")); err == nil {
		t.Error("expected error for inapplicable option")
	}
	if _, err := NewCertCanGenerateCSRResponse(); err != nil {
		t.Error(err)
	}
}

func TestNewCertGenerateCSRRequest(t *testing.T) {
	req, err := NewCertGenerateCSRRequest(
		CertificateID("cert-1"),
		CSRParams(
			CertificateTypeX509(), MinKeySize(4096), KeyType("KT_RSA"),
			CommonName("router"), Country("US"), State("CA"), City("SF"), Org("org"), OrgUnit("ou"),
			IPAddress("10.0.0.1"), EmailID("a@b.c"),
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	if req.GetCertificateId() != "cert-1" {
		t.Errorf("CertificateId = %q", req.GetCertificateId())
	}
	p := req.GetCsrParams()
	if p.GetType() != cert.CertificateType_CT_X509 || p.GetMinKeySize() != 4096 || p.GetKeyType() != cert.KeyType_KT_RSA ||
		p.GetCommonName() != "router" || p.GetCountry() != "US" || p.GetState() != "CA" || p.GetCity() != "SF" ||
		p.GetOrganization() != "org" || p.GetOrganizationalUnit() != "ou" || p.GetIpAddress() != "10.0.0.1" || p.GetEmailId() != "a@b.c" {
		t.Errorf("unexpected CSR params: %v", p)
	}
	// nested option errors propagate
	if _, err := NewCertGenerateCSRRequest(CSRParams(CertificateID("x"))); !errors.Is(err, api.ErrInvalidMsgType) {
		t.Errorf("nested invalid option: %v", err)
	}
	rsp, err := NewCertGenerateCSRResponse(CSR(CertificateTypeX509()))
	if err != nil {
		t.Fatal(err)
	}
	if rsp.GetCsr().GetType() != cert.CertificateType_CT_X509 {
		t.Errorf("CSR = %v", rsp.GetCsr())
	}
}

func TestNewCertLoadCertificateRequest(t *testing.T) {
	req, err := NewCertLoadCertificateRequest(
		CertificateID("cert-1"),
		Certificate(CertificateTypeX509(), CertificateBytes([]byte("cert"))),
		KeyPair(PublicKey([]byte("pub")), PrivateKey([]byte("priv"))),
		CaCertificate(CertificateTypeX509(), CertificateBytes([]byte("ca1"))),
		CaCertificate(CertificateTypeX509(), CertificateBytes([]byte("ca2"))),
	)
	if err != nil {
		t.Fatal(err)
	}
	if req.GetCertificateId() != "cert-1" {
		t.Errorf("CertificateId = %q", req.GetCertificateId())
	}
	if req.GetCertificate().GetType() != cert.CertificateType_CT_X509 || string(req.GetCertificate().GetCertificate()) != "cert" {
		t.Errorf("Certificate = %v", req.GetCertificate())
	}
	if string(req.GetKeyPair().GetPublicKey()) != "pub" || string(req.GetKeyPair().GetPrivateKey()) != "priv" {
		t.Errorf("KeyPair = %v", req.GetKeyPair())
	}
	if len(req.GetCaCertificates()) != 2 || string(req.GetCaCertificates()[1].GetCertificate()) != "ca2" {
		t.Errorf("CaCertificates = %v", req.GetCaCertificates())
	}
	if _, err := NewCertLoadCertificateRequest(Certificate(CommonName("x"))); !errors.Is(err, api.ErrInvalidMsgType) {
		t.Errorf("nested invalid certificate option: %v", err)
	}
	if _, err := NewCertLoadCertificateRequest(KeyPair(CommonName("x"))); !errors.Is(err, api.ErrInvalidMsgType) {
		t.Errorf("nested invalid keypair option: %v", err)
	}
	if _, err := NewCertLoadCertificateRequest(CaCertificate(CommonName("x"))); !errors.Is(err, api.ErrInvalidMsgType) {
		t.Errorf("nested invalid ca option: %v", err)
	}
	if _, err := NewCertLoadCertificateResponse(); err != nil {
		t.Error(err)
	}
}

func TestNewCertLoadCertificateAuthorityBundleRequest(t *testing.T) {
	req, err := NewCertLoadCertificateAuthorityBundleRequest(
		CaCertificate(CertificateBytes([]byte("ca1"))),
		CaCertificate(CertificateBytes([]byte("ca2"))),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(req.GetCaCertificates()) != 2 {
		t.Errorf("CaCertificates = %v", req.GetCaCertificates())
	}
	if _, err := NewCertLoadCertificateAuthorityBundleRequest(CaCertificate(KeySize(1))); err == nil {
		t.Error("expected nested error")
	}
	if _, err := NewCertLoadCertificateAuthorityBundleResponse(); err != nil {
		t.Error(err)
	}
}

func TestNewCertRevokeCertificatesRequestResponse(t *testing.T) {
	req, err := NewCertRevokeCertificatesRequest(CertificateID("a"), CertificateID("b"))
	if err != nil {
		t.Fatal(err)
	}
	if len(req.GetCertificateId()) != 2 || req.GetCertificateId()[1] != "b" {
		t.Errorf("CertificateId = %v", req.GetCertificateId())
	}
	rsp, err := NewCertRevokeCertificatesResponse(CertificateID("a"))
	if err != nil {
		t.Fatal(err)
	}
	if len(rsp.GetRevokedCertificateId()) != 1 || rsp.GetRevokedCertificateId()[0] != "a" {
		t.Errorf("RevokedCertificateId = %v", rsp.GetRevokedCertificateId())
	}
	revErr := new(cert.CertificateRevocationError)
	if err := apply(revErr, CertificateID("c"), ErrorMsg("boom")); err != nil {
		t.Fatal(err)
	}
	if revErr.GetCertificateId() != "c" || revErr.GetErrorMessage() != "boom" {
		t.Errorf("CertificateRevocationError = %v", revErr)
	}
}

func TestNewCertGetCertificatesResponse(t *testing.T) {
	if NewCertGetCertificatesRequest() == nil {
		t.Fatal("nil get request")
	}
	rsp, err := NewCertGetCertificatesResponse(
		CertificateInfo(
			CertificateID("a"),
			Certificate(CertificateTypeX509(), CertificateBytes([]byte("x"))),
			Endpoint(cert.Endpoint_EP_IPSEC_TUNNEL, "tun0"),
			Endpoint(cert.Endpoint_EP_DAEMON, "sshd"),
			ModificationTime(123),
		),
		CertificateInfo(CertificateID("b")),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(rsp.GetCertificateInfo()) != 2 {
		t.Fatalf("got %d infos", len(rsp.GetCertificateInfo()))
	}
	ci := rsp.GetCertificateInfo()[0]
	if ci.GetCertificateId() != "a" || ci.GetCertificate().GetType() != cert.CertificateType_CT_X509 ||
		len(ci.GetEndpoints()) != 2 || ci.GetEndpoints()[1].GetEndpoint() != "sshd" || ci.GetModificationTime() != 123 {
		t.Errorf("CertificateInfo = %v", ci)
	}
	if _, err := NewCertGetCertificatesResponse(CertificateInfo(KeySize(1))); err == nil {
		t.Error("expected nested error")
	}
}

func TestInstallAndRotateWrappers(t *testing.T) {
	inst, err := NewCertInstallGenerateCSRRequest(CertificateID("a"))
	if err != nil {
		t.Fatal(err)
	}
	if inst.GetGenerateCsr().GetCertificateId() != "a" {
		t.Errorf("install gen csr = %v", inst)
	}
	if _, err := NewCertInstallGenerateCSRRequest(KeySize(1)); err == nil {
		t.Error("expected error")
	}
	instLoad, err := NewCertInstallLoadCertificateRequest(CertificateID("a"))
	if err != nil {
		t.Fatal(err)
	}
	if instLoad.GetLoadCertificate().GetCertificateId() != "a" {
		t.Errorf("install load = %v", instLoad)
	}
	if _, err := NewCertInstallLoadCertificateRequest(KeySize(1)); err == nil {
		t.Error("expected error")
	}
	if r, err := NewCertInstallGenerateCSRResponse(CSR(CertificateTypeX509())); err != nil || r.GetGeneratedCsr().GetCsr().GetType() != cert.CertificateType_CT_X509 {
		t.Errorf("install gen csr rsp = %v, %v", r, err)
	}
	if _, err := NewCertInstallGenerateCSRResponse(KeySize(1)); err == nil {
		t.Error("expected error")
	}
	if r, err := NewCertInstallLoadCertificateResponse(); err != nil || r.GetLoadCertificate() == nil {
		t.Errorf("install load rsp = %v, %v", r, err)
	}
	if _, err := NewCertInstallLoadCertificateResponse(KeySize(1)); err == nil {
		t.Error("expected error")
	}

	rot, err := NewCertRotateGenerateCSRRequest(CertificateID("a"))
	if err != nil {
		t.Fatal(err)
	}
	if rot.GetGenerateCsr().GetCertificateId() != "a" {
		t.Errorf("rotate gen csr = %v", rot)
	}
	if _, err := NewCertRotateGenerateCSRRequest(KeySize(1)); err == nil {
		t.Error("expected error")
	}
	rotLoad, err := NewCertRotateLoadCertificateRequest(CertificateID("a"))
	if err != nil {
		t.Fatal(err)
	}
	if rotLoad.GetLoadCertificate().GetCertificateId() != "a" {
		t.Errorf("rotate load = %v", rotLoad)
	}
	if _, err := NewCertRotateLoadCertificateRequest(KeySize(1)); err == nil {
		t.Error("expected error")
	}
	if NewCertRotateFinalizeRequest().GetFinalizeRotation() == nil {
		t.Error("finalize request missing FinalizeRotation")
	}
	if r, err := NewCertRotateGenerateCSRResponse(CSR(CertificateTypeX509())); err != nil || r.GetGeneratedCsr().GetCsr().GetType() != cert.CertificateType_CT_X509 {
		t.Errorf("rotate gen csr rsp = %v, %v", r, err)
	}
	if _, err := NewCertRotateGenerateCSRResponse(KeySize(1)); err == nil {
		t.Error("expected error")
	}
	if r, err := NewCertRotateLoadCertificateResponse(); err != nil || r.GetLoadCertificate() == nil {
		t.Errorf("rotate load rsp = %v, %v", r, err)
	}
	if _, err := NewCertRotateLoadCertificateResponse(KeySize(1)); err == nil {
		t.Error("expected error")
	}
}
