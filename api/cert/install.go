package cert

import "github.com/openconfig/gnoi/cert"

func NewCertInstallGenerateCSRRequest(opts ...CertOption) (*cert.InstallCertificateRequest, error) {
	m, err := NewCertGenerateCSRRequest(opts...)
	if err != nil {
		return nil, err
	}
	return &cert.InstallCertificateRequest{
		InstallRequest: &cert.InstallCertificateRequest_GenerateCsr{
			GenerateCsr: m,
		},
	}, nil
}

func NewCertInstallLoadCertificateRequest(opts ...CertOption) (*cert.InstallCertificateRequest, error) {
	m, err := NewCertLoadCertificateRequest(opts...)
	if err != nil {
		return nil, err
	}
	return &cert.InstallCertificateRequest{
		InstallRequest: &cert.InstallCertificateRequest_LoadCertificate{
			LoadCertificate: m,
		},
	}, nil
}

func NewCertInstallGenerateCSRResponse(opts ...CertOption) (*cert.InstallCertificateResponse, error) {
	m, err := NewCertGenerateCSRResponse(opts...)
	if err != nil {
		return nil, err
	}
	return &cert.InstallCertificateResponse{
		InstallResponse: &cert.InstallCertificateResponse_GeneratedCsr{
			GeneratedCsr: m,
		},
	}, nil
}

func NewCertInstallLoadCertificateResponse(opts ...CertOption) (*cert.InstallCertificateResponse, error) {
	m, err := NewCertLoadCertificateResponse(opts...)
	if err != nil {
		return nil, err
	}
	return &cert.InstallCertificateResponse{
		InstallResponse: &cert.InstallCertificateResponse_LoadCertificate{
			LoadCertificate: m,
		},
	}, nil
}
