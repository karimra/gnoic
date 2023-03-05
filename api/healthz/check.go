package healthz

import gnoihealthz "github.com/openconfig/gnoi/healthz"

func NewCheckRequest(opts ...HealthzOption) (*gnoihealthz.CheckRequest, error) {
	m := &gnoihealthz.CheckRequest{}
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewCheckResponse(opts ...HealthzOption) (*gnoihealthz.CheckResponse, error) {
	m := &gnoihealthz.CheckResponse{}
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
