package system

import gnoisystem "github.com/openconfig/gnoi/system"

func NewSystemPingRequest(opts ...SystemOption) (*gnoisystem.PingRequest, error) {
	m := new(gnoisystem.PingRequest)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewSystemPingResponse(opts ...SystemOption) (*gnoisystem.PingResponse, error) {
	m := new(gnoisystem.PingResponse)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
