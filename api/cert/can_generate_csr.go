package cert

import "github.com/openconfig/gnoi/cert"

func NewCertCanGenerateCSRRequest(opts ...CertOption) (*cert.CanGenerateCSRRequest, error) {
	m := new(cert.CanGenerateCSRRequest)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewCertCanGenerateCSRResponse(opts ...CertOption) (*cert.CanGenerateCSRResponse, error) {
	m := new(cert.CanGenerateCSRResponse)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
