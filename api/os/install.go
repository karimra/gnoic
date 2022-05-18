package os

import (
	gnoios "github.com/openconfig/gnoi/os"
)

func NewOSInstallTransferRequest(opts ...OsOption) (*gnoios.InstallRequest, error) {
	m, err := NewOSTransferRequest(opts...)
	if err != nil {
		return nil, err
	}
	return &gnoios.InstallRequest{
		Request: &gnoios.InstallRequest_TransferRequest{
			TransferRequest: m,
		},
	}, nil

}

func NewOSInstallTransferContent(opts ...OsOption) (*gnoios.InstallRequest, error) {
	m := new(gnoios.InstallRequest)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewOSInstallTransferEnd() *gnoios.InstallRequest {
	return &gnoios.InstallRequest{
		Request: &gnoios.InstallRequest_TransferEnd{
			TransferEnd: &gnoios.TransferEnd{},
		},
	}
}

func NewOSInstallTransferReadyResponse() *gnoios.InstallResponse {
	return &gnoios.InstallResponse{
		Response: &gnoios.InstallResponse_TransferReady{
			TransferReady: &gnoios.TransferReady{},
		},
	}
}

func NewOSInstallTransferProgressResponse(opts ...OsOption) (*gnoios.InstallResponse, error) {
	m, err := NewOSTransferProgress(opts...)
	if err != nil {
		return nil, err
	}
	return &gnoios.InstallResponse{
		Response: &gnoios.InstallResponse_TransferProgress{
			TransferProgress: m,
		},
	}, nil
}

func NewOSInstallSyncProgressResponse(opts ...OsOption) (*gnoios.InstallResponse, error) {
	m, err := NewOSSyncProgress(opts...)
	if err != nil {
		return nil, err
	}
	return &gnoios.InstallResponse{
		Response: &gnoios.InstallResponse_SyncProgress{
			SyncProgress: m,
		},
	}, nil
}

func NewOSInstallValidatedResponse(opts ...OsOption) (*gnoios.InstallResponse, error) {
	m, err := NewOSValidated(opts...)
	if err != nil {
		return nil, err
	}
	return &gnoios.InstallResponse{
		Response: &gnoios.InstallResponse_Validated{
			Validated: m,
		},
	}, nil
}

func NewOSInstallInstallErrorResponse(opts ...OsOption) (*gnoios.InstallResponse, error) {
	m, err := NewOSInstallError(opts...)
	if err != nil {
		return nil, err
	}
	return &gnoios.InstallResponse{
		Response: &gnoios.InstallResponse_InstallError{
			InstallError: m,
		},
	}, nil
}

func NewOSTransferRequest(opts ...OsOption) (*gnoios.TransferRequest, error) {
	m := new(gnoios.TransferRequest)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewOSTransferProgress(opts ...OsOption) (*gnoios.TransferProgress, error) {
	m := new(gnoios.TransferProgress)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewOSSyncProgress(opts ...OsOption) (*gnoios.SyncProgress, error) {
	m := new(gnoios.SyncProgress)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewOSValidated(opts ...OsOption) (*gnoios.Validated, error) {
	m := new(gnoios.Validated)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}

func NewOSInstallError(opts ...OsOption) (*gnoios.InstallError, error) {
	m := new(gnoios.InstallError)
	err := apply(m, opts...)
	if err != nil {
		return nil, err
	}
	return m, nil
}
