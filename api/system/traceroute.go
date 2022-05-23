package system

import gnoisystem "github.com/openconfig/gnoi/system"

func NewSystemTracerouteRequest(opts ...SystemOption) (*gnoisystem.TracerouteRequest, error) {
	m := new(gnoisystem.TracerouteRequest)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewSystemTracerouteResponse(opts ...SystemOption) (*gnoisystem.TracerouteResponse, error) {
	m := new(gnoisystem.TracerouteResponse)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
