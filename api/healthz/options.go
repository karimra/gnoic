package healthz

import (
	"fmt"
	"strings"
	"time"

	"github.com/karimra/gnoic/api"
	"github.com/karimra/gnoic/utils"
	gnoihealthz "github.com/openconfig/gnoi/healthz"
	"github.com/openconfig/gnoi/types"
	"google.golang.org/protobuf/proto"
	timestamppb "google.golang.org/protobuf/types/known/timestamppb"
)

type HealthzOption func(proto.Message) error

// apply is a helper function that simply applies the options to the proto.Message.
// It returns an error if any of the options fails.
func apply(m proto.Message, opts ...HealthzOption) error {
	for _, o := range opts {
		if err := o(m); err != nil {
			return err
		}
	}
	return nil
}

func Path(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Path: %w", api.ErrInvalidMsgType)
		}
		var err error
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoihealthz.GetRequest:
			msg.Path, err = utils.ParsePath(s)
		case *gnoihealthz.AcknowledgeRequest:
			msg.Path, err = utils.ParsePath(s)
		case *gnoihealthz.ListRequest:
			msg.Path, err = utils.ParsePath(s)
		case *gnoihealthz.CheckRequest:
			msg.Path, err = utils.ParsePath(s)
		default:
			return fmt.Errorf("option Path: %w", api.ErrInvalidMsgType)
		}
		return err
	}
}

func ComponentStatus(opts ...HealthzOption) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option ComponentStatus: %w", api.ErrInvalidMsgType)
		}
		m := new(gnoihealthz.ComponentStatus)
		err := apply(m, opts...)
		if err != nil {
			return err
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoihealthz.GetResponse:
			msg.Component = m
		case *gnoihealthz.ComponentStatus:
			if msg.Subcomponents == nil {
				msg.Subcomponents = make([]*gnoihealthz.ComponentStatus, 0, 1)
			}
			msg.Subcomponents = append(msg.Subcomponents, m)
		case *gnoihealthz.ListResponse:
			if msg.Statuses == nil {
				msg.Statuses = make([]*gnoihealthz.ComponentStatus, 0, 1)
			}
			msg.Statuses = append(msg.Statuses, m)
		case *gnoihealthz.AcknowledgeResponse:
			msg.Status = m
		case *gnoihealthz.CheckResponse:
			msg.Status = m
		default:
			return fmt.Errorf("option ComponentStatus: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Status(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Status: %w", api.ErrInvalidMsgType)
		}
		st, ok := gnoihealthz.Status_value[strings.ToUpper(s)]
		if !ok {
			return api.ErrInvalidValue
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoihealthz.ComponentStatus:
			msg.Status = gnoihealthz.Status(st)
		default:
			return fmt.Errorf("option Status: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Status_UNSPECIFIED() func(msg proto.Message) error {
	return Status("UNSPECIFIED}")
}

func Status_HEALTHY() func(msg proto.Message) error {
	return Status("HEALTHY")
}

func Status_UNHEALTHY() func(msg proto.Message) error {
	return Status("UNHEALTHY")
}

func ArtifactHeader(opts ...HealthzOption) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option ArtifactHeader: %w", api.ErrInvalidMsgType)
		}
		m := new(gnoihealthz.ArtifactHeader)
		err := apply(m, opts...)
		if err != nil {
			return err
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoihealthz.ArtifactResponse:
			msg.Contents = &gnoihealthz.ArtifactResponse_Header{
				Header: m,
			}
		case *gnoihealthz.ComponentStatus:
			if msg.Artifacts == nil {
				msg.Artifacts = make([]*gnoihealthz.ArtifactHeader, 0, 1)
			}
			msg.Artifacts = append(msg.Artifacts, m)
		default:
			return fmt.Errorf("option ArtifactHeader: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

// Used for Id and EventId
func ID(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option ID: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoihealthz.ArtifactHeader:
			msg.Id = s
		case *gnoihealthz.ComponentStatus:
			msg.Id = s
		case *gnoihealthz.AcknowledgeRequest:
			msg.Id = s
		case *gnoihealthz.ArtifactRequest:
			msg.Id = s
		case *gnoihealthz.CheckRequest:
			msg.EventId = s
		default:
			return fmt.Errorf("option ID: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Name(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Name: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoihealthz.FileArtifactType:
			msg.Name = s
		default:
			return fmt.Errorf("option Name: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func SysPath(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option SysPath: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoihealthz.FileArtifactType:
			msg.Path = s
		default:
			return fmt.Errorf("option SysPath: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func MimeType(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option MimeType: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoihealthz.FileArtifactType:
			msg.Mimetype = s
		default:
			return fmt.Errorf("option MimeType: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Size(s int64) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Size: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoihealthz.FileArtifactType:
			msg.Size = s
		default:
			return fmt.Errorf("option Size: %w", api.ErrInvalidMsgType)
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
		case *gnoihealthz.FileArtifactType:
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

func Acknowledged(b bool) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Acknowledged: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoihealthz.ComponentStatus:
			msg.Acknowledged = b
		default:
			return fmt.Errorf("option Acknowledged: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Created(t time.Time) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Created: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoihealthz.ComponentStatus:
			msg.Created = timestamppb.New(t)
		default:
			return fmt.Errorf("option Created: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Expires(t time.Time) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Expires: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoihealthz.ComponentStatus:
			msg.Expires = timestamppb.New(t)
		default:
			return fmt.Errorf("option Expires: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func IncludeAcknowledged(b bool) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option IncludeAcknowledged: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoihealthz.ListRequest:
			msg.IncludeAcknowledged = b
		default:
			return fmt.Errorf("option IncludeAcknowledged: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}
