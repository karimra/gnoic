package app

import (
	"context"
	"encoding/json"
	"fmt"

	"github.com/openconfig/gnoi/healthz"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"

	"github.com/karimra/gnoic/api"
	ghealthz "github.com/karimra/gnoic/api/healthz"
)

type healthzAckResponse struct {
	TargetError
	rsp *healthz.AcknowledgeResponse
}

func (a *App) InitHealthzAckFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.HealthzAckPath, "path", "", "path to the component to acknowledge")
	cmd.Flags().StringVar(&a.Config.HealthzAckID, "id", "", "event ID to acknowledge")
	//
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunEHealthzAck(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *healthzAckResponse, numTargets)

	a.wg.Add(numTargets)
	for _, t := range targets {
		go a.HealthAckRequest(cmd.Context(), t, responseChan)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	result := make([]*healthzAckResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q Healthz Acknowledge failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
		a.printProtoMsg(rsp.TargetName, rsp.rsp)
		result = append(result, rsp)
	}

	for _, r := range result {
		switch a.Config.Format {
		case "json":
			b, err := json.MarshalIndent(r.rsp, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal to JSON: %v", err)
			}
			fmt.Println(string(b))
		default:
			fmt.Println(a.healthzGetTree(r.rsp.GetStatus(), "  "))
		}
	}
	return a.handleErrs(errs)
}

func (a *App) HealthAckRequest(ctx context.Context, t *api.Target, rspCh chan<- *healthzAckResponse) {
	defer a.wg.Done()
	ctx = t.AppendMetadata(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
	if err != nil {
		rspCh <- &healthzAckResponse{
			TargetError: TargetError{
				TargetName: t.Config.Address,
				Err:        err,
			},
		}
		return
	}
	defer t.Close()
	rspCh <- a.HealthAck(ctx, t)
}

func (a *App) HealthAck(ctx context.Context, t *api.Target) *healthzAckResponse {
	opts := []ghealthz.HealthzOption{
		ghealthz.Path(a.Config.HealthzAckPath),
		ghealthz.ID(a.Config.HealthzAckID),
	}
	req, err := ghealthz.NewAcknowledgeRequest(opts...)
	if err != nil {
		return &healthzAckResponse{
			TargetError: TargetError{
				TargetName: t.Config.Name,
				Err:        err,
			},
		}
	}
	a.printProtoMsg(t.Config.Name, req)
	hc := healthz.NewHealthzClient(t.Conn())
	rsp, err := hc.Acknowledge(ctx, req)
	return &healthzAckResponse{
		TargetError: TargetError{
			TargetName: t.Config.Name,
			Err:        err,
		},
		rsp: rsp,
	}
}
