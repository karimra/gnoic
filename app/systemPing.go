package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"strconv"
	"strings"
	"time"

	"github.com/openconfig/gnoi/system"
	"github.com/openconfig/gnoi/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/prototext"
)

type targetResponse struct {
	Target   string      `json:"target,omitempty"`
	Response interface{} `json:"response,omitempty"`
}

func (a *App) InitSystemPingFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.SystemPingDestination, "destination", "", "Destination address to ping")
	cmd.Flags().StringVar(&a.Config.SystemPingSource, "source", "", "Source address to ping from")
	cmd.Flags().Int32Var(&a.Config.SystemPingCount, "count", 0, "Number of packets")
	cmd.Flags().DurationVar(&a.Config.SystemPingInterval, "interval", time.Second, "Duration between requests")
	cmd.Flags().DurationVar(&a.Config.SystemPingWait, "wait", 0, "Duration to wait for a response")
	cmd.Flags().Int32Var(&a.Config.SystemPingSize, "size", 0, "Size of request packet. (excluding ICMP header)")
	cmd.Flags().BoolVar(&a.Config.SystemPingDoNotFragment, "do-not-fragment", false, "Set the do not fragment bit. (IPv4 destinations)")
	cmd.Flags().BoolVar(&a.Config.SystemPingDoNotResolve, "do-not-resolve", false, "Do not try resolve the address returned")
	cmd.Flags().StringVar(&a.Config.SystemPingProtocol, "protocol", "", "Layer3 protocol requested for the ping, V4 or V6, defaults to UNSPECIFIED")
	//
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) PreRunESystemPing(cmd *cobra.Command, args []string) error {
	a.Config.SetLocalFlagsFromFile(cmd)
	if a.Config.SystemPingDestination == "" {
		return errors.New("flag --destination is required")
	}
	switch strings.ToUpper(a.Config.SystemPingProtocol) {
	case "V4", "V6", "":
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
			err := a.SystemPing(ctx, t)
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
			wErr := fmt.Errorf("%q System Ping failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
	}
	return a.handleErrs(errs)
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
		L3Protocol:    getL3Protocol(a.Config.SystemPingProtocol),
	}
	a.Logger.Debugf("ping request:\n%s", prototext.Format(req))
	a.printMsg(t.Config.Name, req)
	stream, err := systemClient.Ping(ctx, req)
	if err != nil {
		a.Logger.Errorf("%q creating System Ping stream failed: %v", t.Config.Address, err)
		return err
	}
	for {
		rsp, err := stream.Recv()
		if err == io.EOF {
			a.Logger.Debugf("%q sent EOF", t.Config.Address)
			break
		}
		if err != nil && err != io.EOF {
			a.Logger.Errorf("%q rcv Ping stream failed: %v", t.Config.Address, err)
			return err
		}
		a.Logger.Debugf("ping response %s:\n%s", t.Config.Name, prototext.Format(rsp))
		a.printMsg(t.Config.Name, rsp)
		a.printPingResponse(t.Config.Name, rsp)
	}
	return nil
}

func (a *App) printPingResponse(name string, rsp *system.PingResponse) {
	switch a.Config.Format {
	case "json":
		tRsp := targetResponse{
			Target:   name,
			Response: rsp,
		}
		b, err := json.MarshalIndent(tRsp, "", "  ")
		if err != nil {
			a.Logger.Errorf("failed to marshal ping response from %q: %v", name, err)
			return
		}
		fmt.Println(string(b))
	default:
		sb := strings.Builder{}
		numAddress := len(a.Config.Address)
		if rsp.GetBytes() > 0 {
			if numAddress > 1 {
				sb.WriteString("[")
				sb.WriteString(name)
				sb.WriteString("] ")
			}
			sb.WriteString(strconv.Itoa(int(rsp.GetBytes())))
			sb.WriteString(" bytes from ")
			sb.WriteString(rsp.GetSource())
			sb.WriteString(": icmp_seq=")
			sb.WriteString(strconv.Itoa(int(rsp.GetSequence())))
			sb.WriteString(" ttl=")
			sb.WriteString(strconv.Itoa(int(rsp.GetTtl())))
			sb.WriteString(" time=")
			sb.WriteString(time.Duration(rsp.GetTime()).String())
			fmt.Println(sb.String())
			return
		}
		// summary
		// line1
		if numAddress > 1 {
			sb.WriteString("[")
			sb.WriteString(name)
			sb.WriteString("] ")
		}
		sb.WriteString("--- ")
		sb.WriteString(rsp.GetSource())
		sb.WriteString(" ping statistics ---\n")
		// line2
		if numAddress > 1 {
			sb.WriteString("[")
			sb.WriteString(name)
			sb.WriteString("] ")
		}
		sb.WriteString(strconv.Itoa(int(rsp.GetSent())))
		sb.WriteString(" packets sent, ")
		sb.WriteString(strconv.Itoa(int(rsp.GetReceived())))
		sb.WriteString(" packets received, ")
		sb.WriteString(fmt.Sprintf("%.2f%% packet loss\n", ((1 - (float32(rsp.GetReceived()) / float32(rsp.GetSent()))) * 100)))
		// line3
		if numAddress > 1 {
			sb.WriteString("[")
			sb.WriteString(name)
			sb.WriteString("] ")
		}
		sb.WriteString("round-trip min/avg/max/stddev = ")
		sb.WriteString(formatDurationMS(rsp.GetMinTime()))
		sb.WriteString("/")
		sb.WriteString(formatDurationMS(rsp.GetAvgTime()))
		sb.WriteString("/")
		sb.WriteString(formatDurationMS(rsp.GetMaxTime()))
		sb.WriteString("/")
		sb.WriteString(formatDurationMS(rsp.GetStdDev()))
		sb.WriteString(" ms")
		fmt.Println(sb.String())
		return
	}
}

func formatDurationMS(d int64) string {
	return fmt.Sprintf("%.3f", float64(d)/float64(time.Millisecond))
}

func getL3Protocol(s string) types.L3Protocol {
	switch strings.ToUpper(s) {
	case "V4":
		return types.L3Protocol_IPV4
	case "V6":
		return types.L3Protocol_IPV6
	}
	return types.L3Protocol_UNSPECIFIED
}
