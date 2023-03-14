package app

import (
	"context"
	"fmt"
	"strings"

	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/karimra/gnoic/api"
	gsystem "github.com/karimra/gnoic/api/system"
)

func (a *App) InitSystemKillProcessFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().Uint32Var(&a.Config.SystemKillProcessPID, "pid", 0, "PID to be restarted")
	cmd.Flags().StringVar(&a.Config.SystemKillProcessName, "name", "", "Process to be restarted")
	cmd.Flags().StringVar(&a.Config.SystemKillProcessSignal, "signal", "", "Reboot message")
	cmd.Flags().BoolVar(&a.Config.SystemKillProcessRestart, "restart", false, "Restart Process")
	//
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) PreRunESystemKillProcess(cmd *cobra.Command, args []string) error {
	a.Config.SetLocalFlagsFromFile(cmd)
	if a.Config.SystemKillProcessName == "" && a.Config.SystemKillProcessPID == 0 {
		return fmt.Errorf("specify --name or --pid to be restarted")
	}
	a.Config.SystemKillProcessSignal = strings.ToUpper(a.Config.SystemKillProcessSignal)
	switch a.Config.SystemKillProcessSignal {
	case "TERM":
	case "KILL":
	case "HUP":
	default:
		return fmt.Errorf("unknown kill signal %q", a.Config.SystemKillProcessSignal)
	}
	return nil
}

func (a *App) RunESystemKillProcess(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *TargetError, numTargets)
	a.wg.Add(numTargets)
	for _, t := range targets {
		go a.systemKillProcessRequest(cmd.Context(), t, responseChan)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q System KillProcess failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
	}
	return a.handleErrs(errs)
}

func (a *App) systemKillProcessRequest(ctx context.Context, t *api.Target, rspCh chan<- *TargetError) {
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
	err = a.SystemKillProcess(ctx, t)
	rspCh <- &TargetError{
		TargetName: t.Config.Address,
		Err:        err,
	}
}

func (a *App) SystemKillProcess(ctx context.Context, t *api.Target) error {
	req, err := gsystem.NewSystemKillProcessRequest(
		gsystem.PID(a.Config.SystemKillProcessPID),
		gsystem.ProcessName(a.Config.SystemKillProcessName),
		gsystem.Signal(a.Config.SystemKillProcessSignal),
		gsystem.ProcessRestart(a.Config.SystemKillProcessRestart),
	)
	if err != nil {
		return err
	}
	_, err = t.SystemClient().KillProcess(ctx, req)
	if err != nil {
		return err
	}
	a.Logger.Infof("%q System KillProcess Request successful", t.Config.Address)
	return nil
}
