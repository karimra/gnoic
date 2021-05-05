package app

import (
	"context"
	"fmt"

	"github.com/openconfig/gnoi/file"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
)

type fileRemoveResponse struct {
	TargetError
	file string
}

func (a *App) InitFileRemoveFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.FileRemoveFile, "file", "", "file to remove from the target(s)")
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
			a.Logger.Errorf("%q file Remove failed: %v", rsp.TargetName, rsp.Err)
			errs = append(errs, rsp.Err)
			continue
		}
		result = append(result, rsp)
	}

	for _, err := range errs {
		a.Logger.Errorf("err: %v", err)
	}
	for _, r := range result {
		a.Logger.Infof("%q file %q removed", r.TargetName, r.file)
	}

	//
	if len(errs) > 0 {
		return fmt.Errorf("there was %d error(s)", len(errs))
	}
	a.Logger.Debug("done...")
	return nil
}

func (a *App) FileRemove(ctx context.Context, t *Target) (string, error) {
	fileClient := file.NewFileClient(t.client)
	_, err := fileClient.Remove(ctx, &file.RemoveRequest{
		RemoteFile: a.Config.FileRemoveFile,
	})
	return a.Config.FileRemoveFile, err
}
