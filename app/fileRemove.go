package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/openconfig/gnoi/file"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
)

type fileRemoveResponse struct {
	TargetError
	file []string
}

func (a *App) InitFileRemoveFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringSliceVar(&a.Config.FileRemovePath, "path", []string{}, "remote path pointing to file(s)/dir(s) to remove from the target(s)")
	//
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunEFileRemove(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *fileRemoveResponse, numTargets)

	a.wg.Add(numTargets)
	for _, t := range targets {
		go func(t *Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = a.CreateGrpcClient(ctx, t, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &fileRemoveResponse{
					TargetError: TargetError{
						TargetName: t.Config.Address,
						Err:        err,
					},
				}
				return
			}
			filename, err := a.FileRemove(ctx, t)
			responseChan <- &fileRemoveResponse{
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
	result := make([]*fileRemoveResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q File Remove failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
		result = append(result, rsp)
	}

	for _, r := range result {
		for _, f := range r.file {
			a.Logger.Infof("%q file %q removed successfully", r.TargetName, f)
		}
	}
	return a.handleErrs(errs)
}

func (a *App) FileRemove(ctx context.Context, t *Target) ([]string, error) {
	fileClient := file.NewFileClient(t.client)
	errs := make([]string, 0, len(a.Config.FileRemovePath))
	for _, file := range a.Config.FileRemovePath {
		err := a.fileRemove(ctx, t, fileClient, file)
		if err != nil {
			errs = append(errs, fmt.Sprintf("%v", err))
		}
	}
	var err error
	if len(errs) > 0 {
		err = fmt.Errorf("%v", strings.Join(errs, ",\n"))
	}
	return a.Config.FileRemovePath, err
}

func (a *App) fileRemove(ctx context.Context, t *Target, fileClient file.FileClient, path string) error {
	isDir, err := a.isDir(ctx, fileClient, path)
	if err != nil {
		return err
	}
	if isDir {
		a.Logger.Debugf("%q remote=%q is a directory", t.Config.Address, path)
		r, err := fileClient.Stat(ctx, &file.StatRequest{
			Path: path,
		})
		if err != nil {
			return err
		}
		if len(r.Stats) == 0 {
			return fmt.Errorf("%q path %q is an empty directory", t.Config.Address, path)
		}
		for _, s := range r.Stats {
			a.Logger.Debugf("%q removing file %q", t.Config.Address, s.Path)
			err = a.fileRemove(ctx, t, fileClient, s.Path)
			if err != nil {
				return err
			}
		}
		return nil
	}
	a.Logger.Debugf("%q remote=%q is a file", t.Config.Address, path)
	_, err = fileClient.Remove(ctx, &file.RemoveRequest{
		RemoteFile: path,
	})
	return err
}
