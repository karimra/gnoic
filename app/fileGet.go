package app

import (
	"bytes"
	"context"
	"crypto/md5"
	"crypto/sha256"
	"crypto/sha512"
	"errors"
	"fmt"
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
	targetName string
	file       string
	err        error
}

func (a *App) InitFileGetFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.FileGetFile, "file", "", "file to get from the target(s)")
	cmd.Flags().StringVar(&a.Config.FileGetLocalFile, "local-file", "", "local file name")
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
					targetName: t.Config.Address,
					err:        err,
				}
				return
			}
			filename, err := a.FileGet(ctx, t)
			responseChan <- &fileGetResponse{
				targetName: t.Config.Address,
				file:       filename,
				err:        err,
			}
		}(t)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	result := make([]*fileGetResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.err != nil {
			a.Logger.Errorf("%q file Get failed: %v", rsp.targetName, rsp.err)
			errs = append(errs, rsp.err)
			continue
		}
		result = append(result, rsp)
	}

	for _, err := range errs {
		a.Logger.Errorf("err: %v", err)
	}
	for _, r := range result {
		fmt.Printf("%q file %q saved\n", r.targetName, r.file)
	}
	//
	if len(errs) > 0 {
		return fmt.Errorf("there was %d errors", len(errs))
	}
	a.Logger.Debug("done...")
	return nil
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
		a.Logger.Infof("%q received hash method %s", t.Config.Address, h.Method)
		err = a.compareFileHash(b, h)
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

func (a *App) compareFileHash(b *bytes.Buffer, h *types.HashType) error {
	switch h.Method {
	case types.HashType_MD5:
		m := md5.New()
		m.Write(b.Bytes())
		cHash := m.Sum(nil)
		r := bytes.Compare(cHash, h.Hash)
		if r != 0 {
			a.Logger.Errorf("wrong Hash: received: %x, calculated: %x", h.Hash, cHash)
			return errors.New("wrong hash")
		}
		a.Logger.Debugf("Hash: received: %x, calculated: %x", h.Hash, cHash)
	case types.HashType_SHA256:
		s := sha256.New()
		s.Write(b.Bytes())
		cHash := s.Sum(nil)
		r := bytes.Compare(cHash, h.Hash)
		if r != 0 {
			a.Logger.Errorf("wrong Hash: received: %x, calculated: %x", h.Hash, cHash)
			return errors.New("wrong hash")
		}
		a.Logger.Debugf("Hash: received: %x, calculated: %x", h.Hash, cHash)
	case types.HashType_SHA512:
		s := sha512.New()
		s.Write(b.Bytes())
		cHash := s.Sum(nil)
		r := bytes.Compare(cHash, h.Hash)
		if r != 0 {
			a.Logger.Errorf("wrong Hash: received: %v, calculated: %v", h.Hash, cHash)
			return errors.New("wrong hash")
		}
		a.Logger.Debugf("Hash: received: %x, calculated: %x", h.Hash, cHash)
	case types.HashType_UNSPECIFIED:
		return errors.New("unspecified Hash Type")
	default:
		return errors.New("unknown Hash Type")
	}
	return nil
}
