package cert

import "github.com/openconfig/gnoi/cert"

func NewCertLoadCertificateRequest(opts ...CertOption) (*cert.LoadCertificateRequest, error) {
	m := new(cert.LoadCertificateRequest)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewCertLoadCertificateResponse(opts ...CertOption) (*cert.LoadCertificateResponse, error) {
	m := new(cert.LoadCertificateResponse)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
