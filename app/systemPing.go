package app

import (
	"context"
	"errors"
	"fmt"
	"io"
	"strings"
	"time"

	"github.com/openconfig/gnoi/system"
	"github.com/openconfig/gnoi/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/prototext"
)

type systemPingResponse struct {
	targetName string
	err        error
}

func (a *App) InitSystemPingFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.SystemPingDestination, "destination", "", "Destination address to ping")
	cmd.Flags().StringVar(&a.Config.SystemPingSource, "source", "", "Source address to ping from")
	cmd.Flags().Int32Var(&a.Config.SystemPingCount, "count", 0, "Number of packets")
	cmd.Flags().DurationVar(&a.Config.SystemPingInterval, "interval", time.Second, "Duration between requests")
	cmd.Flags().DurationVar(&a.Config.SystemPingWait, "wait", 0, "Duration to wait for a response")
	cmd.Flags().Int32Var(&a.Config.SystemPingSize, "size", 0, "Duration to wait for a response")
	cmd.Flags().BoolVar(&a.Config.SystemPingDoNotFragment, "do-not-fragment", false, "Set the do not fragment bit. (IPv4 destinations)")
	cmd.Flags().BoolVar(&a.Config.SystemPingDoNotResolve, "do-not-resolve", false, "Do not try resolve the address returned")
	cmd.Flags().StringVar(&a.Config.SystemPingProtocol, "protocol", "v4", "Layer3 protocol requested for the ping, IPv4 or IPv6")
	//
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) PreRunESystemPing(cmd *cobra.Command, args []string) error {
	if a.Config.SystemPingDestination == "" {
		return errors.New("flag --destination is required")
	}
	a.Config.SystemPingProtocol = "IP" + strings.ToUpper(a.Config.SystemPingProtocol)
	switch a.Config.SystemPingProtocol {
	case "IPV4", "IPV6":
	default:
		return fmt.Errorf("unknown protocol %s", a.Config.SystemPingProtocol)
	}
	return nil
}

func (a *App) RunESystemPing(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *systemPingResponse, numTargets)

	a.wg.Add(numTargets)
	for _, t := range targets {
		go func(t *Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = a.CreateGrpcClient(ctx, t, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &systemPingResponse{
					targetName: t.Config.Address,
					err:        err,
				}
				return
			}
			err := a.SystemPing(ctx, t)
			responseChan <- &systemPingResponse{
				targetName: t.Config.Address,
				err:        err,
			}
		}(t)
	}
	a.wg.Wait()
	close(responseChan)
	errs := make([]error, 0, numTargets)
	for rsp := range responseChan {
		if rsp.err != nil {
			a.Logger.Errorf("%q system ping failed: %v", rsp.targetName, rsp.err)
			errs = append(errs, rsp.err)
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

func (a *App) SystemPing(ctx context.Context, t *Target) error {
	systemClient := system.NewSystemClient(t.client)
	req := &system.PingRequest{
		Destination:   a.Config.SystemPingDestination,
		Source:        a.Config.SystemPingSource,
		Count:         a.Config.SystemPingCount,
		Interval:      a.Config.SystemPingInterval.Nanoseconds(),
		Wait:          a.Config.SystemPingWait.Nanoseconds(),
		Size:          a.Config.SystemPingSize,
		DoNotFragment: a.Config.SystemPingDoNotFragment,
		DoNotResolve:  a.Config.SystemPingDoNotResolve,
		L3Protocol:    types.L3Protocol(types.L3Protocol_value[a.Config.SystemPingProtocol]),
	}
	a.Logger.Debug(prototext.Format(req))
	stream, err := systemClient.Ping(ctx, req)
	if err != nil {
		a.Logger.Errorf("creating system ping stream failed: %v", err)
		return err
	}
	for {
		rsp, err := stream.Recv()
		if err == io.EOF {
			a.Logger.Debugf("%q sent EOF", t.Config.Address)
			break
		}
		if err != nil && err != io.EOF {
			a.Logger.Errorf("rcv system ping stream failed: %v", err)
			return err
		}
		fmt.Print(prototext.Format(rsp))
	}
	return nil
}
