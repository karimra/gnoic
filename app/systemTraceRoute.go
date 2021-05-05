package app

import (
	"context"
	"fmt"
	"io"
	"time"

	"github.com/openconfig/gnoi/system"
	"github.com/openconfig/gnoi/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/prototext"
)

func (a *App) InitSystemTracerouteFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.SystemTracerouteDestination, "destination", "", "Destination address to traceroute")
	cmd.Flags().StringVar(&a.Config.SystemTracerouteSource, "source", "", "Source address to traceroute from")
	cmd.Flags().DurationVar(&a.Config.SystemTracerouteInterval, "interval", time.Second, "Duration between requests")
	cmd.Flags().DurationVar(&a.Config.SystemTracerouteWait, "wait", 0, "Duration to wait for a response")
	cmd.Flags().Uint32Var(&a.Config.SystemTracerouteInitialTTL, "initial-ttl", 0, "Initial TTL. (default=1)")
	cmd.Flags().Int32Var(&a.Config.SystemTracerouteMaxTTL, "max-ttl", 0, "Maximum number of hops. (default=30)")
	cmd.Flags().Int32Var(&a.Config.SystemTracerouteSize, "size", 0, "Duration to wait for a response")
	cmd.Flags().BoolVar(&a.Config.SystemTracerouteDoNotFragment, "do-not-fragment", false, "Set the do not fragment bit. (IPv4 destinations)")
	cmd.Flags().BoolVar(&a.Config.SystemTracerouteDoNotResolve, "do-not-resolve", false, "Do not try resolve the address returned")
	cmd.Flags().StringVarP(&a.Config.SystemTracerouteL3Protocol, "l3protocol", "3", "v4", "Layer3 protocol requested for the traceroute, IPv4 or IPv6")
	cmd.Flags().StringVarP(&a.Config.SystemTracerouteL4Protocol, "l4protocol", "4", "ICMP", "Layer4 protocol requested for the traceroute, ICMP, UDP or TCP")
	//
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunESystemTraceRoute(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *TargetError, numTargets)

	a.wg.Add(numTargets)
	for _, t := range targets {
		go func(t *Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = a.CreateGrpcClient(ctx, t, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &TargetError{
					TargetName: t.Config.Address,
					Err:        err,
				}
				return
			}
			err := a.SystemTraceRoute(ctx, t)
			responseChan <- &TargetError{
				TargetName: t.Config.Address,
				Err:        err,
			}
		}(t)
	}
	a.wg.Wait()
	close(responseChan)
	errs := make([]error, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			a.Logger.Errorf("%q system ping failed: %v", rsp.TargetName, rsp.Err)
			errs = append(errs, rsp.Err)
			continue
		}
	}

	for _, err := range errs {
		a.Logger.Errorf("err: %v", err)
	}

	//
	if len(errs) > 0 {
		return fmt.Errorf("there was %d error(s)", len(errs))
	}
	a.Logger.Debug("done...")
	return nil
}

func (a *App) SystemTraceRoute(ctx context.Context, t *Target) error {
	systemClient := system.NewSystemClient(t.client)
	req := &system.TracerouteRequest{
		Destination:   a.Config.SystemTracerouteDestination,
		Source:        a.Config.SystemTracerouteSource,
		InitialTtl:    a.Config.SystemTracerouteInitialTTL,
		MaxTtl:        a.Config.SystemTracerouteMaxTTL,
		Wait:          a.Config.SystemTracerouteWait.Nanoseconds(),
		DoNotFragment: a.Config.SystemTracerouteDoNotFragment,
		DoNotResolve:  a.Config.SystemTracerouteDoNotResolve,
		L3Protocol:    types.L3Protocol(types.L3Protocol_value[a.Config.SystemTracerouteL3Protocol]),
		L4Protocol:    system.TracerouteRequest_L4Protocol(system.TracerouteRequest_L4Protocol_value[a.Config.SystemTracerouteL3Protocol]),
	}
	a.Logger.Debug(prototext.Format(req))
	stream, err := systemClient.Traceroute(ctx, req)
	if err != nil {
		a.Logger.Errorf("creating system traceroute stream failed: %v", err)
		return err
	}
	for {
		rsp, err := stream.Recv()
		if err == io.EOF {
			a.Logger.Debugf("%q sent EOF", t.Config.Address)
			break
		}
		if err != nil && err != io.EOF {
			a.Logger.Errorf("rcv system traceroute stream failed: %v", err)
			return err
		}
		fmt.Print(prototext.Format(rsp))
	}
	return nil
}
