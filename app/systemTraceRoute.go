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

	"github.com/karimra/gnoic/api"
	"github.com/openconfig/gnoi/system"
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
	cmd.Flags().DurationVar(&a.Config.SystemTracerouteWait, "wait", 0, "Duration to wait for a response")
	cmd.Flags().Uint32Var(&a.Config.SystemTracerouteInitialTTL, "initial-ttl", 0, "Initial TTL. (default=1)")
	cmd.Flags().Int32Var(&a.Config.SystemTracerouteMaxTTL, "max-ttl", 0, "Maximum number of hops. (default=30)")
	cmd.Flags().BoolVar(&a.Config.SystemTracerouteDoNotFragment, "do-not-fragment", false, "Set the do not fragment bit. (IPv4 destinations)")
	cmd.Flags().BoolVar(&a.Config.SystemTracerouteDoNotResolve, "do-not-resolve", false, "Do not try resolve the address returned")
	cmd.Flags().StringVarP(&a.Config.SystemTracerouteL3Protocol, "l3protocol", "3", "", "Layer3 protocol requested for the traceroute, v4 or v6, defaults to UNSPECIFIED")
	cmd.Flags().StringVarP(&a.Config.SystemTracerouteL4Protocol, "l4protocol", "4", "ICMP", "Layer4 protocol requested for the traceroute, ICMP, UDP or TCP")
	//
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) PreRunESystemTraceRoute(cmd *cobra.Command, args []string) error {
	if a.Config.SystemTracerouteDestination == "" {
		return errors.New("flag --destination is required")
	}
	switch strings.ToUpper(a.Config.SystemTracerouteL3Protocol) {
	case "V4", "V6", "":
	default:
		return fmt.Errorf("unknown L3 protocol %q", a.Config.SystemTracerouteL3Protocol)
	}
	a.Config.SystemTracerouteL4Protocol = strings.ToUpper(a.Config.SystemTracerouteL4Protocol)
	switch a.Config.SystemTracerouteL4Protocol {
	case "ICMP", "UDP", "TCP":
	default:
		return fmt.Errorf("unknown L4 protocol %q", a.Config.SystemTracerouteL4Protocol)
	}
	return nil
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
		go func(t *api.Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &TargetError{
					TargetName: t.Config.Address,
					Err:        err,
				}
				return
			}
			defer t.Close()
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
			wErr := fmt.Errorf("%q System Traceroute failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
	}
	return a.handleErrs(errs)
}

func (a *App) SystemTraceRoute(ctx context.Context, t *api.Target) error {
	systemClient := system.NewSystemClient(t.Conn())
	req := &system.TracerouteRequest{
		Destination:   a.Config.SystemTracerouteDestination,
		Source:        a.Config.SystemTracerouteSource,
		InitialTtl:    a.Config.SystemTracerouteInitialTTL,
		MaxTtl:        a.Config.SystemTracerouteMaxTTL,
		Wait:          a.Config.SystemTracerouteWait.Nanoseconds(),
		DoNotFragment: a.Config.SystemTracerouteDoNotFragment,
		DoNotResolve:  a.Config.SystemTracerouteDoNotResolve,
		L3Protocol:    getL3Protocol(a.Config.SystemTracerouteL3Protocol),
		L4Protocol:    system.TracerouteRequest_L4Protocol(system.TracerouteRequest_L4Protocol_value[a.Config.SystemTracerouteL4Protocol]),
	}
	a.Logger.Debug(prototext.Format(req))
	a.printMsg(t.Config.Name, req)
	stream, err := systemClient.Traceroute(ctx, req)
	if err != nil {
		a.Logger.Errorf("creating System Traceroute stream failed: %v", err)
		return err
	}
	for {
		rsp, err := stream.Recv()
		if err == io.EOF {
			a.Logger.Debugf("%q sent EOF", t.Config.Address)
			break
		}
		if err != nil && err != io.EOF {
			a.Logger.Errorf("rcv System Traceroute stream failed: %v", err)
			return err
		}
		a.Logger.Debugf("traceroute response %s:\n%s", t.Config.Name, prototext.Format(rsp))
		a.printMsg(t.Config.Name, req)
		a.printTracerouteResponse(t.Config.Name, rsp)
	}
	return nil
}

func (a *App) printTracerouteResponse(name string, rsp *system.TracerouteResponse) {
	switch a.Config.Format {
	case "json":
		tRsp := targetResponse{
			Target:   name,
			Response: rsp,
		}
		b, err := json.MarshalIndent(tRsp, "", "  ")
		if err != nil {
			a.Logger.Errorf("failed to marshal traceroute response from %q: %v", name, err)
			return
		}
		fmt.Println(string(b))
	default:
		sb := strings.Builder{}
		if len(a.Config.Address) > 1 {
			sb.WriteString("[")
			sb.WriteString(name)
			sb.WriteString("] ")
		}
		// Fist msg
		if rsp.DestinationAddress != "" {
			sb.WriteString("Traceroute to ")
			sb.WriteString(rsp.DestinationName)
			sb.WriteString(" (")
			sb.WriteString(rsp.DestinationAddress)
			sb.WriteString("), ")
			sb.WriteString(strconv.Itoa(int(rsp.GetHops())))
			sb.WriteString(" max hops, ")
			sb.WriteString(strconv.Itoa(int(rsp.GetPacketSize())))
			sb.WriteString(" byte packet size")
			fmt.Println(sb.String())
			return
		}
		// rest of messages
		// Hop index
		sb.WriteString(strconv.Itoa(int(rsp.Hop)))
		sb.WriteString(" ")
		switch rsp.State {
		case system.TracerouteResponse_DEFAULT:
			// hostname
			sb.WriteString(rsp.Name)
			sb.WriteString(" ")
			// IP Address
			sb.WriteString("(")
			sb.WriteString(rsp.Address)
			sb.WriteString(") ")
			// AS path
			numAsPath := len(rsp.AsPath)
			if numAsPath > 0 {
				sb.WriteString("[")
				for i := range rsp.AsPath {
					if i+1 == numAsPath {
						sb.WriteString("AS")
						sb.WriteString(strconv.Itoa(int(rsp.AsPath[i])))
						continue
					}
					sb.WriteString("AS")
					sb.WriteString(strconv.Itoa(int(rsp.AsPath[i])))
					sb.WriteString(" ")
				}
				sb.WriteString("] ")
			}
			// MPLS
			if len(rsp.Mpls) > 0 {
				sb.WriteString(fmt.Sprintf("<MPLS:L=%s,E=%s,S=%s,T=%s> ", rsp.Mpls["Label"], rsp.Mpls["E"], rsp.Mpls["S"], rsp.Mpls["TTL"]))
			}
			// RTT
			sb.WriteString(time.Duration(rsp.Rtt).String())
		case system.TracerouteResponse_ICMP:
			sb.WriteString(system.TracerouteResponse_State_name[int32(rsp.State)])
			sb.WriteString(" ")
			sb.WriteString("code=")
			sb.WriteString(strconv.Itoa(int(rsp.GetIcmpCode())))
		case system.TracerouteResponse_NONE:
			sb.WriteString("No Response")
		default:
			if state, ok := system.TracerouteResponse_State_name[int32(rsp.State)]; ok {
				sb.WriteString(state)
			} else {
				sb.WriteString("unexpected state value ")
				sb.WriteString(rsp.State.String())
			}
		}
		fmt.Println(sb.String())
	}
}
