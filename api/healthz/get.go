package healthz

import gnoihealthz "github.com/openconfig/gnoi/healthz"

func NewGetRequest(opts ...HealthzOption) (*gnoihealthz.GetRequest, error) {
	m := &gnoihealthz.GetRequest{}
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewGetResponse(opts ...HealthzOption) (*gnoihealthz.GetResponse, error) {
	m := &gnoihealthz.GetResponse{}
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
