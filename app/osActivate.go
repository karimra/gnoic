package app

import (
	"context"
	"fmt"

	gos "github.com/karimra/gnoic/api/os"
	gnoios "github.com/openconfig/gnoi/os"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
)

type osActivateResponse struct {
	TargetError
	rsp *gnoios.ActivateResponse
}

func (a *App) InitOSActivateFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.OsActivateVersion, "version", "", "package version to be activated")
	cmd.Flags().BoolVar(&a.Config.OsActivateStandbySupervisor, "standby", false, "activate on standby supervisor")
	cmd.Flags().BoolVar(&a.Config.OsActivateNoReboot, "no-reboot", false, "do not reboot after activation")
	//
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) PreRunEOSActivate(cmd *cobra.Command, args []string) error { return nil }
func (a *App) RunEOSActivate(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *osActivateResponse, numTargets)

	a.wg.Add(numTargets)
	for _, t := range targets {
		go func(t *Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = a.CreateGrpcClient(ctx, t, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &osActivateResponse{
					TargetError: TargetError{
						TargetName: t.Config.Name,
						Err:        err,
					},
				}
				return
			}
			rsp, err := a.OsActivate(ctx, t)
			responseChan <- &osActivateResponse{
				TargetError: TargetError{
					TargetName: t.Config.Name,
					Err:        err,
				},
				rsp: rsp,
			}
		}(t)
	}
	a.wg.Wait()
	close(responseChan)
	errs := make([]error, 0, numTargets)
	result := make([]*osActivateResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q Os Activate failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
		result = append(result, rsp)
		a.printMsg(rsp.TargetName, rsp.rsp)
	}
	for _, r := range result {
		a.Logger.Infof("target %q activate response %q", r.TargetName, r.rsp)
	}
	return a.handleErrs(errs)
}

func (a *App) OsActivate(ctx context.Context, t *Target) (*gnoios.ActivateResponse, error) {
	req, err := gos.NewActivateRequest(
		gos.Version(a.Config.OsActivateVersion),
		gos.StandbySupervisor(a.Config.OsActivateStandbySupervisor),
		gos.NoReboot(a.Config.OsActivateNoReboot),
	)
	if err != nil {
		return nil, err
	}
	a.printMsg(t.Config.Name, req)
	return gnoios.NewOSClient(t.client).Activate(ctx, req)
}
