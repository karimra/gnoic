package app

import (
	"bytes"
	"context"
	"errors"
	"fmt"
	"sort"
	"time"

	"github.com/karimra/gnoic/api"
	"github.com/karimra/gnoic/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/openconfig/gnoi/system"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type systemSwitchControlProcessorResponse struct {
	TargetError
	rsp *system.SwitchControlProcessorResponse
}

func (a *App) InitSystemSwitchControlProcessorFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.SystemSwitchControlProcessorPath, "path", "", "Path to control processor to switch to")
	//
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunESystemSwitchControlProcessor(cmd *cobra.Command, args []string) error {
	if a.Config.SystemSwitchControlProcessorPath == "" {
		return errors.New("flag --path is required")
	}
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *systemSwitchControlProcessorResponse, numTargets)
	a.wg.Add(numTargets)
	for _, t := range targets {
		go a.systemSwitchControlProcessorRequest(cmd.Context(), t, responseChan)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	result := make([]*systemSwitchControlProcessorResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q System SwitchControlProcessor failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
		result = append(result, rsp)
	}

	s, err := systemSwitchControlProcessorTable(result)
	if err != nil {
		return err
	}
	fmt.Print(s)
	return a.handleErrs(errs)
}

func (a *App) systemSwitchControlProcessorRequest(ctx context.Context, t *api.Target, rspCh chan<- *systemSwitchControlProcessorResponse) {
	defer a.wg.Done()
	ctx = t.AppendMetadata(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
	if err != nil {
		rspCh <- &systemSwitchControlProcessorResponse{
			TargetError: TargetError{
				TargetName: t.Config.Address,
				Err:        err,
			},
		}
		return
	}
	defer t.Close()
	rsp, err := a.SystemSwitchControlProcessor(ctx, t)
	rspCh <- &systemSwitchControlProcessorResponse{
		TargetError: TargetError{
			TargetName: t.Config.Address,
			Err:        err,
		},
		rsp: rsp,
	}
}

func (a *App) SystemSwitchControlProcessor(ctx context.Context, t *api.Target) (*system.SwitchControlProcessorResponse, error) {
	p, err := utils.ParsePath(a.Config.SystemSwitchControlProcessorPath)
	if err != nil {
		return nil, err
	}
	req := &system.SwitchControlProcessorRequest{
		ControlProcessor: p,
	}
	a.printProtoMsg(t.Config.Name, req)
	rsp, err := t.SystemClient().SwitchControlProcessor(ctx, req)
	if err != nil {
		return nil, err
	}
	a.printProtoMsg(t.Config.Name, rsp)
	a.Logger.Infof("%q System SwitchControlProcessor Request successful", t.Config.Address)
	return rsp, nil
}

func systemSwitchControlProcessorTable(rsps []*systemSwitchControlProcessorResponse) (string, error) {
	tabData := make([][]string, 0, len(rsps))
	for _, rsp := range rsps {
		tabData = append(tabData, []string{
			rsp.TargetName,
			utils.PathToXPath(rsp.rsp.GetControlProcessor()),
			rsp.rsp.GetVersion(),
			time.Unix(0, rsp.rsp.GetUptime()).String(),
		})
	}
	sort.Slice(tabData, func(i, j int) bool {
		return tabData[i][0] < tabData[j][0]
	})
	b := new(bytes.Buffer)
	table := tablewriter.NewWriter(b)
	table.SetHeader([]string{"Target Name", "CP", "Version", "Uptime"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.AppendBulk(tabData)
	table.Render()
	return b.String(), nil
}
