package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/olekukonko/tablewriter"
	gnoios "github.com/openconfig/gnoi/os"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"

	"github.com/karimra/gnoic/api"
	gos "github.com/karimra/gnoic/api/os"
)

type osVerifyResponse struct {
	TargetError
	rsp *gnoios.VerifyResponse
}

func (a *App) InitOSVerifyFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) PreRunEOSVerify(cmd *cobra.Command, args []string) error { return nil }
func (a *App) RunEOSVerify(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *osVerifyResponse, numTargets)

	a.wg.Add(numTargets)
	for _, t := range targets {
		go func(t *api.Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &osVerifyResponse{
					TargetError: TargetError{
						TargetName: t.Config.Name,
						Err:        err,
					},
				}
				return
			}
			defer t.Close()
			rsp, err := a.OsVerify(ctx, t)
			responseChan <- &osVerifyResponse{
				TargetError: TargetError{
					TargetName: t.Config.Name,
					Err:        err,
				},
				rsp: rsp,
			}
		}(t)
	}
	a.wg.Wait()
	close(responseChan)
	errs := make([]error, 0, numTargets)
	result := make([]*osVerifyResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q Os Verify failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
		result = append(result, rsp)
		a.printMsg(rsp.TargetName, rsp.rsp)
	}
	switch a.Config.Format {
	default:
		fmt.Println(a.osVerifyTable(result))
	case "json":
		for _, r := range result {
			tRsp := targetResponse{
				Target:   r.TargetName,
				Response: r.rsp,
			}
			b, err := json.MarshalIndent(tRsp, "", "  ")
			if err != nil {
				a.Logger.Errorf("failed to marshal os verify response from %q: %v", r.TargetName, err)
				continue
			}
			fmt.Println(string(b))
		}
	}

	return a.handleErrs(errs)
}

func (a *App) OsVerify(ctx context.Context, t *api.Target) (*gnoios.VerifyResponse, error) {
	return gnoios.NewOSClient(t.Conn()).Verify(ctx, gos.NewOSVerifyRequest())
}

func (a *App) osVerifyTable(r []*osVerifyResponse) string {
	targetTabData := make([][]string, 0, len(r))
	sort.Slice(r, func(i, j int) bool {
		return r[i].TargetName < r[j].TargetName
	})
	for _, rsp := range r {
		targetTabData = append(targetTabData, []string{
			rsp.TargetName,
			rsp.rsp.GetVersion(),
			rsp.rsp.GetActivationFailMessage(),
		})
		switch sbr := rsp.rsp.GetVerifyStandby().GetState().(type) {
		case *gnoios.VerifyStandby_StandbyState:
			switch sbr.StandbyState.GetState() {
			case gnoios.StandbyState_NON_EXISTENT:
			case gnoios.StandbyState_UNAVAILABLE:
			}
		case *gnoios.VerifyStandby_VerifyResponse:
			targetTabData = append(targetTabData, []string{
				rsp.TargetName + "." + sbr.VerifyResponse.GetId(),
				sbr.VerifyResponse.GetVersion(),
				sbr.VerifyResponse.GetActivationFailMessage(),
			})
		}

	}
	//
	b := new(bytes.Buffer)
	table := tablewriter.NewWriter(b)
	table.SetHeader([]string{"Target Name", "Version", "Activation Fail Msg"})
	formatTable(table)
	table.AppendBulk(targetTabData)
	table.Render()
	return b.String()
}
