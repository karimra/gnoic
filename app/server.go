package app

import (
	"bufio"
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"io"
	"io/fs"
	"io/ioutil"
	"net"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"time"

	scp "github.com/bramvdbogaerde/go-scp"
	gfile "github.com/karimra/gnoic/api/file"
	"github.com/mitchellh/go-homedir"
	"github.com/openconfig/gnoi/common"
	"github.com/openconfig/gnoi/file"
	"github.com/openconfig/gnoi/types"
	"github.com/pkg/sftp"
	"github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"golang.org/x/crypto/ssh"
	"golang.org/x/sys/unix"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (a *App) InitServerFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().BoolVar(&a.Config.ServerFile, "file", false, "start gNOI File service server")
	cmd.Flags().StringVar(&a.Config.ServerFileHash, "file-hash", "md5", "hash type to use at the end of File Get/Transfer RPC. md5, sha256, sha512")
	//
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunEServer(cmd *cobra.Command, args []string) error {
	var l net.Listener
	var err error
	network := "tcp"
	for {
		l, err = net.Listen(network, a.Config.Address[0])
		if err != nil {
			a.Logger.Printf("failed to start gRPC server listener: %v", err)
			time.Sleep(time.Second)
			continue
		}
		break
	}

	homedir, _ := homedir.Dir()
	fileServer := &fserver{
		logger:         a.Logger.WithField("server", "file"),
		s:              grpc.NewServer(),
		rootDir:        homedir,
		fileHashMethod: strings.ToLower(a.Config.ServerFileHash),
	}
	file.RegisterFileServer(fileServer.s, fileServer)
	ctx, cancel := context.WithCancel(a.ctx)
	go func() {
		err = fileServer.s.Serve(l)
		if err != nil {
			a.Logger.Printf("gRPC server shutdown: %v", err)
		}
		cancel()
	}()
	fileServer.logger.Info("file Server started...")
	<-ctx.Done()
	return nil
}

type fserver struct {
	file.UnimplementedFileServer

	logger         *logrus.Entry
	s              *grpc.Server
	rootDir        string
	fileHashMethod string
}

func (s *fserver) Get(req *file.GetRequest, stream file.File_GetServer) error {
	s.logger.Infof("received get request: %+v", req)
	localFile := filepath.Join(s.rootDir, req.GetRemoteFile())
	fi, err := os.Stat(localFile)
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "%v", err)
	}
	if fi.IsDir() {
		return status.Error(codes.InvalidArgument, "requested path is a directory")
	}
	getFile, err := os.Open(localFile)
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "%v", err)
	}
	defer getFile.Close()
	r := bufio.NewReader(getFile)
	buf := make([]byte, 0, 64*1000)

	var h hash.Hash
	switch s.fileHashMethod {
	default:
		h = md5.New()
	case "sha256":
		h = md5.New()
	case "sha512":
		h = md5.New()
	}
OUTER:
	for {
		select {
		case <-stream.Context().Done():
			return status.Errorf(codes.FailedPrecondition, "%v", stream.Context().Err())
		default:
			n, err := r.Read(buf[:cap(buf)])
			if err != nil {
				if err == io.EOF {
					break OUTER
				}
				return status.Errorf(codes.FailedPrecondition, "%v", err)
			}
			buf = buf[:n]
			h.Write(buf)
			err = stream.Send(&file.GetResponse{
				Response: &file.GetResponse_Contents{
					Contents: buf,
				},
			})
			if err != nil {
				return err
			}
		}
	}
	cHash := h.Sum(nil)
	hashRsp, err := gfile.NewGetHashResponse(gfile.HashMD5(cHash))
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "%v", err)
	}
	err = stream.Send(hashRsp)
	if err != nil {
		return status.Errorf(codes.FailedPrecondition, "%v", err)
	}
	return nil
}

func (s *fserver) TransferToRemote(ctx context.Context, req *file.TransferToRemoteRequest) (*file.TransferToRemoteResponse, error) {
	s.logger.Infof("received transfer request: %+v", req)
	localFile := filepath.Join(s.rootDir, req.GetLocalPath())
	fi, err := os.Stat(localFile)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
	}
	if fi.IsDir() {
		return nil, status.Error(codes.InvalidArgument, "cannot transfer directory")
	}
	f, err := os.Open(localFile)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
	}
	defer f.Close()

	var addr string
	var remotePath string
	switch req.GetRemoteDownload().GetProtocol() {
	case common.RemoteDownload_HTTP:
		return nil, status.Errorf(
			codes.Unimplemented,
			"RemoteDownload Protocol %v not implemented",
			req.GetRemoteDownload().GetProtocol())
	case common.RemoteDownload_HTTPS:
		return nil, status.Errorf(
			codes.Unimplemented,
			"RemoteDownload Protocol %v not implemented",
			req.GetRemoteDownload().GetProtocol())
	case common.RemoteDownload_SCP:
		config := &ssh.ClientConfig{
			User: req.GetRemoteDownload().GetCredentials().GetUsername(),
			Auth: []ssh.AuthMethod{
				ssh.Password(req.GetRemoteDownload().GetCredentials().GetCleartext()),
			},
			Timeout:         30 * time.Second,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		if sep := strings.Index(req.GetRemoteDownload().GetPath(), ":"); sep > 0 {
			addr = req.GetRemoteDownload().GetPath()[:sep]
			remotePath = req.GetRemoteDownload().GetPath()[sep+1:]
			s.logger.Debugf("remote sftp address: %s, remote file: %s", addr, remotePath)
		} else {
			return nil, fmt.Errorf("failed to parse remoteDownloadPath")
		}
		// create ssh conn
		conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", addr), config)
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		// create scp client
		client, err := scp.NewClientBySSH(conn)
		if err != nil {
			return nil, err
		}
		err = client.CopyFile(ctx, f, remotePath, "0"+strconv.Itoa(int(decimalToOctal(uint32(fi.Mode().Perm())))))
		if err != nil {
			return nil, err
		}
	case common.RemoteDownload_SFTP:
		config := &ssh.ClientConfig{
			User: req.GetRemoteDownload().GetCredentials().GetUsername(),
			Auth: []ssh.AuthMethod{
				ssh.Password(req.GetRemoteDownload().GetCredentials().GetCleartext()),
			},
			Timeout:         30 * time.Second,
			HostKeyCallback: ssh.InsecureIgnoreHostKey(),
		}
		if sep := strings.Index(req.GetRemoteDownload().GetPath(), ":"); sep > 0 {
			addr = req.GetRemoteDownload().GetPath()[:sep]
			remotePath = req.GetRemoteDownload().GetPath()[sep+1:]
			s.logger.Debugf("remote sftp address: %s, remote file: %s", addr, remotePath)
		} else {
			return nil, fmt.Errorf("failed to parse remoteDownloadPath")
		}
		// create ssh conn
		conn, err := ssh.Dial("tcp", fmt.Sprintf("%s:22", addr), config)
		if err != nil {
			return nil, err
		}
		defer conn.Close()
		// create sftp client
		sc, err := sftp.NewClient(conn)
		if err != nil {
			return nil, err
		}
		defer sc.Close()
		// open remote file
		dstFile, err := sc.OpenFile(remotePath, os.O_WRONLY|os.O_CREATE|os.O_TRUNC)
		if err != nil {
			return nil, err
		}
		defer dstFile.Close()
		// copy local file to destination
		_, err = io.Copy(dstFile, f)
		if err != nil {
			return nil, err
		}
	default:
		return nil, status.Errorf(codes.InvalidArgument, "unknown RemoteDownload Protocol: %v", req.GetRemoteDownload().GetProtocol())
	}
	// calculate file hash
	var h hash.Hash
	switch s.fileHashMethod {
	case "md5":
		h = md5.New()
	case "sha256":
		h = sha256.New()
	case "sha512":
		h = sha512.New()
	}
	// rewind the file
	f.Seek(0, io.SeekStart)
	r := bufio.NewReader(f)

	buf := make([]byte, 0, 64000*10)
OUTER:
	for {
		n, err := r.Read(buf[:cap(buf)])
		if err != nil {
			if err == io.EOF {
				break OUTER
			}
			return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
		}
		buf = buf[:n]
		h.Write(buf)
		if err != nil {
			return nil, err
		}
	}
	cHash := h.Sum(nil)
	hashRsp, err := gfile.NewTransferResponse(gfile.HashMD5(cHash))
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
	}
	return hashRsp, nil
}

func (s *fserver) Put(stream file.File_PutServer) error {
	s.logger.Infof("received put request")
	var tempFile *os.File
	var rFile *os.File
	var rFileMode os.FileMode

	// get the first stream request, must be an Open Request
	req, err := stream.Recv()
	if err != nil {
		return err
	}

	switch req := req.GetRequest().(type) {
	case *file.PutRequest_Open:
		s.logger.Infof("received put request Open: %v", req)
		remoteFile := req.Open.GetRemoteFile()
		if remoteFile == "" {
			return status.Errorf(codes.InvalidArgument, "remote_file cannot be empty")
		}
		remoteFile = filepath.Join(s.rootDir, remoteFile)
		dir := filepath.Dir(remoteFile)
		rFileMode = fs.FileMode(octalToDecimal(req.Open.GetPermissions()))
		err = os.MkdirAll(dir, 0744) // TODO:
		if err != nil {
			return status.Errorf(codes.FailedPrecondition, "failed to create dir: %v", err)
		}
		// create file
		rFile, err = os.OpenFile(
			remoteFile,
			os.O_RDWR|os.O_CREATE|os.O_TRUNC,
			rFileMode,
		)
		if err != nil {
			return status.Errorf(codes.FailedPrecondition, "failed to create file: %v", err)
		}
		defer rFile.Close()
		// create temp file
		tempFile, err = ioutil.TempFile(dir, filepath.Base(remoteFile))
		if err != nil {
			return status.Errorf(codes.FailedPrecondition, "%v", err)
		}
		defer tempFile.Close()
	default:
		return status.Errorf(codes.InvalidArgument, "initial message must be PutRequest_Open: received %T", req)
	}

	for {
		req, err := stream.Recv()
		if err != nil {
			return status.Errorf(codes.FailedPrecondition, "%v", err)
		}
		switch req := req.GetRequest().(type) {
		case *file.PutRequest_Open:
			return status.Error(codes.InvalidArgument, "unexpected PutRequest_Open message")
		case *file.PutRequest_Contents:
			_, err = tempFile.Write(req.Contents)
			if err != nil {
				return status.Errorf(codes.FailedPrecondition, "%v", err)
			}
		case *file.PutRequest_Hash:
			var h hash.Hash
			switch req.Hash.GetMethod() {
			case types.HashType_MD5:
				h = md5.New()
			case types.HashType_SHA256:
				h = sha256.New()
			case types.HashType_SHA512:
				h = sha512.New()
			default:
				return status.Errorf(codes.InvalidArgument, "unexpected HashType: %v", req.Hash.GetMethod())
			}
			// close temp file
			tempFileName := tempFile.Name()
			tempFile.Close()
			tf, err := os.Open(tempFileName)
			if err != nil {
				return status.Errorf(codes.FailedPrecondition, "open temp file err: %v", err)
			}

			// read temp file, calculate hash
			r := bufio.NewReader(tf)
			buf := make([]byte, 0, 1024*1024*1024)
			for {
				n, err := r.Read(buf[:cap(buf)])
				if err != nil {
					if err == io.EOF {
						break
					}
					return status.Errorf(codes.FailedPrecondition, "%v", err)
				}
				buf = buf[:n]
				h.Write(buf)
			}

			cHash := h.Sum(nil)
			if !bytes.Equal(cHash, req.Hash.GetHash()) {
				return status.Errorf(codes.FailedPrecondition, "wrong hash: expected %x, received: %x", cHash, req.Hash.GetHash())
			}
			// hash ok
			// rename file
			err = os.Rename(tempFileName, rFile.Name())
			if err != nil {
				return status.Errorf(codes.FailedPrecondition, "failed to rename temp file: %v", err)
			}
			// change file perms
			err = os.Chmod(rFile.Name(), rFileMode)
			if err != nil {
				return status.Errorf(codes.FailedPrecondition, "failed chmod: %v", err)
			}
			// send put rsp
			return stream.SendAndClose(new(file.PutResponse))
		default:
			return status.Error(codes.InvalidArgument, "unexpected message type")
		}
	}
}

func (s *fserver) Stat(ctx context.Context, req *file.StatRequest) (*file.StatResponse, error) {
	s.logger.Infof("received file stat request: %+v", req)
	statPath := filepath.Join(s.rootDir, req.Path)
	fi, err := os.Stat(statPath)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
	}
	oldUmask := unix.Umask(0)
	unix.Umask(oldUmask)
	opts := []gfile.FileOption{}
	if fi.IsDir() {
		files, err := os.ReadDir(statPath)
		if err != nil {
			return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
		}
		for _, fs := range files {
			ffi, err := fs.Info()
			if err != nil {
				return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
			}
			opts = append(opts, gfile.StatInfo(
				gfile.Path(filepath.Join(req.Path, fs.Name())),
				gfile.LastModified(uint64(ffi.ModTime().UnixNano())),
				gfile.Permissions(decimalToOctal(uint32(ffi.Mode().Perm()))),
				gfile.Size(uint64(ffi.Size())),
				gfile.Umask(decimalToOctal(uint32(oldUmask))),
			))
		}
	} else {
		opts = append(opts,
			gfile.StatInfo(
				gfile.Path(req.Path),
				gfile.LastModified(uint64(fi.ModTime().UnixNano())),
				gfile.Permissions(decimalToOctal(uint32(fi.Mode().Perm()))),
				gfile.Size(uint64(fi.Size())),
				gfile.Umask(decimalToOctal(uint32(oldUmask))),
			),
		)
	}
	rsp, err := gfile.NewStatResponse(opts...)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
	}
	return rsp, nil
}

func (s *fserver) Remove(ctx context.Context, req *file.RemoveRequest) (*file.RemoveResponse, error) {
	s.logger.Infof("received file remove request: %+v", req)
	if req.GetRemoteFile() == "" {
		return nil, status.Error(codes.InvalidArgument, "remote_file cannot be empty")
	}
	statPath := filepath.Join(s.rootDir, req.GetRemoteFile())
	fi, err := os.Stat(statPath)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
	}
	if fi.IsDir() {
		return nil, status.Error(codes.InvalidArgument, "remote_file cannot be a directory")
	}
	err = os.Remove(statPath)
	if err != nil {
		return nil, status.Errorf(codes.FailedPrecondition, "%v", err)
	}
	return new(file.RemoveResponse), nil
}

func decimalToOctal(d uint32) uint32 {
	remainders := make([]uint32, 0)
	var v = d
	for v != 0 {
		r := v % 8
		v = v / 8
		remainders = append(remainders, r)
	}
	var o uint32
	for i := len(remainders) - 1; i >= 0; i-- {
		o = o*10 + remainders[i]
	}
	return o
}

func octalToDecimal(d uint32) uint32 {
	remainders := make([]uint32, 0)
	var v = d
	for v != 0 {
		r := v % 10
		v = v / 10
		remainders = append(remainders, r)
	}
	var o uint32
	for i := len(remainders) - 1; i >= 0; i-- {
		o = o*8 + remainders[i]
	}
	return o
}
