package file

import gnoifile "github.com/openconfig/gnoi/file"

func NewPutOpenRequest(opts ...FileOption) (*gnoifile.PutRequest, error) {
	m := &gnoifile.PutRequest{
		Request: &gnoifile.PutRequest_Open{
			Open: &gnoifile.PutRequest_Details{},
		},
	}
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewPutContentRequest(opts ...FileOption) (*gnoifile.PutRequest, error) {
	m := &gnoifile.PutRequest{
		Request: &gnoifile.PutRequest_Contents{},
	}
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewPutHashRequest(opts ...FileOption) (*gnoifile.PutRequest, error) {
	m := &gnoifile.PutRequest{
		Request: &gnoifile.PutRequest_Hash{},
	}
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
