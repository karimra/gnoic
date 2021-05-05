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

	"github.com/openconfig/gnoi/file"
	"github.com/openconfig/gnoi/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
)

const (
	defaultChunkSize = 64 * 1000
)

type filePutResponse struct {
	TargetError
	file string
}

func (a *App) InitFilePutFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.FilePutFile, "file", "", "file to put on the target(s)")
	cmd.Flags().StringVar(&a.Config.FilePutRemoteFile, "remote-name", "", "file remote name, defaults to the path Base of the local file")
	cmd.Flags().Uint64Var(&a.Config.FilePutWriteSize, "chunk-size", defaultChunkSize, "chunk write size in Bytes, default is used if set to 0")
	cmd.Flags().Uint32Var(&a.Config.FilePutPermissions, "permission", 0777, "file permissions, in octal format. If set to 0, the local system file permissions are used")
	cmd.Flags().StringVar(&a.Config.FilePutHashMethod, "hash-method", "MD5", "hash method, one of MD5, SHA256 or SHA512. If another value is supplied MD5 is used")
	//
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) PreRunEFilePut(cmd *cobra.Command, args []string) error {
	if a.Config.FilePutFile == "" {
		return errors.New("missing --file flag")
	}
	if a.Config.FilePutWriteSize == 0 {
		a.Config.FilePutWriteSize = defaultChunkSize
	}
	a.Config.FilePutHashMethod = strings.ToUpper(a.Config.FilePutHashMethod)
	switch a.Config.FilePutHashMethod {
	case "MD5", "SHA256", "SHA512":
	default:
		a.Config.FilePutHashMethod = "MD5"
	}
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
		a.Logger.Infof("%q file %q written successfully", r.TargetName, r.file)
	}
	return a.handleErrs(errs)
}

func (a *App) FilePut(ctx context.Context, t *Target) (string, error) {
	fi, err := os.Stat(a.Config.FilePutFile)
	if err != nil {
		return "", err
	}
	if fi.IsDir() {
		// TODO:
		return "", fmt.Errorf("%q file put direcotries is not supported,... yet", t.Config.Address)
	}
	// open local file
	f, err := os.Open(a.Config.FilePutFile)
	if err != nil {
		return "", err
	}

	fileClient := file.NewFileClient(t.client)
	stream, err := fileClient.Put(ctx)
	if err != nil {
		return "", err
	}
	defer stream.CloseSend()
	if a.Config.FilePutRemoteFile == "" {
		a.Config.FilePutRemoteFile = filepath.Base(a.Config.FilePutFile)
	}
	if a.Config.FilePutPermissions == 0 {
		perm := "0" + strconv.FormatUint(uint64(fi.Mode().Perm()), 8)
		a.Logger.Infof("setting permission to %s", perm)
		operm, err := strconv.ParseInt(perm, 8, 64)
		if err != nil {
			return "", err
		}
		a.Config.FilePutPermissions = uint32(operm)
	}
	req := &file.PutRequest{
		Request: &file.PutRequest_Open{
			Open: &file.PutRequest_Details{
				RemoteFile:  a.Config.FilePutRemoteFile,
				Permissions: a.Config.FilePutPermissions,
			},
		},
	}
	a.Logger.Debug(req)
	err = stream.Send(req)
	if err != nil {
		return "", err
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
		b := make([]byte, a.Config.FilePutWriteSize)
		n, err := f.Read(b)
		if err != nil && err != io.EOF {
			return "", err
		}
		if err == io.EOF || n == 0 {
			break
		}
		h.Write(b[:n])
		a.Logger.Infof("writing %d byte(s) to %q", n, t.Config.Address)
		reqContents := &file.PutRequest{
			Request: &file.PutRequest_Contents{
				Contents: b[:n],
			},
		}
		a.Logger.Debug(reqContents)
		err = stream.Send(reqContents)
		if err != nil {
			return "", err
		}
	}
	// send Hash
	a.Logger.Infof("sending file hash to %q", t.Config.Address)
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
		return "", err
	}
	_, err = stream.CloseAndRecv()
	return a.Config.FilePutRemoteFile, err
}
