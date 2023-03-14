package app

import (
	"context"
	"fmt"

	"github.com/karimra/gnoic/api"
	"github.com/karimra/gnoic/utils"
	"github.com/openconfig/gnoi/system"
	"github.com/openconfig/gnoi/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/encoding/prototext"
)

func (a *App) InitSystemCancelRebootFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.SystemCancelRebootMessage, "message", "", "Cancel Reboot message")
	cmd.Flags().StringArrayVar(&a.Config.SystemCancelRebootSubcomponents, "subcomponent", []string{}, "Cancel Reboot subscomponents")
	//
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunESystemCancelReboot(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}
	subcomponents := make([]*types.Path, len(a.Config.SystemCancelRebootSubcomponents))
	for i, p := range a.Config.SystemRebootStatusSubscomponents {
		subcomponents[i], err = utils.ParsePath(p)
		if err != nil {
			return err
		}
	}
	numTargets := len(targets)
	responseChan := make(chan *TargetError, numTargets)
	a.wg.Add(numTargets)
	for _, t := range targets {
		go a.systemCancelRebootRequest(cmd.Context(), t, subcomponents, responseChan)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q System CancelReboot failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
	}
	return a.handleErrs(errs)
}

func (a *App) systemCancelRebootRequest(ctx context.Context, t *api.Target, subcomponents []*types.Path, rspCh chan<- *TargetError) {
	defer a.wg.Done()
	ctx = t.AppendMetadata(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
	if err != nil {
		rspCh <- &TargetError{
			TargetName: t.Config.Address,
			Err:        err,
		}
		return
	}
	defer t.Close()
	err = a.SystemCancelReboot(ctx, t, subcomponents)
	rspCh <- &TargetError{
		TargetName: t.Config.Address,
		Err:        err,
	}
}

func (a *App) SystemCancelReboot(ctx context.Context, t *api.Target, subcomponents []*types.Path) error {
	req := &system.CancelRebootRequest{
		Message:       a.Config.SystemCancelRebootMessage,
		Subcomponents: subcomponents,
	}
	a.Logger.Debugf("%q System CancelReboot Request: %s", t.Config.Address, prototext.Format(req))
	a.printProtoMsg(t.Config.Name, req)
	_, err := t.SystemClient().CancelReboot(ctx, req)
	if err != nil {
		return err
	}
	a.Logger.Infof("%q System CancelReboot Request was successful", t.Config.Address)
	return nil
}
