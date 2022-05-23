package app

import (
	"context"
	"fmt"
	"os"
	"strings"

	"github.com/karimra/gnoic/api"
	"github.com/openconfig/gnoi/healthz"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
)

type healthzGetResponse struct {
	TargetError
	rsp *healthz.GetResponse
}

func (a *App) InitHealthzFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) InitHealthzGetFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.HealthzGetPath, "path", "", "path to the component to try to fetch healthz state for")
	//
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunEHealthzGet(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *healthzGetResponse, numTargets)

	a.wg.Add(numTargets)
	for _, t := range targets {
		go func(t *api.Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &healthzGetResponse{
					TargetError: TargetError{
						TargetName: t.Config.Address,
						Err:        err,
					},
				}
				return
			}
			defer t.Close()
			responseChan <- a.HealthzGet(ctx, t)
		}(t)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	result := make([]*healthzGetResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q Healthz Get failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
	}
	for _, r := range result {
		a.printMsg(r.TargetName, r.rsp)
		fmt.Printf("target %q:\n", r.TargetName)
		a.healthzGetTree(r.rsp.GetComponent(), "")
	}
	return a.handleErrs(errs)
}

func (a *App) HealthzGet(ctx context.Context, t *api.Target) *healthzGetResponse {
	p, err := ParsePath(a.Config.HealthzGetPath)
	if err != nil {
		return &healthzGetResponse{
			TargetError: TargetError{
				TargetName: t.Config.Name,
				Err:        err,
			},
		}
	}
	req := &healthz.GetRequest{
		Path: p,
	}
	a.printMsg(t.Config.Name, req)
	hc := healthz.NewHealthzClient(t.Conn())
	rsp, err := hc.Get(ctx, req)
	return &healthzGetResponse{
		TargetError: TargetError{
			TargetName: t.Config.Name,
			Err:        err,
		},
		rsp: rsp,
	}
}

func (a *App) healthzGetTree(comp *healthz.ComponentStatus, indent string) string {
	b := new(strings.Builder)
	fmt.Fprintf(os.Stdout, "%spath: %s\n", indent, pathToXPath(comp.GetPath()))
	fmt.Fprintf(os.Stdout, "%sstatus: %s\n", indent, comp.GetStatus().String())
	if comp.GetHealthz() != nil {
		fmt.Fprintf(os.Stdout, "%stype-url: %s\n", indent, comp.GetHealthz().GetTypeUrl())
		fmt.Fprintf(os.Stdout, "%svalue: %s\n", indent, string(comp.Healthz.GetValue()))
	}

	if len(comp.GetSubcomponents()) > 0 {
		fmt.Fprintf(os.Stdout, "%subcomponents:\n", indent)
		for i, sc := range comp.GetSubcomponents() {
			fmt.Fprintf(os.Stdout, "%s\t%d.\n", indent, i)
			b.WriteString(a.healthzGetTree(sc, indent+"\t"))
		}
	}
	return b.String()
}
