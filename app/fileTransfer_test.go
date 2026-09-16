package app

import (
	"strings"
	"testing"

	"github.com/openconfig/gnoi/common"
	"github.com/openconfig/gnoi/file"
	"github.com/openconfig/gnoi/types"
)

func TestTransferFileRemoteDownload(t *testing.T) {
	tests := []struct {
		name     string
		remote   string
		source   string
		wantProt common.RemoteDownload_Protocol
		wantPath string
		wantUser string
		wantPass string
		wantErr  bool
	}{
		{
			name:     "scp with credentials",
			remote:   "scp://user:pass@server.com:/path/to/file",
			source:   "10.0.0.1",
			wantProt: common.RemoteDownload_SCP,
			wantPath: "server.com:/path/to/file",
			wantUser: "user",
			wantPass: "pass",
		},
		{
			name:     "sftp with credentials containing colon in password",
			remote:   "sftp://user:pa:ss@server.com:/path",
			wantProt: common.RemoteDownload_SFTP,
			wantPath: "server.com:/path",
			wantUser: "user",
			wantPass: "pa:ss",
		},
		{
			name:     "http without credentials",
			remote:   "http://server.com/path/to/file",
			wantProt: common.RemoteDownload_HTTP,
			wantPath: "server.com/path/to/file",
		},
		{
			name:     "https without credentials",
			remote:   "https://server.com/file",
			wantProt: common.RemoteDownload_HTTPS,
			wantPath: "server.com/file",
		},
		{
			name:    "missing protocol",
			remote:  "server.com:/path",
			wantErr: true,
		},
		{
			name:    "credentials without password",
			remote:  "scp://user@server.com:/path",
			wantErr: true,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			a := newTestApp(t)
			a.Config.FileTransferRemote = tt.remote
			a.Config.FileTransferSourceAddress = tt.source
			rd, err := a.transferFileRemoteDownload()
			if tt.wantErr {
				if err == nil {
					t.Fatalf("expected error, got %v", rd)
				}
				return
			}
			if err != nil {
				t.Fatal(err)
			}
			if rd.GetProtocol() != tt.wantProt {
				t.Errorf("Protocol = %v, want %v", rd.GetProtocol(), tt.wantProt)
			}
			if rd.GetPath() != tt.wantPath {
				t.Errorf("Path = %q, want %q", rd.GetPath(), tt.wantPath)
			}
			if rd.GetSourceAddress() != tt.source {
				t.Errorf("SourceAddress = %q, want %q", rd.GetSourceAddress(), tt.source)
			}
			if tt.wantUser == "" {
				if rd.GetCredentials() != nil {
					t.Errorf("unexpected credentials %v", rd.GetCredentials())
				}
				return
			}
			if rd.GetCredentials().GetUsername() != tt.wantUser || rd.GetCredentials().GetCleartext() != tt.wantPass {
				t.Errorf("credentials = %v, want %s:%s", rd.GetCredentials(), tt.wantUser, tt.wantPass)
			}
		})
	}
}

func TestTransferTable(t *testing.T) {
	a := newTestApp(t)
	out := a.transferTable([]*fileTransferResponse{
		{
			TargetError: TargetError{TargetName: "r1"},
			rsp: &file.TransferToRemoteResponse{
				Hash: &types.HashType{Method: types.HashType_SHA256, Hash: []byte{0xde, 0xad}},
			},
		},
	})
	for _, want := range []string{"Target Name", "Hash Method", "r1", "SHA256", "dead"} {
		if !strings.Contains(out, want) {
			t.Errorf("table output missing %q:\n%s", want, out)
		}
	}
}
