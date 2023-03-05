package healthz

import gnoihealthz "github.com/openconfig/gnoi/healthz"

func NewAcknowledgeRequest(opts ...HealthzOption) (*gnoihealthz.AcknowledgeRequest, error) {
	m := &gnoihealthz.AcknowledgeRequest{}
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewAcknowledgeResponse(opts ...HealthzOption) (*gnoihealthz.AcknowledgeResponse, error) {
	m := &gnoihealthz.AcknowledgeResponse{}
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
