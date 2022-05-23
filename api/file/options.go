package file

import (
	"fmt"
	"strings"

	"github.com/karimra/gnoic/api"
	"github.com/openconfig/gnoi/common"
	gnoifile "github.com/openconfig/gnoi/file"
	"github.com/openconfig/gnoi/types"
	"google.golang.org/protobuf/proto"
)

type FileOption func(proto.Message) error

// apply is a helper function that simply applies the options to the proto.Message.
// It returns an error if any of the options fails.
func apply(m proto.Message, opts ...FileOption) error {
	for _, o := range opts {
		if err := o(m); err != nil {
			return err
		}
	}
	return nil
}

func FileName(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option FileName: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoifile.GetRequest:
			msg.RemoteFile = s
		case *gnoifile.RemoveRequest:
			msg.RemoteFile = s
		case *gnoifile.PutRequest:
			switch m := msg.GetRequest().(type) {
			case *gnoifile.PutRequest_Open:
				m.Open.RemoteFile = s
			default:
				return api.ErrInvalidMsgType
			}
		case *gnoifile.TransferToRemoteRequest:
			msg.LocalPath = s
		default:
			return fmt.Errorf("option FileName: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Content(b []byte) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Content: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoifile.GetResponse:
			switch m := msg.GetResponse().(type) {
			case *gnoifile.GetResponse_Contents:
				m.Contents = b
			default:
				return fmt.Errorf("option Content: %w", api.ErrInvalidMsgType)
			}
		case *gnoifile.PutRequest:
			switch m := msg.GetRequest().(type) {
			case *gnoifile.PutRequest_Contents:
				m.Contents = b
			default:
				return fmt.Errorf("option Content: %w", api.ErrInvalidMsgType)
			}
		default:
			return fmt.Errorf("option Content: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Hash(method string, b []byte) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Hash: %w", api.ErrInvalidMsgType)
		}
		ht, ok := types.HashType_HashMethod_value[strings.ToUpper(method)]
		if !ok {
			return api.ErrInvalidValue
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoifile.GetResponse:
			switch rsp := msg.GetResponse().(type) {
			case *gnoifile.GetResponse_Hash:
				rsp.Hash = &types.HashType{
					Method: types.HashType_HashMethod(ht),
					Hash:   b,
				}
			default:
				return fmt.Errorf("option Hash: %w", api.ErrInvalidMsgType)
			}
		case *gnoifile.PutRequest:
			switch m := msg.GetRequest().(type) {
			case *gnoifile.PutRequest_Hash:
				m.Hash = &types.HashType{
					Method: types.HashType_HashMethod(ht),
					Hash:   b,
				}
			default:
				return fmt.Errorf("option Hash: %w", api.ErrInvalidMsgType)
			}
		case *types.Credentials:
			msg.Password = &types.Credentials_Hashed{
				Hashed: &types.HashType{
					Method: types.HashType_HashMethod(ht),
					Hash:   b,
				},
			}
		case *gnoifile.TransferToRemoteResponse:
			msg.Hash = &types.HashType{
				Method: types.HashType_HashMethod(ht),
				Hash:   b,
			}
		default:
			return fmt.Errorf("option Hash: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func HashMD5(b []byte) func(msg proto.Message) error {
	return Hash("MD5", b)
}

func HashSHA256(b []byte) func(msg proto.Message) error {
	return Hash("SHA256", b)
}

func HashSHA512(b []byte) func(msg proto.Message) error {
	return Hash("SHA512", b)
}

func HashUNSPECIFIED(b []byte) func(msg proto.Message) error {
	return Hash("UNSPECIFIED", b)
}

func StatInfo(opts ...FileOption) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option StatInfo: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoifile.StatResponse:
			if msg.Stats == nil {
				msg.Stats = make([]*gnoifile.StatInfo, 0, 1)
			}
			m := new(gnoifile.StatInfo)
			err := apply(m, opts...)
			if err != nil {
				return err
			}
			msg.Stats = append(msg.Stats, m)
		default:
			return fmt.Errorf("option StatInfo: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Path(p string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Path: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoifile.StatRequest:
			msg.Path = p
		case *gnoifile.StatInfo:
			msg.Path = p
		case *common.RemoteDownload:
			msg.Path = p
		default:
			return fmt.Errorf("option Path: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func LastModified(i uint64) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option LastModified: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoifile.StatInfo:
			msg.LastModified = i
		default:
			return fmt.Errorf("option LastModified: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Permissions(i uint32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Permissions: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoifile.StatInfo:
			msg.Permissions = i
		case *gnoifile.PutRequest:
			switch rsp := msg.GetRequest().(type) {
			case *gnoifile.PutRequest_Open:
				rsp.Open.Permissions = i
			default:
				return fmt.Errorf("option Permissions: %w", api.ErrInvalidMsgType)
			}
		default:
			return fmt.Errorf("option Permissions: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Size(i uint64) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Size: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoifile.StatInfo:
			msg.Size = i
		default:
			return fmt.Errorf("option Size: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Umask(i uint32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Umask: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoifile.StatInfo:
			msg.Umask = i
		default:
			return fmt.Errorf("option Umask: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func RemoteDownloadProtocol(p string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option RemoteDownloadProtocol: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *common.RemoteDownload:
			v, ok := common.RemoteDownload_Protocol_value[strings.ToUpper(p)]
			if !ok {
				return api.ErrInvalidValue
			}
			msg.Protocol = common.RemoteDownload_Protocol(v)
		default:
			return fmt.Errorf("option RemoteDownloadProtocol: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func RemoteDownloadProtocolSFTP() func(msg proto.Message) error {
	return RemoteDownloadProtocol("SFTP")
}

func RemoteDownloadProtocolHTTP() func(msg proto.Message) error {
	return RemoteDownloadProtocol("HTTP")
}

func RemoteDownloadProtocolHTTPS() func(msg proto.Message) error {
	return RemoteDownloadProtocol("HTTPS")
}

func RemoteDownloadProtocolSCP() func(msg proto.Message) error {
	return RemoteDownloadProtocol("SCP")
}

func RemoteDownloadProtocolCustom(i uint32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option RemoteDownloadProtocolCustom: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *common.RemoteDownload:
			msg.Protocol = common.RemoteDownload_Protocol(i)
		default:
			return fmt.Errorf("option RemoteDownloadProtocolCustom: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Credentials(opts ...FileOption) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Credentials: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *common.RemoteDownload:
			msg.Credentials = new(types.Credentials)
			return apply(msg.Credentials, opts...)
		default:
			return fmt.Errorf("option Credentials: %w", api.ErrInvalidMsgType)
		}
	}
}

func Username(uname string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Username: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *types.Credentials:
			msg.Username = uname
		default:
			return fmt.Errorf("option Username: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Password(password string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Password: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *types.Credentials:
			msg.Password = &types.Credentials_Cleartext{
				Cleartext: password,
			}
		default:
			return fmt.Errorf("option Password: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func SourceAddress(saddr string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option SourceAddress: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *common.RemoteDownload:
			msg.SourceAddress = saddr
		default:
			return fmt.Errorf("option SourceAddress: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}
