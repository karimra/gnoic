package cert

import (
	"fmt"
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
			return fmt.Errorf("option CertificateType: %w", api.ErrInvalidMsgType)
		}
		ctv, ok := cert.CertificateType_value[strings.ToUpper(ct)]
		if !ok {
			return fmt.Errorf("option CertificateType: %w", api.ErrInvalidValue)
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
		default:
			return fmt.Errorf("option CertificateType: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func CertificateTypeX509() func(msg proto.Message) error {
	return CertificateType("CT_X509")
}

func CertificateBytes(b []byte) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option CertificateBytes: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.Certificate:
			msg.Certificate = b
		default:
			return fmt.Errorf("option CertificateBytes: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func CertificateInfo(opts ...CertOption) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option CertificateInfo: %w", api.ErrInvalidMsgType)
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
		default:
			return fmt.Errorf("option CertificateInfo: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Certificate(opts ...CertOption) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Certificate: %w", api.ErrInvalidMsgType)
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
		default:
			return fmt.Errorf("option Certificate: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func CertificateID(id string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option CertificateID: %w", api.ErrInvalidMsgType)
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
		default:
			return fmt.Errorf("option CertificateID: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func CaCertificate(opts ...CertOption) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option CaCertificate: %w", api.ErrInvalidMsgType)
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
				msg.CaCertificates = make([]*cert.Certificate, 0, 1)
			}
			msg.CaCertificates = append(msg.CaCertificates, m)
		default:
			return fmt.Errorf("option CaCertificate: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func ErrorMsg(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option ErrorMsg: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CertificateRevocationError:
			msg.ErrorMessage = s
		default:
			return fmt.Errorf("option ErrorMsg: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func CSRParams(opts ...CertOption) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option CSRParams: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.GenerateCSRRequest:
			m := new(cert.CSRParams)
			err := apply(m, opts...)
			if err != nil {
				return err
			}
			msg.CsrParams = m
		default:
			return fmt.Errorf("option CSRParams: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func CSR(opts ...CertOption) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option CSR: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.GenerateCSRResponse:
			m := new(cert.CSR)
			err := apply(m, opts...)
			if err != nil {
				return err
			}
			msg.Csr = m
		default:
			return fmt.Errorf("option CSR: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func KeySize(ks uint32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option KeySize: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CanGenerateCSRRequest:
			msg.KeySize = ks
		case *cert.CSRParams:
			msg.MinKeySize = ks
		default:
			return fmt.Errorf("option KeySize: %w", api.ErrInvalidMsgType)
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
			return fmt.Errorf("option KeyType: %w", api.ErrInvalidMsgType)
		}
		ktv, ok := cert.KeyType_value[strings.ToUpper(kt)]
		if !ok {
			return api.ErrInvalidValue
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CanGenerateCSRRequest:
			msg.KeyType = cert.KeyType(ktv)
		case *cert.CSRParams:
			msg.KeyType = cert.KeyType(ktv)
		default:
			return fmt.Errorf("option KeyType: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func KeyPair(opts ...CertOption) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option KeyPair: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.LoadCertificateRequest:
			m := new(cert.KeyPair)
			err := apply(m, opts...)
			if err != nil {
				return err
			}
			msg.KeyPair = m
		default:
			return fmt.Errorf("option KeyPair: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func PublicKey(pubKey []byte) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option PublicKey: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.KeyPair:
			msg.PublicKey = pubKey
		default:
			return fmt.Errorf("option PublicKey: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func PrivateKey(privKey []byte) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option PrivateKey: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.KeyPair:
			msg.PrivateKey = privKey
		default:
			return fmt.Errorf("option PrivateKey: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func CommonName(cn string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option CommonName: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CSRParams:
			msg.CommonName = cn
		default:
			return fmt.Errorf("option CommonName: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Country(c string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Country: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CSRParams:
			msg.Country = c
		default:
			return fmt.Errorf("option Country: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func State(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option State: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CSRParams:
			msg.State = s
		default:
			return fmt.Errorf("option State: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func City(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option City: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CSRParams:
			msg.City = s
		default:
			return fmt.Errorf("option City: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Org(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Org: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CSRParams:
			msg.Organization = s
		default:
			return fmt.Errorf("option Org: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func OrgUnit(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option OrgUnit: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CSRParams:
			msg.OrganizationalUnit = s
		default:
			return fmt.Errorf("option OrgUnit: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func IPAddress(ipAddr string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option IPAddress: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CSRParams:
			msg.IpAddress = ipAddr
		default:
			return fmt.Errorf("option IPAddress: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func EmailID(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option EmailID: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CSRParams:
			msg.EmailId = s
		default:
			return fmt.Errorf("option EmailID: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Endpoint(typ cert.Endpoint_Type, endp string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Endpoint: %w", api.ErrInvalidMsgType)
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
		default:
			return fmt.Errorf("option Endpoint: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func ModificationTime(mt int64) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option ModificationTime: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *cert.CertificateInfo:
			msg.ModificationTime = mt
		default:
			return fmt.Errorf("option ModificationTime: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}
