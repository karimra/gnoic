package os

import (
	"github.com/karimra/gnoic/api"
	gnoios "github.com/openconfig/gnoi/os"
	"google.golang.org/protobuf/proto"
)

type OsOption func(proto.Message) error

// apply is a helper function that simply applies the options to the proto.Message.
// It returns an error if any of the options fails.
func apply(m proto.Message, opts ...OsOption) error {
	for _, o := range opts {
		if err := o(m); err != nil {
			return err
		}
	}
	return nil
}

func Version(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}

		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoios.TransferRequest:
			msg.Version = s
		case *gnoios.ActivateRequest:
			msg.Version = s
		case *gnoios.VerifyResponse:
			msg.Version = s
		case *gnoios.StandbyResponse:
			msg.Version = s
		case *gnoios.Validated:
			msg.Version = s
		}
		return nil
	}
}

func Description(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}

		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoios.Validated:
			msg.Description = s
		}
		return nil
	}
}

func StandbySupervisor(b bool) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}

		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoios.TransferRequest:
			msg.StandbySupervisor = b
		case *gnoios.ActivateRequest:
			msg.StandbySupervisor = b
		}
		return nil
	}
}

func NoReboot(b bool) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}

		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoios.ActivateRequest:
			msg.NoReboot = b
		}
		return nil
	}
}

func ErrorType(e int32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}

		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoios.ActivateError:
			msg.Type = gnoios.ActivateError_Type(e)
		case *gnoios.InstallError:
			msg.Type = gnoios.InstallError_Type(e)
		}
		return nil
	}
}

func ErrorDetail(d string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}

		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoios.ActivateError:
			msg.Detail = d
		case *gnoios.InstallError:
			msg.Detail = d
		}
		return nil
	}
}

func ActivationFailMsg(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}

		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoios.VerifyResponse:
			msg.ActivationFailMessage = s
		case *gnoios.StandbyResponse:
			msg.ActivationFailMessage = s
		}
		return nil
	}
}

func BytesReceived(i uint64) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}

		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoios.TransferProgress:
			msg.BytesReceived = i
		}
		return nil
	}
}

func PercentageTransferred(i uint32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}

		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoios.SyncProgress:
			msg.PercentageTransferred = i
		}
		return nil
	}
}
