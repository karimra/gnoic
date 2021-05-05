package app

import (
	"context"
	"fmt"

	"github.com/openconfig/gnoi/system"
	"github.com/openconfig/gnoi/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
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
	subcomponents := make([]*types.Path, len(a.Config.SystemRebootSubscomponents))
	for i, p := range a.Config.SystemRebootStatusSubscomponents {
		subcomponents[i], err = ParsePath(p)
		if err != nil {
			return err
		}
	}
	numTargets := len(targets)
	responseChan := make(chan *TargetError, numTargets)
	a.wg.Add(numTargets)
	for _, t := range targets {
		go func(t *Target, subcomponents []*types.Path) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = a.CreateGrpcClient(ctx, t, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &TargetError{
					TargetName: t.Config.Address,
					Err:        err,
				}
				return
			}
			err := a.SystemCancelReboot(ctx, t, subcomponents)
			responseChan <- &TargetError{
				TargetName: t.Config.Address,
				Err:        err,
			}
		}(t, subcomponents)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			a.Logger.Errorf("%q system cancel reboot failed: %v", rsp.TargetName, rsp.Err)
			errs = append(errs, rsp.Err)
			continue
		}
	}
	for _, err := range errs {
		a.Logger.Errorf("err: %v", err)
	}

	//
	if len(errs) > 0 {
		return fmt.Errorf("there was %d error(s)", len(errs))
	}
	a.Logger.Debug("done...")

	return nil
}

func (a *App) SystemCancelReboot(ctx context.Context, t *Target, subcomponents []*types.Path) error {
	systemClient := system.NewSystemClient(t.client)
	req := &system.CancelRebootRequest{
		Message:       a.Config.SystemCancelRebootMessage,
		Subcomponents: subcomponents,
	}
	_, err := systemClient.CancelReboot(ctx, req)
	if err != nil {
		return err
	}
	a.Logger.Infof("%q CancelReboot request successful", t.Config.Address)
	return nil
}
