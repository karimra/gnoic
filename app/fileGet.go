package app

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"fmt"
	"hash"
	"os"
	"path/filepath"
	"strings"

	"github.com/openconfig/gnoi/file"
	"github.com/openconfig/gnoi/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/prototext"
)

type fileGetResponse struct {
	TargetError
	file string
}

func (a *App) InitFileGetFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.FileGetFile, "file", "", "file to get from the target(s)")
	cmd.Flags().StringVar(&a.Config.FileGetLocalFile, "local-file", "", "local file name, defaults to the path base of the retrieved file")
	cmd.Flags().BoolVar(&a.Config.FileGetTargetPrefix, "target-prefix", false, "save file with the target name as prefix")
	//
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunEFileGet(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *fileGetResponse, numTargets)

	a.wg.Add(numTargets)
	for _, t := range targets {
		go func(t *Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = a.CreateGrpcClient(ctx, t, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &fileGetResponse{
					TargetError: TargetError{
						TargetName: t.Config.Address,
						Err:        err,
					},
				}
				return
			}
			filename, err := a.FileGet(ctx, t)
			responseChan <- &fileGetResponse{
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
	result := make([]*fileGetResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q File Get failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
		result = append(result, rsp)
	}

	for _, r := range result {
		a.Logger.Infof("%q file %q saved", r.TargetName, r.file)
	}
	return a.handleErrs(errs)
}

func (a *App) FileGet(ctx context.Context, t *Target) (string, error) {
	fileClient := file.NewFileClient(t.client)
	stream, err := fileClient.Get(ctx, &file.GetRequest{
		RemoteFile: a.Config.FileGetFile,
	})
	if err != nil {
		return "", err
	}
	b := new(bytes.Buffer)
	for {
		b.Grow(64000)
		getResponse, err := stream.Recv()
		if err != nil {
			return "", err
		}

		a.Logger.Debugf(prototext.Format(getResponse))

		content := getResponse.GetContents()
		if content != nil {
			a.Logger.Infof("%q received %d bytes", t.Config.Address, len(content))
			b.Write(content)
			continue
		}
		h := getResponse.GetHash()
		if h == nil {
			a.Logger.Infof("%q received nil hash", t.Config.Address)
			return "", nil
		}
		a.Logger.Debugf("%q received hash method %s", t.Config.Address, h.Method)
		err = a.compareFileHash(t.Config.Address, b, h)
		if err != nil {
			return "", fmt.Errorf("%q hash err: %v", t.Config.Address, err)
		}
		break
	}
	name := a.Config.FileGetLocalFile
	if name == "" {
		name = filepath.Base(a.Config.FileGetFile)
	}
	if a.Config.FileGetTargetPrefix {
		name = strings.Join([]string{t.Config.Address, name}, "_")
	}
	f, err := os.Create(name)
	if err != nil {
		return "", err
	}
	defer f.Close()
	f.Write(b.Bytes())
	a.Logger.Debugf("%q wrote local file %q", t.Config.Address, name)
	return name, nil
}

func (a *App) compareFileHash(tName string, b *bytes.Buffer, ht *types.HashType) error {
	var r int
	var cHash []byte
	var h hash.Hash
	switch ht.Method {
	case types.HashType_MD5:
		h = md5.New()
	case types.HashType_SHA256:
		h = sha256.New()
	case types.HashType_SHA512:
		h = sha512.New()
	case types.HashType_UNSPECIFIED:
		return fmt.Errorf("%q unspecified Hash Type", tName)
	default:
		return fmt.Errorf("%q unknown Hash Type %q", tName, ht.Method)
	}

	h.Write(b.Bytes())
	cHash = h.Sum(nil)
	r = bytes.Compare(cHash, ht.Hash)
	if r != 0 {
		a.Logger.Errorf("%q wrong Hash_%s: received: %x", tName, ht.Method.String(), ht.Hash)
		a.Logger.Errorf("%q wrong Hash_%s: calculated: %x", tName, ht.Method.String(), cHash)
		return fmt.Errorf("%q wrong Hash_%s: recv: %x, calc: %x", tName, ht.Method.String(), ht.Hash, cHash)
	}
	a.Logger.Debugf("%q Hash_%s recv: %x", tName, ht.Method.String(), ht.Hash)
	a.Logger.Debugf("%q Hash_%s calc: %x", tName, ht.Method.String(), cHash)
	return nil
}
