package os

import gnoios "github.com/openconfig/gnoi/os"

func NewActivateRequest(opts ...OsOption) (*gnoios.ActivateRequest, error) {
	m := new(gnoios.ActivateRequest)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewActivateOKResponse() *gnoios.ActivateResponse {
	return &gnoios.ActivateResponse{
		Response: &gnoios.ActivateResponse_ActivateOk{
			ActivateOk: &gnoios.ActivateOK{},
		},
	}
}

func NewActivateErrorResponse(opts ...OsOption) (*gnoios.ActivateResponse, error) {
	m := new(gnoios.ActivateError)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return &gnoios.ActivateResponse{
		Response: &gnoios.ActivateResponse_ActivateError{
			ActivateError: m,
		},
	}, nil
}
