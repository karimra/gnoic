package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"sort"

	"github.com/karimra/gnoic/api"
	"github.com/olekukonko/tablewriter"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

type reflectionResponse struct {
	TargetError
	rsp *reflectpb.ServerReflectionResponse
}

func (a *App) InitServicesFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunEServices(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}
	numTargets := len(targets)
	responseChan := make(chan *reflectionResponse, numTargets)
	a.wg.Add(numTargets)
	for _, t := range targets {
		go func(t *api.Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &reflectionResponse{
					TargetError: TargetError{
						TargetName: t.Config.Name,
						Err:        err,
					},
				}
				return
			}
			defer t.Close()

			rfc := reflectpb.NewServerReflectionClient(t.Conn())
			info, err := rfc.ServerReflectionInfo(ctx)
			if err != nil {
				responseChan <- &reflectionResponse{
					TargetError: TargetError{
						TargetName: t.Config.Name,
						Err:        err,
					},
				}
				return
			}
			err = info.Send(&reflectpb.ServerReflectionRequest{
				MessageRequest: &reflectpb.ServerReflectionRequest_ListServices{},
			})
			if err != nil {
				responseChan <- &reflectionResponse{
					TargetError: TargetError{
						TargetName: t.Config.Name,
						Err:        err,
					},
				}
				return
			}
			rsp, err := info.Recv()

			responseChan <- &reflectionResponse{
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
	result := make([]*reflectionResponse, 0, numTargets)

	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q Services failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
		result = append(result, rsp)
		a.printMsg(rsp.TargetName, rsp.rsp)
	}

	switch a.Config.Format {
	default:
		fmt.Println(a.reflectionServicesTable(result))
	case "json":
		for _, r := range result {
			tRsp := targetResponse{
				Target:   r.TargetName,
				Response: r.rsp,
			}
			b, err := json.MarshalIndent(tRsp, "", "  ")
			if err != nil {
				a.Logger.Errorf("failed to marshal Services response from %q: %v", r.TargetName, err)
				continue
			}
			fmt.Println(string(b))
		}
	}
	return a.handleErrs(errs)
}

func (a *App) reflectionServicesTable(r []*reflectionResponse) string {
	targetTabData := make([][]string, 0, len(r))
	sort.Slice(r, func(i, j int) bool {
		return r[i].TargetName < r[j].TargetName
	})
	for _, rsp := range r {
		switch r := rsp.rsp.MessageResponse.(type) {
		case *reflectpb.ServerReflectionResponse_ListServicesResponse:
			for _, srv := range r.ListServicesResponse.GetService() {
				targetTabData = append(targetTabData, []string{
					rsp.TargetName,
					srv.GetName(),
				})
			}

		}
	}
	//
	b := new(bytes.Buffer)
	table := tablewriter.NewWriter(b)
	table.SetHeader([]string{"Target Name", "Service"})
	formatTable(table)
	table.AppendBulk(targetTabData)
	table.Render()
	return b.String()
}
