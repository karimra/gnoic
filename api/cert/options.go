package cert

import (
	"errors"
	"strings"

	"github.com/openconfig/gnoi/cert"
	"google.golang.org/protobuf/proto"
)

type CertOption func(proto.Message) error

// ErrInvalidMsgType is returned by a CertOption in case the Option is supplied
// an unexpected proto.Message
var ErrInvalidMsgType = errors.New("invalid message type")

// ErrInvalidValue is returned by a CertOption in case the Option is supplied
// an unexpected value.
var ErrInvalidValue = errors.New("invalid value")

// apply is a helper function that simply applies the options to the proto.Message.
// It returns an error if any of the options fails.
func apply(m proto.Message, opts ...CertOption) error {
	for _, o := range opts {
		if err := o(m); err != nil {
			return err
		}
	}
	return nil
}

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

func NewCertInstallRequest(opts ...CertOption) (*cert.InstallCertificateRequest, error) {
	m := new(cert.InstallCertificateRequest)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewCertInstallResponse(opts ...CertOption) (*cert.InstallCertificateResponse, error) {
	m := new(cert.InstallCertificateResponse)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewCertRotateRequest(opts ...CertOption) (*cert.RotateCertificateRequest, error) {
	m := new(cert.RotateCertificateRequest)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewCertRotateResponse(opts ...CertOption) (*cert.RotateCertificateResponse, error) {
	m := new(cert.RotateCertificateResponse)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

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

func NewCertGetCertificatesRequest(opts ...CertOption) (*cert.GetCertificatesRequest, error) {
	m := new(cert.GetCertificatesRequest)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewCertGetCertificatesResponse(opts ...CertOption) (*cert.GetCertificatesResponse, error) {
	m := new(cert.GetCertificatesResponse)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

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

func KeyType(kt string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return ErrInvalidMsgType
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CanGenerateCSRRequest:
			ktv, ok := cert.KeyType_value[strings.ToUpper(kt)]
			if !ok {
				return ErrInvalidValue
			}
			msg.KeyType = cert.KeyType(ktv)
		}
		return nil
	}
}

func CertificateType(ct string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return ErrInvalidMsgType
		}
		ctv, ok := cert.CertificateType_value[strings.ToUpper(ct)]
		if !ok {
			return ErrInvalidValue
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CanGenerateCSRRequest:
			msg.CertificateType = cert.CertificateType(ctv)
		case *cert.Certificate:
			msg.Type = cert.CertificateType(ctv)
		case *cert.CSR:
			msg.Type = cert.CertificateType(ctv)
		case *cert.CSRParams:
			msg.Type = cert.CertificateType(ctv)
		}
		return nil
	}
}

func KeySize(ks uint32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return ErrInvalidMsgType
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CanGenerateCSRRequest:
			msg.KeySize = ks
		case *cert.CSRParams:
			msg.MinKeySize = ks
		}
		return nil
	}
}

func MinKeySize(ks uint32) func(msg proto.Message) error {
	return KeySize(ks)
}
