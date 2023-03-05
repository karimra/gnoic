package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/karimra/gnoic/api"
	"github.com/karimra/gnoic/utils"
	"github.com/openconfig/gnoi/system"
	"github.com/openconfig/gnoi/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
)

func (a *App) InitSystemRebootFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.SystemRebootMethod, "method", "COLD", "Reboot method")
	cmd.Flags().DurationVar(&a.Config.SystemRebootDelay, "delay", 0, "Reboot delay")
	cmd.Flags().StringVar(&a.Config.SystemRebootMessage, "message", "", "Reboot message")
	cmd.Flags().StringArrayVar(&a.Config.SystemRebootSubscomponents, "subcomponent", []string{}, "Reboot subscomponents")
	cmd.Flags().BoolVar(&a.Config.SystemRebootForce, "force", false, "force reboot")
	//
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) PreRunESystemReboot(cmd *cobra.Command, args []string) error {
	a.Config.SetLocalFlagsFromFile(cmd)
	a.Config.SystemRebootMethod = strings.ToUpper(a.Config.SystemRebootMethod)
	switch a.Config.SystemRebootMethod {
	case "COLD":
	case "POWERDOWN":
	case "HALT":
	case "WARM":
	case "POWERUP":
	default:
		return fmt.Errorf("unknown reboot method %q", a.Config.SystemRebootMethod)
	}
	return nil
}

func (a *App) RunESystemReboot(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}
	subcomponents := make([]*types.Path, len(a.Config.SystemRebootSubscomponents))
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
		go func(t *api.Target, subcomponents []*types.Path) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &TargetError{
					TargetName: t.Config.Address,
					Err:        err,
				}
				return
			}
			defer t.Close()
			err := a.SystemReboot(ctx, t, subcomponents)
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
			wErr := fmt.Errorf("%q System Reboot failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
	}
	return a.handleErrs(errs)
}

func (a *App) SystemReboot(ctx context.Context, t *api.Target, subcomponents []*types.Path) error {
	req := &system.RebootRequest{
		Method:        system.RebootMethod(system.RebootMethod_value[a.Config.SystemRebootMethod]),
		Delay:         uint64(a.Config.SystemRebootDelay.Nanoseconds()),
		Message:       a.Config.SystemRebootMessage,
		Subcomponents: subcomponents,
		Force:         a.Config.SystemRebootForce,
	}
	_, err := t.SystemClient().Reboot(ctx, req)
	if err != nil {
		return err
	}
	a.Logger.Infof("%q System Reboot Request successful", t.Config.Address)
	return nil
}
