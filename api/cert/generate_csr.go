package cert

import "github.com/openconfig/gnoi/cert"

func NewCertGenerateCSRRequest(opts ...CertOption) (*cert.GenerateCSRRequest, error) {
	m := new(cert.GenerateCSRRequest)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewCertGenerateCSRResponse(opts ...CertOption) (*cert.GenerateCSRResponse, error) {
	m := new(cert.GenerateCSRResponse)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
