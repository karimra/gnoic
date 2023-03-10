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
	"google.golang.org/grpc/metadata"
	reflectpb "google.golang.org/grpc/reflection/grpc_reflection_v1alpha"
)

type reflectionResponse struct {
	TargetError
	rsp *reflectpb.ServerReflectionResponse
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
		go a.reflectionServicesRequest(cmd.Context(), t, responseChan)
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
	a.printCMDOutput(result, a.reflectionServicesTable)
	return a.handleErrs(errs)
}

func (a *App) reflectionServicesTable(rs []*reflectionResponse) string {
	targetTabData := make([][]string, 0, len(rs))
	sort.Slice(rs, func(i, j int) bool {
		return rs[i].TargetName < rs[j].TargetName
	})
	for _, rsp := range rs {
		switch r := rsp.rsp.MessageResponse.(type) {
		case *reflectpb.ServerReflectionResponse_ListServicesResponse:
			for _, srv := range r.ListServicesResponse.GetService() {
				targetTabData = append(targetTabData, []string{
					rsp.TargetName,
					srv.GetName(),
				})
			}
		default:
			a.Logger.Printf("%s: unexpected message type: %T", rsp.TargetName, rsp.rsp)
		}
	}

	b := new(bytes.Buffer)
	table := tablewriter.NewWriter(b)
	table.SetHeader([]string{"Target Name", "Service"})
	formatTable(table)
	table.AppendBulk(targetTabData)
	table.Render()
	return b.String()
}

func (a *App) reflectionServicesRequest(ctx context.Context, t *api.Target, rspCh chan<- *reflectionResponse) {
	defer a.wg.Done()
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()
	ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

	err := t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
	if err != nil {
		rspCh <- &reflectionResponse{
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
		rspCh <- &reflectionResponse{
			TargetError: TargetError{
				TargetName: t.Config.Name,
				Err:        err,
			},
		}
		return
	}
	req := &reflectpb.ServerReflectionRequest{
		MessageRequest: &reflectpb.ServerReflectionRequest_ListServices{},
	}
	err = info.Send(req)
	if err != nil {
		rspCh <- &reflectionResponse{
			TargetError: TargetError{
				TargetName: t.Config.Name,
				Err:        err,
			},
		}
		return
	}
	rsp, err := info.Recv()
	rspCh <- &reflectionResponse{
		TargetError: TargetError{
			TargetName: t.Config.Name,
			Err:        err,
		},
		rsp: rsp,
	}
}

func (a *App) printCMDOutput(rs []*reflectionResponse, fn func([]*reflectionResponse) string) {
	switch a.Config.Format {
	default:
		fmt.Println(fn(rs))
	case "json":
		for _, r := range rs {
			tRsp := targetResponse{
				Target:   r.TargetName,
				Response: r.rsp,
			}
			b, err := json.MarshalIndent(tRsp, "", "  ")
			if err != nil {
				a.Logger.Errorf("failed to marshal Target response from %q: %v", r.TargetName, err)
				continue
			}
			fmt.Println(string(b))
		}
	}
}
