package app

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/karimra/gnoic/api"
	gsystem "github.com/karimra/gnoic/api/system"
	"github.com/olekukonko/tablewriter"
	"github.com/openconfig/gnoi/system"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type systemTimeResponse struct {
	TargetError
	rsp *system.TimeResponse
}

func (a *App) InitSystemTimeFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	//
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunESystemTime(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *systemTimeResponse, numTargets)
	a.wg.Add(numTargets)
	for _, t := range targets {
		go a.systemTimeRequest(cmd.Context(), t, responseChan)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	result := make([]*systemTimeResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q System Time failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
		result = append(result, rsp)
	}
	s, err := systemTimeTable(result)
	if err != nil {
		return err
	}
	fmt.Print(s)
	return a.handleErrs(errs)
}

func (a *App) systemTimeRequest(ctx context.Context, t *api.Target, rspCh chan<- *systemTimeResponse) {
	defer a.wg.Done()
	ctx = t.AppendMetadata(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
	if err != nil {
		rspCh <- &systemTimeResponse{
			TargetError: TargetError{
				TargetName: t.Config.Address,
				Err:        err,
			},
		}
		return
	}
	defer t.Close()
	rsp, err := t.SystemClient().Time(ctx, gsystem.NewSystemTimeRequest())
	rspCh <- &systemTimeResponse{
		TargetError: TargetError{
			TargetName: t.Config.Address,
			Err:        err,
		},
		rsp: rsp,
	}
}

func systemTimeTable(rsps []*systemTimeResponse) (string, error) {
	tabData := make([][]string, 0, len(rsps))
	for _, rsp := range rsps {
		tabData = append(tabData, []string{
			rsp.TargetName,
			time.Unix(0, int64(rsp.rsp.GetTime())).String(),
			strconv.FormatUint(rsp.rsp.GetTime(), 10),
		})
	}
	// TODO: calc delta
	sort.Slice(tabData, func(i, j int) bool {
		return tabData[i][0] < tabData[j][0]
	})
	b := new(bytes.Buffer)
	table := tablewriter.NewWriter(b)
	table.SetHeader([]string{"Target Name", "Time", "Timestamp"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.AppendBulk(tabData)
	table.Render()
	return b.String(), nil
}
