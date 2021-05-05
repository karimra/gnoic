package app

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/openconfig/gnoi/system"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
)

type systemTimeResponse struct {
	targetName string
	rsp        *system.TimeResponse
	err        error
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
		go func(t *Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = a.CreateGrpcClient(ctx, t, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &systemTimeResponse{
					targetName: t.Config.Address,
					err:        err,
				}
				return
			}
			rsp, err := a.SystemTime(ctx, t)
			responseChan <- &systemTimeResponse{
				targetName: t.Config.Address,
				rsp:        rsp,
				err:        err,
			}
		}(t)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	result := make([]*systemTimeResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.err != nil {
			a.Logger.Errorf("%q system time failed: %v", rsp.targetName, rsp.err)
			errs = append(errs, rsp.err)
			continue
		}
		result = append(result, rsp)
	}
	for _, err := range errs {
		a.Logger.Errorf("err: %v", err)
	}
	s, err := systemTimeTable(result)
	if err != nil {
		return err
	}
	fmt.Print(s)
	//
	if len(errs) > 0 {
		return fmt.Errorf("there was %d error(s)", len(errs))
	}
	a.Logger.Debug("done...")
	return nil
}

func (a *App) SystemTime(ctx context.Context, t *Target) (*system.TimeResponse, error) {
	systemClient := system.NewSystemClient(t.client)
	return systemClient.Time(ctx, new(system.TimeRequest))
}

func systemTimeTable(rsps []*systemTimeResponse) (string, error) {
	tabData := make([][]string, 0, len(rsps))
	for _, rsp := range rsps {
		tabData = append(tabData, []string{
			rsp.targetName,
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
