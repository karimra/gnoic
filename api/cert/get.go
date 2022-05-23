package cert

import "github.com/openconfig/gnoi/cert"

func NewCertGetCertificatesRequest(opts ...CertOption) *cert.GetCertificatesRequest {
	return new(cert.GetCertificatesRequest)

}

func NewCertGetCertificatesResponse(opts ...CertOption) (*cert.GetCertificatesResponse, error) {
	m := new(cert.GetCertificatesResponse)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
