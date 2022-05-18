package cert

import (
	"strings"

	"github.com/karimra/gnoic/api"
	"github.com/openconfig/gnoi/cert"
	"google.golang.org/protobuf/proto"
)

type CertOption func(proto.Message) error

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

func CertificateType(ct string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}
		ctv, ok := cert.CertificateType_value[strings.ToUpper(ct)]
		if !ok {
			return api.ErrInvalidValue
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

func CertificateInfo(opts ...CertOption) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.GetCertificatesResponse:
			m := new(cert.CertificateInfo)
			err := apply(m, opts...)
			if err != nil {
				return err
			}
			if len(msg.CertificateInfo) == 0 {
				msg.CertificateInfo = make([]*cert.CertificateInfo, 0, 1)
			}
			msg.CertificateInfo = append(msg.CertificateInfo, m)
		}
		return nil
	}
}

func Certificate(opts ...CertOption) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.LoadCertificateRequest:
			m := new(cert.Certificate)
			err := apply(m, opts...)
			if err != nil {
				return err
			}
			msg.Certificate = m
		case *cert.CertificateInfo:
			m := new(cert.Certificate)
			err := apply(m, opts...)
			if err != nil {
				return err
			}
			msg.Certificate = m
		}
		return nil
	}
}

func CertificateID(id string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.GenerateCSRRequest:
			msg.CertificateId = id
		case *cert.LoadCertificateRequest:
			msg.CertificateId = id
		case *cert.CertificateInfo:
			msg.CertificateId = id
		case *cert.RevokeCertificatesRequest:
			if msg.CertificateId == nil {
				msg.CertificateId = make([]string, 0, 1)
			}
			msg.CertificateId = append(msg.CertificateId, id)
		case *cert.RevokeCertificatesResponse:
			if msg.RevokedCertificateId == nil {
				msg.RevokedCertificateId = make([]string, 0, 1)
			}
			msg.RevokedCertificateId = append(msg.RevokedCertificateId, id)
		case *cert.CertificateRevocationError:
			msg.CertificateId = id
		}
		return nil
	}
}

func CaCertificate(opts ...CertOption) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.LoadCertificateRequest:
			m := new(cert.Certificate)
			err := apply(m, opts...)
			if err != nil {
				return err
			}
			if len(msg.CaCertificates) == 0 {
				msg.CaCertificates = make([]*cert.Certificate, 0, 1)
			}
			msg.CaCertificates = append(msg.CaCertificates, m)
		case *cert.LoadCertificateAuthorityBundleRequest:
			m := new(cert.Certificate)
			err := apply(m, opts...)
			if err != nil {
				return err
			}
			if len(msg.CaCertificates) == 0 {
				msg.CaCertificates = make([]*cert.Certificate, 0)
			}
			msg.CaCertificates = append(msg.CaCertificates, m)
		}
		return nil
	}
}

func ErrorMsg(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CertificateRevocationError:
			msg.ErrorMessage = s
		}
		return nil
	}
}

func CSRParams(opts ...CertOption) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.GenerateCSRRequest:
			m := new(cert.CSRParams)
			err := apply(m, opts...)
			if err != nil {
				return err
			}
			msg.CsrParams = m
		}
		return nil
	}
}

func CSR(opts ...CertOption) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.GenerateCSRResponse:
			m := new(cert.CSR)
			err := apply(m, opts...)
			if err != nil {
				return err
			}
			msg.Csr = m
		}
		return nil
	}
}

func KeySize(ks uint32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
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

func KeyType(kt string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CanGenerateCSRRequest:
			ktv, ok := cert.KeyType_value[strings.ToUpper(kt)]
			if !ok {
				return api.ErrInvalidValue
			}
			msg.KeyType = cert.KeyType(ktv)
		}
		return nil
	}
}

func CommonName(cn string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CSRParams:
			msg.CommonName = cn
		}
		return nil
	}
}

func Country(c string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CSRParams:
			msg.Country = c
		}
		return nil
	}
}

func State(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CSRParams:
			msg.State = s
		}
		return nil
	}
}

func City(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CSRParams:
			msg.City = s
		}
		return nil
	}
}

func Org(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CSRParams:
			msg.Organization = s
		}
		return nil
	}
}

func OrgUnit(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CSRParams:
			msg.OrganizationalUnit = s
		}
		return nil
	}
}

func IPAddress(ipAddr string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CSRParams:
			msg.IpAddress = ipAddr
		}
		return nil
	}
}

func EmailID(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CSRParams:
			msg.EmailId = s
		}
		return nil
	}
}

func Endpoint(typ cert.Endpoint_Type, endp string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CertificateInfo:
			if msg.Endpoints == nil {
				msg.Endpoints = make([]*cert.Endpoint, 0, 1)
			}
			msg.Endpoints = append(msg.Endpoints, &cert.Endpoint{
				Type:     typ,
				Endpoint: endp,
			})
		}
		return nil
	}
}

func ModificationTime(mt int64) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CertificateInfo:
			msg.ModificationTime = mt
		}
		return nil
	}
}
