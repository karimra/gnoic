package app

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/karimra/gnoic/api"
	"github.com/karimra/gnoic/utils"
	"github.com/olekukonko/tablewriter"
	"github.com/openconfig/gnoi/system"
	"github.com/openconfig/gnoi/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/protobuf/encoding/prototext"
)

type systemRebootStatusResponse struct {
	TargetError
	rsp *system.RebootStatusResponse
}

func (a *App) InitSystemRebootStatusFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringArrayVar(&a.Config.SystemRebootStatusSubcomponents, "subcomponent", []string{}, "Reboot subcomponents")
	//
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunESystemRebootStatus(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}
	subcomponents := make([]*types.Path, len(a.Config.SystemRebootStatusSubcomponents))
	for i, p := range a.Config.SystemRebootStatusSubcomponents {
		subcomponents[i], err = utils.ParsePath(p)
		if err != nil {
			return err
		}
	}
	numTargets := len(targets)
	responseChan := make(chan *systemRebootStatusResponse, numTargets)
	a.wg.Add(numTargets)
	for _, t := range targets {
		go a.systemRebootStatusRequest(cmd.Context(), t, subcomponents, responseChan)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	result := make([]*systemRebootStatusResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q System Reboot Status failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
		result = append(result, rsp)
	}

	s, err := SystemRebootStatusTable(result)
	if err != nil {
		return err
	}
	fmt.Print(s)
	return a.handleErrs(errs)
}

func (a *App) systemRebootStatusRequest(ctx context.Context, t *api.Target, subcomponents []*types.Path, rspCh chan<- *systemRebootStatusResponse) {
	defer a.wg.Done()
	ctx = t.AppendMetadata(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
	if err != nil {
		rspCh <- &systemRebootStatusResponse{
			TargetError: TargetError{
				TargetName: t.Config.Address,
				Err:        err,
			},
		}
		return
	}
	defer t.Close()
	rsp, err := a.SystemRebootStatus(ctx, t, subcomponents)
	rspCh <- &systemRebootStatusResponse{
		TargetError: TargetError{
			TargetName: t.Config.Address,
			Err:        err,
		},
		rsp: rsp,
	}
}

func (a *App) SystemRebootStatus(ctx context.Context, t *api.Target, subcomponents []*types.Path) (*system.RebootStatusResponse, error) {
	req := &system.RebootStatusRequest{
		Subcomponents: subcomponents,
	}

	resp, err := t.SystemClient().RebootStatus(ctx, req)
	if err != nil {
		return nil, err
	}
	a.Logger.Debugf("%q response: %s", t.Config.Address, prototext.Format(resp))
	a.Logger.Infof("%q rebootStatus active=%v, timeTillReboot=%s, rebootTime=%s, rebootCount=%d",
		t.Config.Address, resp.Active,
		time.Duration(resp.Wait), time.Unix(0, int64(resp.When)).String(),
		resp.Count,
	)
	return resp, nil
}

func SystemRebootStatusTable(rsps []*systemRebootStatusResponse) (string, error) {
	tabData := make([][]string, 0, len(rsps))
	for _, rsp := range rsps {
		rebootTime := ""
		if rsp.rsp.GetWhen() > 0 {
			rebootTime = time.Unix(0, int64(rsp.rsp.GetWhen())).Format(time.RFC3339)
		}
		tabData = append(tabData, []string{
			rsp.TargetName,
			fmt.Sprintf("%t", rsp.rsp.GetActive()),
			time.Duration(rsp.rsp.GetWait()).String(),
			rebootTime,
			rsp.rsp.GetReason(),
			strconv.FormatUint(uint64(rsp.rsp.GetCount()), 10),
		})
	}

	sort.Slice(tabData, func(i, j int) bool {
		return tabData[i][0] < tabData[j][0]
	})
	b := new(bytes.Buffer)
	table := tablewriter.NewWriter(b)
	table.SetHeader(
		[]string{"Target Name", "Active",
			"Duration Until Reboot", "Reboot Time", "Reason", "Count"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.AppendBulk(tabData)
	table.Render()
	return b.String(), nil
}
