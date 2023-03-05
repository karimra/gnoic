package os

import (
	"github.com/karimra/gnoic/api"
	gnoios "github.com/openconfig/gnoi/os"
	"google.golang.org/protobuf/proto"
)

func NewOSVerifyRequest() *gnoios.VerifyRequest {
	return new(gnoios.VerifyRequest)
}

func NewOSVerifyResponse(opts ...OsOption) (*gnoios.VerifyResponse, error) {
	m := new(gnoios.VerifyResponse)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func VerifyStandbyState(s gnoios.StandbyState_State) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}

		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoios.VerifyResponse:
			msg.VerifyStandby = &gnoios.VerifyStandby{
				State: &gnoios.VerifyStandby_StandbyState{
					StandbyState: &gnoios.StandbyState{
						State: s,
					},
				},
			}
		}
		return nil
	}
}

func VerifyStandbyStateUNSUPPORTED() func(msg proto.Message) error {
	return VerifyStandbyState(gnoios.StandbyState_UNSUPPORTED)
}

func VerifyStandbyStateNON_EXISTENT() func(msg proto.Message) error {
	return VerifyStandbyState(gnoios.StandbyState_NON_EXISTENT)
}

func VerifyStandbyStateUNAVAILABLE() func(msg proto.Message) error {
	return VerifyStandbyState(gnoios.StandbyState_UNAVAILABLE)
}

func VerifyStandbyResponse(opts ...OsOption) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}

		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoios.VerifyResponse:
			m := new(gnoios.StandbyResponse)
			err := apply(m, opts...)
			if err != nil {
				return err
			}
			msg.VerifyStandby = &gnoios.VerifyStandby{
				State: &gnoios.VerifyStandby_VerifyResponse{
					VerifyResponse: m,
				},
			}
		}
		return nil
	}
}

func StandbyResponseID(id string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return api.ErrInvalidMsgType
		}

		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoios.StandbyResponse:
			msg.Id = id
		}
		return nil
	}
}
