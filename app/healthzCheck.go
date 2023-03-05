package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/olekukonko/tablewriter"
	"github.com/openconfig/gnoi/healthz"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"

	"github.com/karimra/gnoic/api"
	ghealthz "github.com/karimra/gnoic/api/healthz"
	"github.com/karimra/gnoic/utils"
)

type healthzCheckResponse struct {
	TargetError
	rsp *healthz.CheckResponse
}

func (a *App) InitHealthzCheckFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.HealthzCheckPath, "path", "", "path to the component to try to fetch healthz state for")
	cmd.Flags().StringVar(&a.Config.HealthzCheckID, "id", "", "event ID")
	//
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunEHealthzCheck(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *healthzCheckResponse, numTargets)

	a.wg.Add(numTargets)
	for _, t := range targets {
		go func(t *api.Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &healthzCheckResponse{
					TargetError: TargetError{
						TargetName: t.Config.Address,
						Err:        err,
					},
				}
				return
			}
			defer t.Close()
			responseChan <- a.HealthzCheck(ctx, t)
		}(t)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	result := make([]*healthzCheckResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q Healthz Check failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
		result = append(result, rsp)
	}

	for _, r := range result {
		fmt.Printf("target %q:\n", r.TargetName)
		a.printMsg(r.TargetName, r.rsp)
		switch a.Config.Format {
		case "json":
			b, err := json.MarshalIndent(r.rsp, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal to JSON: %v", err)
			}
			fmt.Println(string(b))
		default:
			s, err := healthzCheckResponseTable(result)
			if err != nil {
				return err
			}
			fmt.Println(s)
		}
	}
	return a.handleErrs(errs)
}

func (a *App) HealthzCheck(ctx context.Context, t *api.Target) *healthzCheckResponse {
	opts := []ghealthz.HealthzOption{
		ghealthz.Path(a.Config.HealthzCheckPath),
		ghealthz.ID(a.Config.HealthzCheckID),
	}
	req, err := ghealthz.NewCheckRequest(opts...)
	if err != nil {
		return &healthzCheckResponse{
			TargetError: TargetError{
				TargetName: t.Config.Name,
				Err:        err,
			},
		}
	}
	a.printMsg(t.Config.Name, req)
	hc := healthz.NewHealthzClient(t.Conn())
	rsp, err := hc.Check(ctx, req)
	return &healthzCheckResponse{
		TargetError: TargetError{
			TargetName: t.Config.Name,
			Err:        err,
		},
		rsp: rsp,
	}
}

func healthzCheckResponseTable(rsps []*healthzCheckResponse) (string, error) {
	tabData := make([][]string, 0, len(rsps))
	for _, rsp := range rsps {
		status := rsp.rsp.GetStatus()
		xpath := utils.PathToXPath(status.GetPath())
		if len(status.GetArtifacts()) > 0 {
			for _, artf := range status.GetArtifacts() {
				tabData = append(tabData, []string{
					rsp.TargetName,
					status.GetId(),
					xpath,
					status.GetStatus().String(),
					status.GetCreated().AsTime().String(),
					artf.GetId(),
					artifactType(artf),
				})
			}
		} else {
			tabData = append(tabData, []string{
				rsp.TargetName,
				status.GetId(),
				xpath,
				status.GetStatus().String(),
				status.GetCreated().AsTime().String(),
				"",
				"",
			})
		}

	}
	sort.Slice(tabData, func(i, j int) bool {
		if tabData[i][0] == tabData[j][0] {
			return tabData[i][1] < tabData[j][1]
		}
		return tabData[i][0] < tabData[j][0]
	})
	b := new(bytes.Buffer)
	table := tablewriter.NewWriter(b)
	table.SetHeader([]string{"Target Name", "ID", "Path", "Status", "Created At", "Artifact ID", "Artifact Type"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.AppendBulk(tabData)
	table.Render()
	return b.String(), nil
}
