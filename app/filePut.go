package app

import (
	"context"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
	"hash"
	"io"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"sync"

	"github.com/openconfig/gnoi/file"
	"github.com/openconfig/gnoi/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/prototext"
)

const (
	defaultChunkSize = 64 * 1024
)

type filePutResponse struct {
	TargetError
	file []string
}

func (a *App) InitFilePutFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringSliceVar(&a.Config.FilePutFile, "file", []string{}, "file(s) to put on the target(s)")
	cmd.Flags().StringVar(&a.Config.FilePutDst, "dst", "", "destination file/directory name")
	cmd.Flags().Uint64Var(&a.Config.FilePutChunkSize, "chunk-size", defaultChunkSize, "chunk write size in Bytes, default is used if set to 0")
	cmd.Flags().Uint32Var(&a.Config.FilePutPermissions, "permission", 0777, "file permissions, in octal format. If set to 0, the local system file permissions are used")
	cmd.Flags().StringVar(&a.Config.FilePutHashMethod, "hash-method", "MD5", "hash method, one of MD5, SHA256 or SHA512. If another value is supplied MD5 is used")
	//
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) PreRunEFilePut(cmd *cobra.Command, args []string) error {
	a.Config.SetLocalFlagsFromFile(cmd)
	if len(a.Config.FilePutFile) == 0 {
		return errors.New("missing --file flag")
	}
	if a.Config.FilePutChunkSize == 0 {
		a.Config.FilePutChunkSize = defaultChunkSize
	}
	a.Config.FilePutHashMethod = strings.ToUpper(a.Config.FilePutHashMethod)
	switch a.Config.FilePutHashMethod {
	case "MD5", "SHA256", "SHA512":
	default:
		a.Config.FilePutHashMethod = "MD5"
	}
	foundFiles := []string{}
	for _, path := range a.Config.FilePutFile {
		err := filepath.Walk(path, func(path string, fi os.FileInfo, err error) error {
			if err != nil {
				return err
			}
			if !fi.IsDir() {
				foundFiles = append(foundFiles, path)
			}
			return nil
		})
		if err != nil {
			a.Logger.Errorf("failed walking directory %q: %v", path, err)
			return err
		}
	}
	a.Config.FilePutFile = foundFiles
	return nil
}

func (a *App) RunEFilePut(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *filePutResponse, numTargets)

	a.wg.Add(numTargets)
	for _, t := range targets {
		go func(t *Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = a.CreateGrpcClient(ctx, t, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &filePutResponse{
					TargetError: TargetError{
						TargetName: t.Config.Address,
						Err:        err,
					},
				}
				return
			}
			filename, err := a.FilePut(ctx, t)
			responseChan <- &filePutResponse{
				TargetError: TargetError{
					TargetName: t.Config.Address,
					Err:        err,
				},
				file: filename,
			}
		}(t)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	result := make([]*filePutResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q File Put failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
		result = append(result, rsp)
	}

	for _, r := range result {
		for _, f := range r.file {
			a.Logger.Infof("%q file %q written successfully", r.TargetName, f)
		}
	}
	return a.handleErrs(errs)
}

func (a *App) FilePut(ctx context.Context, t *Target) ([]string, error) {
	numFiles := len(a.Config.FilePutFile)

	errChan := make(chan error, numFiles)
	fileChan := make(chan string, numFiles)

	fileClient := file.NewFileClient(t.client)

	wg := new(sync.WaitGroup)
	wg.Add(numFiles)
	for _, filename := range a.Config.FilePutFile {
		go func(filename string) {
			defer wg.Done()
			fi, err := os.Stat(filename)
			if err != nil {
				errChan <- fmt.Errorf("file %q stat err: %v", filename, err)
				return
			}
			fPerm := a.Config.FilePutPermissions
			if fPerm == 0 {
				perm := "0" + strconv.FormatUint(uint64(fi.Mode().Perm()), 8)
				a.Logger.Infof("setting permission to %s", perm)
				operm, err := strconv.ParseInt(perm, 8, 64)
				if err != nil {
					errChan <- fmt.Errorf("file %q perm read err: %v", filename, err)
					return
				}
				fPerm = uint32(operm)
			}
			var remoteName = a.Config.FilePutDst
			if numFiles > 1 {
				remoteName = filepath.Join(remoteName, filename)
			}

			err = a.filePut(ctx, t, fileClient, filename, remoteName, fPerm)
			if err != nil {
				errChan <- err
				return
			}
			fileChan <- filename
		}(filename)
	}
	wg.Wait()
	close(errChan)
	close(fileChan)

	errs := make([]string, 0, numFiles)
	files := make([]string, 0, numFiles)

	for f := range fileChan {
		files = append(files, f)
	}
	for e := range errChan {
		errs = append(errs, fmt.Sprintf("%v", e))
	}
	var err error
	if len(errs) > 0 {
		err = fmt.Errorf("%s", strings.Join(errs, ",\n"))
	}
	return files, err
}

func (a *App) filePut(ctx context.Context, t *Target, fileClient file.FileClient, localFile, remote string, perm uint32) error {
	// open local file
	f, err := os.Open(localFile)
	if err != nil {
		a.Logger.Errorf("failed opening file %q: %v", localFile, err)
		return err
	}
	// start stream
	stream, err := fileClient.Put(ctx)
	if err != nil {
		return err
	}
	defer stream.CloseSend()
	//
	req := &file.PutRequest{
		Request: &file.PutRequest_Open{
			Open: &file.PutRequest_Details{
				RemoteFile:  remote,
				Permissions: perm,
			},
		},
	}
	a.Logger.Debug(prototext.Format(req))
	err = stream.Send(req)
	if err != nil {
		return err
	}

	// init hash.Hash
	var h hash.Hash
	switch a.Config.FilePutHashMethod {
	case "MD5":
		h = md5.New()
	case "SHA256":
		h = sha256.New()
	case "SHA512":
		h = sha512.New()
	}

	// send file in chunks
	for {
		b := make([]byte, a.Config.FilePutChunkSize)
		n, err := f.Read(b)
		if err != nil && err != io.EOF {
			return err
		}
		if err == io.EOF || n == 0 {
			break
		}
		h.Write(b[:n])
		a.Logger.Debugf("%q file=%q, remote=%q writing %d byte(s)", t.Config.Address, localFile, remote, n)
		reqContents := &file.PutRequest{
			Request: &file.PutRequest_Contents{
				Contents: b[:n],
			},
		}
		a.Logger.Debug(reqContents)
		err = stream.Send(reqContents)
		if err != nil {
			return err
		}
	}
	// send Hash
	a.Logger.Infof("%q sending file=%q hash", t.Config.Address, localFile)
	reqHash := &file.PutRequest{
		Request: &file.PutRequest_Hash{
			Hash: &types.HashType{
				Method: types.HashType_HashMethod(types.HashType_HashMethod_value[a.Config.FilePutHashMethod]),
				Hash:   h.Sum(nil),
			},
		},
	}
	a.Logger.Debug(reqHash)
	err = stream.Send(reqHash)
	if err != nil {
		return err
	}
	_, err = stream.CloseAndRecv()
	return err
}
