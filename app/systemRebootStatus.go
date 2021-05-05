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
	"github.com/openconfig/gnoi/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/prototext"
)

type systemRebootStatusResponse struct {
	targetName string
	rsp        *system.RebootStatusResponse
	err        error
}

func (a *App) InitSystemRebootStatusFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringArrayVar(&a.Config.SystemRebootStatusSubscomponents, "subcomponent", []string{}, "Reboot subscomponents")
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
	subcomponents := make([]*types.Path, len(a.Config.SystemRebootStatusSubscomponents))
	for i, p := range a.Config.SystemRebootStatusSubscomponents {
		subcomponents[i], err = ParsePath(p)
		if err != nil {
			return err
		}
	}
	numTargets := len(targets)
	responseChan := make(chan *systemRebootStatusResponse, numTargets)
	a.wg.Add(numTargets)
	for _, t := range targets {
		go func(t *Target, subcomponents []*types.Path) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = a.CreateGrpcClient(ctx, t, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &systemRebootStatusResponse{
					targetName: t.Config.Address,
					err:        err,
				}
				return
			}
			rsp, err := a.SystemRebootStatus(ctx, t, subcomponents)
			responseChan <- &systemRebootStatusResponse{
				targetName: t.Config.Address,
				rsp:        rsp,
				err:        err,
			}
		}(t, subcomponents)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	result := make([]*systemRebootStatusResponse, 0, numTargets)

	for rsp := range responseChan {
		if rsp.err != nil {
			a.Logger.Errorf("%q system reboot status failed: %v", rsp.targetName, rsp.err)
			errs = append(errs, rsp.err)
			continue
		}
		result = append(result, rsp)
	}

	for _, err := range errs {
		a.Logger.Errorf("err: %v", err)
	}
	s, err := SystemRebootStatusTable(result)
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

func (a *App) SystemRebootStatus(ctx context.Context, t *Target, subcomponents []*types.Path) (*system.RebootStatusResponse, error) {
	systemClient := system.NewSystemClient(t.client)
	req := &system.RebootStatusRequest{
		Subcomponents: subcomponents,
	}

	resp, err := systemClient.RebootStatus(ctx, req)
	if err != nil {
		return nil, err
	}
	prototext.Format(resp)
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
		tabData = append(tabData, []string{
			rsp.targetName,
			fmt.Sprintf("%t", rsp.rsp.GetActive()),
			time.Duration(rsp.rsp.GetWait()).String(),
			time.Unix(0, int64(rsp.rsp.GetWhen())).String(),
			rsp.rsp.GetReason(),
			strconv.FormatUint(uint64(rsp.rsp.GetCount()), 10),
		})
	}

	sort.Slice(tabData, func(i, j int) bool {
		return tabData[i][0] < tabData[j][0]
	})
	b := new(bytes.Buffer)
	table := tablewriter.NewWriter(b)
	table.SetHeader([]string{"Target Name", "Subcomponents", "Active", "Duration Until Reboot", "Reboot Time", "Reason", "Count"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.AppendBulk(tabData)
	table.Render()
	return b.String(), nil
}
