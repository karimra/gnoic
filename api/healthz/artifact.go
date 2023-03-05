package healthz

import gnoihealthz "github.com/openconfig/gnoi/healthz"

func NewArtifactRequest(opts ...HealthzOption) (*gnoihealthz.ArtifactRequest, error) {
	m := &gnoihealthz.ArtifactRequest{}
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewArtifactResponse(opts ...HealthzOption) (*gnoihealthz.ArtifactResponse, error) {
	m := &gnoihealthz.ArtifactResponse{}
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
