package file

import gnoifile "github.com/openconfig/gnoi/file"

func NewStatRequest(opts ...FileOption) (*gnoifile.StatRequest, error) {
	m := new(gnoifile.StatRequest)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewStatResponse(opts ...FileOption) (*gnoifile.StatResponse, error) {
	m := new(gnoifile.StatResponse)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
