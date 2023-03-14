package app

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"strings"
	"time"

	"github.com/karimra/gnoic/api"
	gsystem "github.com/karimra/gnoic/api/system"
	"github.com/openconfig/gnoi/system"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
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
	cmd.Flags().BoolVar(&a.Config.SystemTracerouteDoNotLookupAsn, "do-not-lookup-asn", false, "Do not try to lookup ASN")
	//
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) PreRunESystemTraceRoute(cmd *cobra.Command, args []string) error {
	if a.Config.SystemTracerouteDestination == "" {
		return errors.New("flag --destination is required")
	}
	switch v := strings.ToUpper(a.Config.SystemTracerouteL3Protocol); v {
	case "V4", "V6":
		a.Config.SystemTracerouteL3Protocol = "IP" + v
	case "":
		a.Config.SystemTracerouteL3Protocol = "UNSPECIFIED"
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
		go a.systemTraceRouteRequest(cmd.Context(), t, responseChan)
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

func (a *App) systemTraceRouteRequest(ctx context.Context, t *api.Target, rspCh chan<- *TargetError) {
	defer a.wg.Done()
	ctx = t.AppendMetadata(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
	if err != nil {
		rspCh <- &TargetError{
			TargetName: t.Config.Address,
			Err:        err,
		}
		return
	}
	defer t.Close()
	err = a.SystemTraceRoute(ctx, t)
	rspCh <- &TargetError{
		TargetName: t.Config.Address,
		Err:        err,
	}
}

func (a *App) SystemTraceRoute(ctx context.Context, t *api.Target) error {
	req, err := gsystem.NewSystemTracerouteRequest(
		gsystem.Destination(a.Config.SystemTracerouteDestination),
		gsystem.Source(a.Config.SystemTracerouteSource),
		gsystem.InitialTTL(a.Config.SystemTracerouteInitialTTL),
		gsystem.TTL(a.Config.SystemTracerouteMaxTTL),
		gsystem.Wait(a.Config.SystemTracerouteWait.Nanoseconds()),
		gsystem.DoNotFragment(a.Config.SystemTracerouteDoNotFragment),
		gsystem.DoNotResolve(a.Config.SystemTracerouteDoNotResolve),
		gsystem.L3Protocol(a.Config.SystemTracerouteL3Protocol),
		gsystem.L4Protocol(a.Config.SystemTracerouteL4Protocol),
		gsystem.DoNotLookupAsn(a.Config.SystemTracerouteDoNotLookupAsn),
	)
	if err != nil {
		return err
	}
	a.Logger.Debug(prototext.Format(req))
	a.printProtoMsg(t.Config.Name, req)
	stream, err := t.SystemClient().Traceroute(ctx, req)
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
		a.printProtoMsg(t.Config.Name, rsp)
		a.printTracerouteResponse(t.Config.Name, rsp)
	}
	return nil
}

func (a *App) printTracerouteResponse(name string, rsp *system.TracerouteResponse) {
	a.pm.Lock()
	defer a.pm.Unlock()

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
		fmt.Fprintln(os.Stdout, string(b))
	default:
		sb := &strings.Builder{}
		if len(a.Config.Address) > 1 {
			sb.WriteString("[")
			sb.WriteString(name)
			sb.WriteString("] ")
		}
		// Fist msg
		if rsp.DestinationAddress != "" {
			fmt.Fprint(sb, "traceroute to ")
			tracerouteHostnameIPString(sb, rsp.DestinationName, rsp.DestinationAddress)
			fmt.Fprintf(sb, ", %d max hops, %d byte packets",
				rsp.GetHops(),
				rsp.GetPacketSize(),
			)
			fmt.Fprintln(os.Stdout, sb.String())
			return
		}
		// rest of messages
		// Hop index
		fmt.Fprintf(sb, "%d ", rsp.GetHop())
		// hostname (IP)
		tracerouteHostnameIPString(sb, rsp.Name, rsp.Address)
		// AS path
		tracerouteASPathString(sb, rsp.GetAsPath())
		// MPLS
		if len(rsp.Mpls) > 0 {
			fmt.Fprintf(sb, "<MPLS:L=%s,E=%s,S=%s,T=%s> ", rsp.Mpls["Label"], rsp.Mpls["E"], rsp.Mpls["S"], rsp.Mpls["TTL"])
		}
		// RTT
		if rsp.Rtt != 0 {
			fmt.Fprintf(sb, " %s", time.Duration(rsp.Rtt).String())
		}
		tracerouteStateString(sb, rsp)
		fmt.Fprintln(os.Stdout, sb.String())
	}
}

func tracerouteHostnameIPString(sb *strings.Builder, name, addr string) {
	if name == "" && addr == "" {
		return
	}
	if name == "" {
		fmt.Fprintf(sb, "%s (%s)", addr, addr)
		return
	}
	fmt.Fprintf(sb, "%s (%s)", name, addr)
}

func tracerouteASPathString(sb *strings.Builder, as []int32) {
	numAsPath := len(as)
	if numAsPath == 0 {
		return
	}
	sb.WriteString("[")
	for i := range as {
		fmt.Fprintf(sb, "AS%d", as[i])
		if i+1 != numAsPath {
			sb.WriteString(" ")
		}
	}
	sb.WriteString("] ")
}

func tracerouteStateString(sb *strings.Builder, rsp *system.TracerouteResponse) {
	switch rsp.State {
	case system.TracerouteResponse_DEFAULT:
	case system.TracerouteResponse_NONE:
		fmt.Fprint(sb, " !No Response")
	case system.TracerouteResponse_UNKNOWN:
		fmt.Fprint(sb, " !Unknown Response Type")
	case system.TracerouteResponse_ICMP:
		if rsp.IcmpCode != 0 {
			fmt.Fprintf(sb, " !%d", rsp.IcmpCode)
		}
	case system.TracerouteResponse_HOST_UNREACHABLE:
		fmt.Fprint(sb, " !H")
	case system.TracerouteResponse_NETWORK_UNREACHABLE:
		fmt.Fprint(sb, " !N")
	case system.TracerouteResponse_PROTOCOL_UNREACHABLE:
		fmt.Fprint(sb, " !P")
	case system.TracerouteResponse_SOURCE_ROUTE_FAILED:
		fmt.Fprint(sb, " !S")
	case system.TracerouteResponse_FRAGMENTATION_NEEDED:
		fmt.Fprint(sb, " !F")
	case system.TracerouteResponse_PROHIBITED:
		fmt.Fprint(sb, " !X")
	case system.TracerouteResponse_PRECEDENCE_VIOLATION:
		fmt.Fprint(sb, " !V")
	case system.TracerouteResponse_PRECEDENCE_CUTOFF:
		fmt.Fprint(sb, " !C")
	default:
		fmt.Fprint(sb, " !unexpected response state")
	}
}
