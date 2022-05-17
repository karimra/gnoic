package cert

import "github.com/openconfig/gnoi/cert"

func NewCertRevokeCertificatesRequest(opts ...CertOption) (*cert.RevokeCertificatesRequest, error) {
	m := new(cert.RevokeCertificatesRequest)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewCertRevokeCertificatesResponse(opts ...CertOption) (*cert.RevokeCertificatesResponse, error) {
	m := new(cert.RevokeCertificatesResponse)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
