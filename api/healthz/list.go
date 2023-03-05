package healthz

import gnoihealthz "github.com/openconfig/gnoi/healthz"

func NewListRequest(opts ...HealthzOption) (*gnoihealthz.ListRequest, error) {
	m := &gnoihealthz.ListRequest{}
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewListResponse(opts ...HealthzOption) (*gnoihealthz.ListResponse, error) {
	m := &gnoihealthz.ListResponse{}
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
