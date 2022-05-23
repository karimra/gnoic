package cert

import (
	"github.com/openconfig/gnoi/cert"
)

func NewCertRotateGenerateCSRRequest(opts ...CertOption) (*cert.RotateCertificateRequest, error) {
	m, err := NewCertGenerateCSRRequest(opts...)
	if err != nil {
		return nil, err
	}
	return &cert.RotateCertificateRequest{
		RotateRequest: &cert.RotateCertificateRequest_GenerateCsr{
			GenerateCsr: m,
		},
	}, nil
}

func NewCertRotateLoadCertificateRequest(opts ...CertOption) (*cert.RotateCertificateRequest, error) {
	m, err := NewCertLoadCertificateRequest(opts...)
	if err != nil {
		return nil, err
	}
	return &cert.RotateCertificateRequest{
		RotateRequest: &cert.RotateCertificateRequest_LoadCertificate{
			LoadCertificate: m,
		},
	}, nil
}

func NewCertRotateFinalizeRequest(opts ...CertOption) *cert.RotateCertificateRequest {
	return &cert.RotateCertificateRequest{
		RotateRequest: &cert.RotateCertificateRequest_FinalizeRotation{
			FinalizeRotation: &cert.FinalizeRequest{},
		},
	}
}

func NewCertRotateGenerateCSRResponse(opts ...CertOption) (*cert.RotateCertificateResponse, error) {
	m, err := NewCertGenerateCSRResponse(opts...)
	if err != nil {
		return nil, err
	}
	return &cert.RotateCertificateResponse{
		RotateResponse: &cert.RotateCertificateResponse_GeneratedCsr{
			GeneratedCsr: m,
		},
	}, nil
}

func NewCertRotateLoadCertificateResponse(opts ...CertOption) (*cert.RotateCertificateResponse, error) {
	m, err := NewCertLoadCertificateResponse(opts...)
	if err != nil {
		return nil, err
	}
	return &cert.RotateCertificateResponse{
		RotateResponse: &cert.RotateCertificateResponse_LoadCertificate{
			LoadCertificate: m,
		},
	}, nil
}
