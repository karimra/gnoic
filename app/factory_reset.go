package app

import (
	"context"
	"fmt"

	"github.com/karimra/gnoic/api"
	"github.com/openconfig/gnoi/factory_reset"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type factoryResetStartResponse struct {
	TargetError
	rsp *factory_reset.StartResponse
}

func (a *App) InitFactoryResetStartFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().BoolVar(&a.Config.FactoryResetStartFactoryOS, "factory-os", false, "instructs the Target to rollback the OS to the same version as it shipped from factory.")
	cmd.Flags().BoolVar(&a.Config.FactoryResetStartZeroFill, "zero-fill", false, "instructs the Target to zero fill persistent storage state data.")

	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunEFactoryResetStart(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *factoryResetStartResponse, numTargets)

	a.wg.Add(numTargets)
	for _, t := range targets {
		go a.factoryResetStartRequest(cmd.Context(), t, responseChan)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	result := make([]*factoryResetStartResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q FactoryReset Start failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
		result = append(result, rsp)
	}
	for _, r := range result {
		a.printProtoMsg(r.TargetName, r.rsp)
	}
	return a.handleErrs(errs)
}

func (a *App) factoryResetStartRequest(ctx context.Context, t *api.Target, rspCh chan<- *factoryResetStartResponse) {
	defer a.wg.Done()
	ctx = t.AppendMetadata(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
	if err != nil {
		rspCh <- &factoryResetStartResponse{
			TargetError: TargetError{
				TargetName: t.Config.Address,
				Err:        err,
			},
		}
		return
	}
	defer t.Close()
	rspCh <- a.factoryResetStart(ctx, t)
}

func (a *App) factoryResetStart(ctx context.Context, t *api.Target) *factoryResetStartResponse {
	req := &factory_reset.StartRequest{
		FactoryOs: a.Config.FactoryResetStartFactoryOS,
		ZeroFill:  a.Config.FactoryResetStartZeroFill,
	}
	a.printProtoMsg(t.Config.Name, req)
	fr := factory_reset.NewFactoryResetClient(t.Conn())
	rsp, err := fr.Start(ctx, req)
	return &factoryResetStartResponse{
		TargetError: TargetError{
			TargetName: t.Config.Name,
			Err:        err,
		},
		rsp: rsp,
	}
}
