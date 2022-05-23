package file

import gnoifile "github.com/openconfig/gnoi/file"

func NewTransferRequest(opts ...FileOption) (*gnoifile.TransferToRemoteRequest, error) {
	m := new(gnoifile.TransferToRemoteRequest)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewTransferResponse(opts ...FileOption) (*gnoifile.TransferToRemoteResponse, error) {
	m := new(gnoifile.TransferToRemoteResponse)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
