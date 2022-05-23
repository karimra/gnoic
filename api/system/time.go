package system

import gnoisystem "github.com/openconfig/gnoi/system"

func NewSystemTimeRequest(opts ...SystemOption) *gnoisystem.TimeRequest {
	return new(gnoisystem.TimeRequest)
}

func NewSystemTimeResponse(opts ...SystemOption) (*gnoisystem.TimeResponse, error) {
	m := new(gnoisystem.TimeResponse)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
