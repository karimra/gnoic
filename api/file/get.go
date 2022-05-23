package file

import gnoifile "github.com/openconfig/gnoi/file"

func NewGetRequest(opts ...FileOption) (*gnoifile.GetRequest, error) {
	m := new(gnoifile.GetRequest)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewGetContentsResponse(opts ...FileOption) (*gnoifile.GetResponse, error) {
	m := &gnoifile.GetResponse{
		Response: &gnoifile.GetResponse_Contents{},
	}
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewGetHashResponse(opts ...FileOption) (*gnoifile.GetResponse, error) {
	m := &gnoifile.GetResponse{
		Response: &gnoifile.GetResponse_Hash{},
	}
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
