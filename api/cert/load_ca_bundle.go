package cert

import "github.com/openconfig/gnoi/cert"

func NewCertLoadCertificateAuthorityBundleRequest(opts ...CertOption) (*cert.LoadCertificateAuthorityBundleRequest, error) {
	m := new(cert.LoadCertificateAuthorityBundleRequest)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewCertLoadCertificateAuthorityBundleResponse(opts ...CertOption) (*cert.LoadCertificateAuthorityBundleResponse, error) {
	m := new(cert.LoadCertificateAuthorityBundleResponse)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
