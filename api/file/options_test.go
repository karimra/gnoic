package file

import (
	"errors"
	"testing"

	"github.com/openconfig/gnoi/common"
	gnoifile "github.com/openconfig/gnoi/file"
	"github.com/openconfig/gnoi/types"
	"google.golang.org/protobuf/proto"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/karimra/gnoic/api"
)

// every option must reject a nil message and a message of an unexpected type.
func TestOptions_InvalidMessage(t *testing.T) {
	opts := map[string]FileOption{
		"FileName":                     FileName("f"),
		"Content":                      Content([]byte("x")),
		"Hash":                         Hash("MD5", []byte("x")),
		"HashMD5":                      HashMD5([]byte("x")),
		"HashSHA256":                   HashSHA256([]byte("x")),
		"HashSHA512":                   HashSHA512([]byte("x")),
		"HashUNSPECIFIED":              HashUNSPECIFIED([]byte("x")),
		"StatInfo":                     StatInfo(),
		"Path":                         Path("p"),
		"LastModified":                 LastModified(1),
		"Permissions":                  Permissions(1),
		"Size":                         Size(1),
		"Umask":                        Umask(1),
		"RemoteDownloadProtocol":       RemoteDownloadProtocol("SFTP"),
		"RemoteDownloadProtocolSFTP":   RemoteDownloadProtocolSFTP(),
		"RemoteDownloadProtocolHTTP":   RemoteDownloadProtocolHTTP(),
		"RemoteDownloadProtocolHTTPS":  RemoteDownloadProtocolHTTPS(),
		"RemoteDownloadProtocolSCP":    RemoteDownloadProtocolSCP(),
		"RemoteDownloadProtocolCustom": RemoteDownloadProtocolCustom(1),
		"Credentials":                  Credentials(),
		"Username":                     Username("u"),
		"Password":                     Password("p"),
		"SourceAddress":                SourceAddress("1.1.1.1"),
	}
	for name, o := range opts {
		t.Run(name, func(t *testing.T) {
			if err := o(nil); !errors.Is(err, api.ErrInvalidMsgType) {
				t.Errorf("nil message: got %v, want ErrInvalidMsgType", err)
			}
			if err := o(&emptypb.Empty{}); !errors.Is(err, api.ErrInvalidMsgType) {
				t.Errorf("wrong message type: got %v, want ErrInvalidMsgType", err)
			}
		})
	}
}

func TestOptions_WrongOneof(t *testing.T) {
	// options that only apply to a specific oneof variant must fail on the others
	contents, _ := NewPutContentRequest()
	if err := FileName("f")(contents); !errors.Is(err, api.ErrInvalidMsgType) {
		t.Errorf("FileName on PutRequest_Contents: %v", err)
	}
	if err := Permissions(1)(contents); !errors.Is(err, api.ErrInvalidMsgType) {
		t.Errorf("Permissions on PutRequest_Contents: %v", err)
	}
	open, _ := NewPutOpenRequest()
	if err := Content([]byte("x"))(open); !errors.Is(err, api.ErrInvalidMsgType) {
		t.Errorf("Content on PutRequest_Open: %v", err)
	}
	if err := Hash("MD5", nil)(open); !errors.Is(err, api.ErrInvalidMsgType) {
		t.Errorf("Hash on PutRequest_Open: %v", err)
	}
	hashRsp, _ := NewGetHashResponse()
	if err := Content([]byte("x"))(hashRsp); !errors.Is(err, api.ErrInvalidMsgType) {
		t.Errorf("Content on GetResponse_Hash: %v", err)
	}
	contentRsp, _ := NewGetContentsResponse()
	if err := Hash("MD5", nil)(contentRsp); !errors.Is(err, api.ErrInvalidMsgType) {
		t.Errorf("Hash on GetResponse_Contents: %v", err)
	}
}

func TestOptions_InvalidValue(t *testing.T) {
	if err := Hash("CRC32", nil)(&types.Credentials{}); !errors.Is(err, api.ErrInvalidValue) {
		t.Errorf("Hash with unknown method: %v", err)
	}
	if err := RemoteDownloadProtocol("FTP")(&common.RemoteDownload{}); !errors.Is(err, api.ErrInvalidValue) {
		t.Errorf("RemoteDownloadProtocol with unknown protocol: %v", err)
	}
}

func TestNewGetRequest(t *testing.T) {
	req, err := NewGetRequest(FileName("/cf3/file.txt"))
	if err != nil {
		t.Fatal(err)
	}
	if req.GetRemoteFile() != "/cf3/file.txt" {
		t.Errorf("RemoteFile = %q", req.GetRemoteFile())
	}
	if _, err := NewGetRequest(Path("x")); err == nil {
		t.Error("expected error from inapplicable option")
	}
}

func TestNewGetResponses(t *testing.T) {
	c, err := NewGetContentsResponse(Content([]byte("hello")))
	if err != nil {
		t.Fatal(err)
	}
	if string(c.GetContents()) != "hello" {
		t.Errorf("Contents = %q", c.GetContents())
	}
	h, err := NewGetHashResponse(HashSHA256([]byte{1, 2}))
	if err != nil {
		t.Fatal(err)
	}
	if h.GetHash().GetMethod() != types.HashType_SHA256 || string(h.GetHash().GetHash()) != "\x01\x02" {
		t.Errorf("Hash = %v", h.GetHash())
	}
	if _, err := NewGetHashResponse(Content(nil)); err == nil {
		t.Error("expected error")
	}
	if _, err := NewGetContentsResponse(Hash("MD5", nil)); err == nil {
		t.Error("expected error")
	}
}

func TestNewPutRequests(t *testing.T) {
	open, err := NewPutOpenRequest(FileName("remote.bin"), Permissions(644))
	if err != nil {
		t.Fatal(err)
	}
	if open.GetOpen().GetRemoteFile() != "remote.bin" || open.GetOpen().GetPermissions() != 644 {
		t.Errorf("Open = %v", open.GetOpen())
	}
	contents, err := NewPutContentRequest(Content([]byte("abc")))
	if err != nil {
		t.Fatal(err)
	}
	if string(contents.GetContents()) != "abc" {
		t.Errorf("Contents = %q", contents.GetContents())
	}
	for method, want := range map[string]types.HashType_HashMethod{
		"md5": types.HashType_MD5, "SHA256": types.HashType_SHA256, "sha512": types.HashType_SHA512, "unspecified": types.HashType_UNSPECIFIED,
	} {
		h, err := NewPutHashRequest(Hash(method, []byte("h")))
		if err != nil {
			t.Fatalf("%s: %v", method, err)
		}
		if h.GetHash().GetMethod() != want {
			t.Errorf("%s: Method = %v, want %v", method, h.GetHash().GetMethod(), want)
		}
	}
	if _, err := NewPutHashRequest(Hash("nope", nil)); !errors.Is(err, api.ErrInvalidValue) {
		t.Errorf("unknown hash method: %v", err)
	}
	if _, err := NewPutOpenRequest(Content(nil)); err == nil {
		t.Error("expected error")
	}
}

func TestNewStatRequestResponse(t *testing.T) {
	req, err := NewStatRequest(Path("/cf3"))
	if err != nil {
		t.Fatal(err)
	}
	if req.GetPath() != "/cf3" {
		t.Errorf("Path = %q", req.GetPath())
	}
	rsp, err := NewStatResponse(
		StatInfo(Path("/cf3/a"), LastModified(10), Permissions(644), Size(20), Umask(22)),
		StatInfo(Path("/cf3/b")),
	)
	if err != nil {
		t.Fatal(err)
	}
	if len(rsp.GetStats()) != 2 {
		t.Fatalf("got %d stats, want 2", len(rsp.GetStats()))
	}
	want := &gnoifile.StatInfo{Path: "/cf3/a", LastModified: 10, Permissions: 644, Size: 20, Umask: 22}
	if !proto.Equal(rsp.GetStats()[0], want) {
		t.Errorf("stat[0] = %v, want %v", rsp.GetStats()[0], want)
	}
	if rsp.GetStats()[1].GetPath() != "/cf3/b" {
		t.Errorf("stat[1] = %v", rsp.GetStats()[1])
	}
	// errors from nested options propagate
	if _, err := NewStatResponse(StatInfo(FileName("x"))); !errors.Is(err, api.ErrInvalidMsgType) {
		t.Errorf("nested invalid option: %v", err)
	}
}

func TestNewTransferRequestResponse(t *testing.T) {
	req, err := NewTransferRequest(FileName("/cf3/local.bin"))
	if err != nil {
		t.Fatal(err)
	}
	if req.GetLocalPath() != "/cf3/local.bin" {
		t.Errorf("LocalPath = %q", req.GetLocalPath())
	}

	rd := new(common.RemoteDownload)
	err = apply(rd,
		Path("srv:/tmp/x"),
		RemoteDownloadProtocolSCP(),
		SourceAddress("10.0.0.1"),
		Credentials(Username("u"), Password("p")),
	)
	if err != nil {
		t.Fatal(err)
	}
	if rd.GetPath() != "srv:/tmp/x" || rd.GetProtocol() != common.RemoteDownload_SCP || rd.GetSourceAddress() != "10.0.0.1" {
		t.Errorf("RemoteDownload = %v", rd)
	}
	if rd.GetCredentials().GetUsername() != "u" || rd.GetCredentials().GetCleartext() != "p" {
		t.Errorf("Credentials = %v", rd.GetCredentials())
	}
	if err := Credentials(Path("x"))(rd); !errors.Is(err, api.ErrInvalidMsgType) {
		t.Errorf("nested invalid credentials option: %v", err)
	}
	// hashed credentials
	creds := new(types.Credentials)
	if err := HashSHA512([]byte("h"))(creds); err != nil {
		t.Fatal(err)
	}
	if creds.GetHashed().GetMethod() != types.HashType_SHA512 {
		t.Errorf("Credentials hashed = %v", creds.GetHashed())
	}
	for _, tt := range []struct {
		opt  FileOption
		want common.RemoteDownload_Protocol
	}{
		{RemoteDownloadProtocolSFTP(), common.RemoteDownload_SFTP},
		{RemoteDownloadProtocolHTTP(), common.RemoteDownload_HTTP},
		{RemoteDownloadProtocolHTTPS(), common.RemoteDownload_HTTPS},
		{RemoteDownloadProtocolCustom(42), common.RemoteDownload_Protocol(42)},
	} {
		rd := new(common.RemoteDownload)
		if err := tt.opt(rd); err != nil {
			t.Fatal(err)
		}
		if rd.GetProtocol() != tt.want {
			t.Errorf("Protocol = %v, want %v", rd.GetProtocol(), tt.want)
		}
	}

	rsp, err := NewTransferResponse(HashMD5([]byte("h")))
	if err != nil {
		t.Fatal(err)
	}
	if rsp.GetHash().GetMethod() != types.HashType_MD5 {
		t.Errorf("Hash = %v", rsp.GetHash())
	}
}

func TestFileName_RemoveRequest(t *testing.T) {
	req := new(gnoifile.RemoveRequest)
	if err := FileName("/cf3/x")(req); err != nil {
		t.Fatal(err)
	}
	if req.GetRemoteFile() != "/cf3/x" {
		t.Errorf("RemoteFile = %q", req.GetRemoteFile())
	}
}
