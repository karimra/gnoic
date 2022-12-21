package system

import gnoisystem "github.com/openconfig/gnoi/system"

func NewSystemKillProcessRequest(opts ...SystemOption) (*gnoisystem.KillProcessRequest, error) {
	m := new(gnoisystem.KillProcessRequest)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
