package app

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/rand"
	"io"
	"net"
	"os"
	"path/filepath"
	"testing"
	"time"

	"github.com/openconfig/gnoi/file"
	"github.com/openconfig/gnoi/types"
	log "github.com/sirupsen/logrus"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/karimra/gnoic/api"
	gfile "github.com/karimra/gnoic/api/file"
	"github.com/karimra/gnoic/config"
)

// startFileServer runs the gNOI File server on a unix socket rooted at
// rootDir and returns a gnoic target address for it.
func startFileServer(t *testing.T, rootDir, hashMethod string) string {
	t.Helper()
	sock := filepath.Join(t.TempDir(), "gnoic.sock")
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	logger := log.New()
	logger.SetOutput(io.Discard)
	fs := &fserver{
		logger:         log.NewEntry(logger),
		s:              grpc.NewServer(),
		rootDir:        rootDir,
		fileHashMethod: hashMethod,
	}
	file.RegisterFileServer(fs.s, fs)
	go fs.s.Serve(l)
	t.Cleanup(fs.s.Stop)
	return "unix://" + sock
}

// connectTarget creates a connected api.Target for addr.
func connectTarget(t *testing.T, ctx context.Context, addr string) *api.Target {
	t.Helper()
	tg := api.NewTargetFromConfig(&config.TargetConfig{
		Name:       "test-target",
		Address:    addr,
		Insecure:   ptr(true),
		SkipVerify: ptr(false),
		Username:   ptr(""),
		Password:   ptr(""),
		Timeout:    5 * time.Second,
	})
	if err := tg.CreateGrpcClient(ctx); err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() { tg.Close() })
	return tg
}

func randomBytes(t *testing.T, n int) []byte {
	t.Helper()
	b := make([]byte, n)
	if _, err := rand.Read(b); err != nil {
		t.Fatal(err)
	}
	return b
}

func TestFileServer_PutStatGetRemove(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	serverRoot := t.TempDir()
	tg := connectTarget(t, ctx, startFileServer(t, serverRoot, "md5"))

	// --- Put: local file -> server, chunked, with per-file permissions
	localDir := t.TempDir()
	content := randomBytes(t, 100*1024+17) // not a multiple of the chunk size
	localFile := filepath.Join(localDir, "image.bin")
	if err := os.WriteFile(localFile, content, 0600); err != nil {
		t.Fatal(err)
	}

	a := newTestApp(t)
	a.Config.FilePutFile = []string{localFile}
	a.Config.FilePutDst = "cf3/uploads/image.bin"
	a.Config.FilePutChunkSize = 8 * 1024
	a.Config.FilePutPermissions = 644
	a.Config.FilePutHashMethod = "MD5"

	written, err := a.FilePut(ctx, tg)
	if err != nil {
		t.Fatalf("FilePut: %v", err)
	}
	if len(written) != 1 || written[0] != localFile {
		t.Errorf("FilePut returned %v", written)
	}
	remotePath := filepath.Join(serverRoot, "cf3/uploads/image.bin")
	got, err := os.ReadFile(remotePath)
	if err != nil {
		t.Fatalf("remote file not written: %v", err)
	}
	if !bytes.Equal(got, content) {
		t.Fatalf("remote content mismatch: got %d bytes, want %d", len(got), len(content))
	}
	if fi, _ := os.Stat(remotePath); fi.Mode().Perm() != 0644 {
		t.Errorf("remote file mode = %v, want 0644", fi.Mode().Perm())
	}
	// no leftover temp files next to the uploaded file
	entries, _ := os.ReadDir(filepath.Dir(remotePath))
	if len(entries) != 1 {
		names := make([]string, 0, len(entries))
		for _, e := range entries {
			names = append(names, e.Name())
		}
		t.Errorf("unexpected files in upload dir: %v", names)
	}

	// --- Stat: single file and directory listing
	a.Config.FileStatPath = []string{"cf3/uploads/image.bin"}
	stats, err := a.FileStat(ctx, tg)
	if err != nil {
		t.Fatalf("FileStat(file): %v", err)
	}
	if len(stats) != 1 || stats[0].IsDir || stats[0].StatInfo.GetSize() != uint64(len(content)) || stats[0].StatInfo.GetPermissions() != 644 {
		t.Errorf("FileStat(file) = %+v", stats)
	}
	a.Config.FileStatPath = []string{"cf3"}
	a.Config.FileStatRecursive = true
	stats, err = a.FileStat(ctx, tg)
	if err != nil {
		t.Fatalf("FileStat(dir): %v", err)
	}
	// cf3/uploads (dir) + cf3/uploads/image.bin (file)
	if len(stats) != 2 || !stats[0].IsDir || stats[0].StatInfo.GetPath() != "cf3/uploads" || stats[1].IsDir || stats[1].StatInfo.GetPath() != "cf3/uploads/image.bin" {
		t.Errorf("FileStat(dir, recursive) = %+v", stats)
	}

	// --- Get: server -> local, hash verified
	getDst := t.TempDir()
	a.Config.FileGetFile = []string{"cf3/uploads/image.bin"}
	a.Config.FileGetDst = getDst
	files, err := a.FileGet(ctx, tg)
	if err != nil {
		t.Fatalf("FileGet: %v", err)
	}
	if len(files) != 1 {
		t.Errorf("FileGet returned %v", files)
	}
	downloaded, err := os.ReadFile(filepath.Join(getDst, "cf3/uploads/image.bin"))
	if err != nil {
		t.Fatalf("downloaded file missing: %v", err)
	}
	if !bytes.Equal(downloaded, content) {
		t.Errorf("downloaded content mismatch")
	}

	// --- Get a directory: fetches every file underneath
	getDst2 := t.TempDir()
	a.Config.FileGetFile = []string{"cf3"}
	a.Config.FileGetDst = getDst2
	files, err = a.FileGet(ctx, tg)
	if err != nil {
		t.Fatalf("FileGet(dir): %v", err)
	}
	if len(files) != 1 || files[0] != "cf3/uploads/image.bin" {
		t.Errorf("FileGet(dir) returned %v", files)
	}
	if _, err := os.Stat(filepath.Join(getDst2, "cf3/uploads/image.bin")); err != nil {
		t.Errorf("FileGet(dir) did not write the file: %v", err)
	}

	// NOTE: --target-prefix is not exercised here: it currently joins the
	// prefix onto the full local path, producing a cwd-relative path (bug,
	// fixed in a later phase).

	// --- Remove
	fc := tg.FileClient()
	if _, err := fc.Remove(ctx, &file.RemoveRequest{RemoteFile: "cf3/uploads/image.bin"}); err != nil {
		t.Fatalf("Remove: %v", err)
	}
	if _, err := os.Stat(remotePath); !os.IsNotExist(err) {
		t.Errorf("file still exists after Remove: %v", err)
	}
	if _, err := fc.Remove(ctx, &file.RemoveRequest{RemoteFile: "cf3/uploads/image.bin"}); status.Code(err) != codes.FailedPrecondition {
		t.Errorf("Remove(missing) code = %v, want FailedPrecondition", status.Code(err))
	}
	if _, err := fc.Remove(ctx, &file.RemoveRequest{RemoteFile: "cf3"}); status.Code(err) != codes.InvalidArgument {
		t.Errorf("Remove(dir) code = %v, want InvalidArgument", status.Code(err))
	}
	if _, err := fc.Remove(ctx, &file.RemoveRequest{}); status.Code(err) != codes.InvalidArgument {
		t.Errorf("Remove(empty) code = %v, want InvalidArgument", status.Code(err))
	}
}

func TestFilePut_MultipleFilesAndPermissionsFromLocal(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	serverRoot := t.TempDir()
	tg := connectTarget(t, ctx, startFileServer(t, serverRoot, "md5"))

	localDir := t.TempDir()
	f1 := filepath.Join(localDir, "a.txt")
	f2 := filepath.Join(localDir, "sub", "b.txt")
	if err := os.MkdirAll(filepath.Dir(f2), 0755); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f1, []byte("aaa"), 0640); err != nil {
		t.Fatal(err)
	}
	if err := os.WriteFile(f2, []byte("bbb"), 0600); err != nil {
		t.Fatal(err)
	}

	a := newTestApp(t)
	a.Config.FilePutFile = []string{f1, f2}
	a.Config.FilePutDst = "dst"
	a.Config.FilePutChunkSize = defaultChunkSize
	a.Config.FilePutPermissions = 0 // use local permissions
	a.Config.FilePutHashMethod = "SHA256"

	written, err := a.FilePut(ctx, tg)
	if err != nil {
		t.Fatalf("FilePut: %v", err)
	}
	if len(written) != 2 {
		t.Errorf("FilePut returned %v", written)
	}
	// with several files the local path is appended to --dst
	for _, tt := range []struct {
		local string
		mode  os.FileMode
		body  string
	}{
		{f1, 0640, "aaa"},
		{f2, 0600, "bbb"},
	} {
		remote := filepath.Join(serverRoot, "dst", tt.local)
		b, err := os.ReadFile(remote)
		if err != nil {
			t.Errorf("%s not uploaded: %v", tt.local, err)
			continue
		}
		if string(b) != tt.body {
			t.Errorf("%s content = %q", tt.local, b)
		}
		if fi, _ := os.Stat(remote); fi.Mode().Perm() != tt.mode {
			t.Errorf("%s mode = %v, want %v", tt.local, fi.Mode().Perm(), tt.mode)
		}
	}
}

func TestFilePut_MissingLocalFile(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tg := connectTarget(t, ctx, startFileServer(t, t.TempDir(), "md5"))

	a := newTestApp(t)
	a.Config.FilePutFile = []string{filepath.Join(t.TempDir(), "nope.bin")}
	a.Config.FilePutDst = "x"
	a.Config.FilePutChunkSize = defaultChunkSize
	a.Config.FilePutHashMethod = "MD5"
	if _, err := a.FilePut(ctx, tg); err == nil {
		t.Fatal("expected error for missing local file")
	}
}

func TestFileServer_Put_RejectsBadHash(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	serverRoot := t.TempDir()
	tg := connectTarget(t, ctx, startFileServer(t, serverRoot, "md5"))

	stream, err := tg.FileClient().Put(ctx)
	if err != nil {
		t.Fatal(err)
	}
	open, _ := gfile.NewPutOpenRequest(gfile.FileName("bad.bin"), gfile.Permissions(644))
	if err := stream.Send(open); err != nil {
		t.Fatal(err)
	}
	contents, _ := gfile.NewPutContentRequest(gfile.Content([]byte("payload")))
	if err := stream.Send(contents); err != nil {
		t.Fatal(err)
	}
	wrong := md5.Sum([]byte("something else"))
	h, _ := gfile.NewPutHashRequest(gfile.HashMD5(wrong[:]))
	if err := stream.Send(h); err != nil {
		t.Fatal(err)
	}
	_, err = stream.CloseAndRecv()
	if status.Code(err) != codes.FailedPrecondition {
		t.Fatalf("CloseAndRecv error = %v, want FailedPrecondition", err)
	}
}

func TestFileServer_Put_RejectsBadFirstMessage(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	tg := connectTarget(t, ctx, startFileServer(t, t.TempDir(), "md5"))

	// first message must be Open
	stream, err := tg.FileClient().Put(ctx)
	if err != nil {
		t.Fatal(err)
	}
	contents, _ := gfile.NewPutContentRequest(gfile.Content([]byte("x")))
	if err := stream.Send(contents); err != nil {
		t.Fatal(err)
	}
	if _, err := stream.CloseAndRecv(); status.Code(err) != codes.InvalidArgument {
		t.Errorf("non-Open first message: code = %v, want InvalidArgument", status.Code(err))
	}

	// Open with an empty remote file name
	stream, err = tg.FileClient().Put(ctx)
	if err != nil {
		t.Fatal(err)
	}
	open, _ := gfile.NewPutOpenRequest(gfile.Permissions(644))
	if err := stream.Send(open); err != nil {
		t.Fatal(err)
	}
	if _, err := stream.CloseAndRecv(); status.Code(err) != codes.InvalidArgument {
		t.Errorf("empty remote_file: code = %v, want InvalidArgument", status.Code(err))
	}
}

func TestFileServer_Get_Errors(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	serverRoot := t.TempDir()
	if err := os.MkdirAll(filepath.Join(serverRoot, "dir"), 0755); err != nil {
		t.Fatal(err)
	}
	tg := connectTarget(t, ctx, startFileServer(t, serverRoot, "md5"))
	fc := tg.FileClient()

	recvErr := func(remote string) error {
		stream, err := fc.Get(ctx, &file.GetRequest{RemoteFile: remote})
		if err != nil {
			return err
		}
		_, err = stream.Recv()
		return err
	}
	if err := recvErr("missing.bin"); status.Code(err) != codes.FailedPrecondition {
		t.Errorf("Get(missing) code = %v, want FailedPrecondition", status.Code(err))
	}
	if err := recvErr("dir"); status.Code(err) != codes.InvalidArgument {
		t.Errorf("Get(dir) code = %v, want InvalidArgument", status.Code(err))
	}
	if _, err := fc.Stat(ctx, &file.StatRequest{Path: "missing"}); status.Code(err) != codes.FailedPrecondition {
		t.Errorf("Stat(missing) code = %v, want FailedPrecondition", status.Code(err))
	}
}

func TestFileServer_Get_EmptyFile(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
	defer cancel()
	serverRoot := t.TempDir()
	if err := os.WriteFile(filepath.Join(serverRoot, "empty"), nil, 0644); err != nil {
		t.Fatal(err)
	}
	tg := connectTarget(t, ctx, startFileServer(t, serverRoot, "md5"))

	stream, err := tg.FileClient().Get(ctx, &file.GetRequest{RemoteFile: "empty"})
	if err != nil {
		t.Fatal(err)
	}
	rsp, err := stream.Recv()
	if err != nil {
		t.Fatal(err)
	}
	emptyMD5 := md5.Sum(nil)
	if rsp.GetHash().GetMethod() != types.HashType_MD5 || !bytes.Equal(rsp.GetHash().GetHash(), emptyMD5[:]) {
		t.Errorf("empty file hash = %v", rsp.GetHash())
	}
	if _, err := stream.Recv(); err != io.EOF {
		t.Errorf("expected EOF after hash, got %v", err)
	}
}

func TestCompareFileHash(t *testing.T) {
	a := newTestApp(t)
	data := []byte("hello gnoi")
	sum := md5.Sum(data)
	if err := a.compareFileHash("t", bytes.NewBuffer(data), &types.HashType{Method: types.HashType_MD5, Hash: sum[:]}); err != nil {
		t.Errorf("matching md5: %v", err)
	}
	if err := a.compareFileHash("t", bytes.NewBuffer(data), &types.HashType{Method: types.HashType_MD5, Hash: []byte{1}}); err == nil {
		t.Error("expected error for wrong md5")
	}
	if err := a.compareFileHash("t", bytes.NewBuffer(data), &types.HashType{Method: types.HashType_UNSPECIFIED}); err == nil {
		t.Error("expected error for unspecified hash type")
	}
	if err := a.compareFileHash("t", bytes.NewBuffer(data), &types.HashType{Method: types.HashType_HashMethod(99)}); err == nil {
		t.Error("expected error for unknown hash type")
	}
}
