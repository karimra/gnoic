package app

import (
	"context"
	"fmt"

	"github.com/karimra/gnoic/api"
	gos "github.com/karimra/gnoic/api/os"
	gnoios "github.com/openconfig/gnoi/os"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
		go a.OsActivateRequest(cmd.Context(), t, responseChan)
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
		a.printProtoMsg(rsp.TargetName, rsp.rsp)
	}
	for _, r := range result {
		a.Logger.Infof("target %q activate response %q", r.TargetName, r.rsp)
	}
	return a.handleErrs(errs)
}

func (a *App) OsActivateRequest(ctx context.Context, t *api.Target, rspCh chan<- *osActivateResponse) {
	defer a.wg.Done()
	ctx = t.AppendMetadata(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
	if err != nil {
		rspCh <- &osActivateResponse{
			TargetError: TargetError{
				TargetName: t.Config.Name,
				Err:        err,
			},
		}
		return
	}
	defer t.Close()
	rsp, err := a.OsActivate(ctx, t)
	rspCh <- &osActivateResponse{
		TargetError: TargetError{
			TargetName: t.Config.Name,
			Err:        err,
		},
		rsp: rsp,
	}
}

func (a *App) OsActivate(ctx context.Context, t *api.Target) (*gnoios.ActivateResponse, error) {
	req, err := gos.NewActivateRequest(
		gos.Version(a.Config.OsActivateVersion),
		gos.StandbySupervisor(a.Config.OsActivateStandbySupervisor),
		gos.NoReboot(a.Config.OsActivateNoReboot),
	)
	if err != nil {
		return nil, err
	}
	a.printProtoMsg(t.Config.Name, req)
	return gnoios.NewOSClient(t.Conn()).Activate(ctx, req)
}
