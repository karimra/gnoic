package system

import gnoisystem "github.com/openconfig/gnoi/system"

func NewSetPackagePackageRequest(opts ...SystemOption) (*gnoisystem.SetPackageRequest, error) {
	m := &gnoisystem.SetPackageRequest{
		Request: &gnoisystem.SetPackageRequest_Package{
			Package: &gnoisystem.Package{},
		},
	}
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewSetPackageHashRequest(opts ...SystemOption) (*gnoisystem.SetPackageRequest, error) {
	m := &gnoisystem.SetPackageRequest{
		Request: &gnoisystem.SetPackageRequest_Hash{},
	}

	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
