package system

import (
	"fmt"
	"strings"
	"time"

	gnoisystem "github.com/openconfig/gnoi/system"
	"github.com/openconfig/gnoi/types"
	"google.golang.org/protobuf/proto"

	"github.com/karimra/gnoic/api"
)

type SystemOption func(proto.Message) error

// apply is a helper function that simply applies the options to the proto.Message.
// It returns an error if any of the options fails.
func apply(m proto.Message, opts ...SystemOption) error {
	for _, o := range opts {
		if err := o(m); err != nil {
			return err
		}
	}
	return nil
}

func Destination(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Destination: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.PingRequest:
			msg.Destination = s
		case *gnoisystem.TracerouteRequest:
			msg.Destination = s
		case *gnoisystem.TracerouteResponse:
			msg.DestinationAddress = s
		default:
			return fmt.Errorf("option Destination: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func DestinationName(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option DestinationName: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.TracerouteResponse:
			msg.DestinationName = s
		default:
			return fmt.Errorf("option DestinationName: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Source(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Source: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.PingRequest:
			msg.Source = s
		case *gnoisystem.PingResponse:
			msg.Source = s
		case *gnoisystem.TracerouteRequest:
			msg.Source = s
		default:
			return fmt.Errorf("option Source: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Count(i int32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Count: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.PingRequest:
			msg.Count = i
		default:
			return fmt.Errorf("option Count: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Interval(i int64) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Interval: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.PingRequest:
			msg.Interval = i
		default:
			return fmt.Errorf("option Interval: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Wait(i int64) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Wait: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.PingRequest:
			msg.Wait = i
		case *gnoisystem.TracerouteRequest:
			msg.Wait = i
		default:
			return fmt.Errorf("option Wait: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Size(i int32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Size: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.PingRequest:
			msg.Size = i
		case *gnoisystem.TracerouteResponse:
			msg.PacketSize = i
		default:
			return fmt.Errorf("option Size: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func DoNotFragment(b bool) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option DoNotFragment: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.PingRequest:
			msg.DoNotFragment = b
		case *gnoisystem.TracerouteRequest:
			msg.DoNotFragment = b
		default:
			return fmt.Errorf("option DoNotFragment: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func DoNotResolve(b bool) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option DoNotResolve: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.PingRequest:
			msg.DoNotResolve = b
		case *gnoisystem.TracerouteRequest:
			msg.DoNotResolve = b
		default:
			return fmt.Errorf("optionDoNotResolve: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func L3Protocol(p string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option L3Protocol: %w", api.ErrInvalidMsgType)
		}
		l3p, ok := types.L3Protocol_value[strings.ToUpper(p)]
		if !ok {
			return fmt.Errorf("option L3Protocol: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.PingRequest:
			msg.L3Protocol = types.L3Protocol(l3p)
		case *gnoisystem.TracerouteRequest:
			msg.L3Protocol = types.L3Protocol(l3p)
		default:
			return fmt.Errorf("option L3Protocol: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func L3ProtocolIPv4() func(msg proto.Message) error {
	return L3Protocol("IPV4")
}

func L3ProtocolIPv6() func(msg proto.Message) error {
	return L3Protocol("IPV6")
}

func L3ProtocolUNSPECIFIED() func(msg proto.Message) error {
	return L3Protocol("UNSPECIFIED")
}

func L3ProtocolCustom(i int32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option L3Protocol: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.PingRequest:
			msg.L3Protocol = types.L3Protocol(i)
		case *gnoisystem.TracerouteRequest:
			msg.L3Protocol = types.L3Protocol(i)
		default:
			return fmt.Errorf("option L3Protocol: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Time(i int64) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Time: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.PingResponse:
			msg.Time = i
		default:
			return fmt.Errorf("option Time: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func CurrentTime(i uint64) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option CurrentTime: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.TimeResponse:
			msg.Time = i
		default:
			return fmt.Errorf("option CurrentTime: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func CurrentTimeNow() func(msg proto.Message) error {
	return CurrentTime(uint64(time.Now().Nanosecond()))
}

func Sent(i int32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Sent: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.PingResponse:
			msg.Sent = i
		default:
			return fmt.Errorf("option Sent: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Received(i int32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Received: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.PingResponse:
			msg.Received = i
		default:
			return fmt.Errorf("option Received: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func MinTime(i int64) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option MinTime: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.PingResponse:
			msg.MinTime = i
		default:
			return fmt.Errorf("option MinTime: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func AvgTime(i int64) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option AvgTime: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.PingResponse:
			msg.AvgTime = i
		default:
			return fmt.Errorf("option AvgTime: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func MaxTime(i int64) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option MaxTime: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.PingResponse:
			msg.MaxTime = i
		default:
			return fmt.Errorf("option MaxTime: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func StdDev(i int64) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option StdDev: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.PingResponse:
			msg.StdDev = i
		default:
			return fmt.Errorf("option StdDev: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Bytes(i int32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Bytes: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.PingResponse:
			msg.Bytes = i
		default:
			return fmt.Errorf("option Bytes: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Sequence(i int32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Sequence: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.PingResponse:
			msg.Sequence = i
		default:
			return fmt.Errorf("option Sequence: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func TTL(i int32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option TTL: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.PingResponse:
			msg.Ttl = i
		case *gnoisystem.TracerouteRequest:
			msg.MaxTtl = i
		default:
			return fmt.Errorf("option TTL: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func InitialTTL(i uint32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option InitialTTL: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.TracerouteRequest:
			msg.InitialTtl = i
		default:
			return fmt.Errorf("option InitialTTL: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Hops(i int32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Hops: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.TracerouteResponse:
			msg.Hops = i
		default:
			return fmt.Errorf("option Hops: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Hop(i int32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Hop: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.TracerouteResponse:
			msg.Hop = i
		default:
			return fmt.Errorf("option Hop: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Address(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Address: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.TracerouteResponse:
			msg.Address = s
		default:
			return fmt.Errorf("option Address: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Name(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Name: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.TracerouteResponse:
			msg.Name = s
		default:
			return fmt.Errorf("option Name: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func RTT(s int64) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option RTT: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.TracerouteResponse:
			msg.Rtt = s
		default:
			return fmt.Errorf("option RTT: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func State(s string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option State: %w", api.ErrInvalidMsgType)
		}
		st, ok := gnoisystem.TracerouteResponse_State_value[strings.ToUpper(s)]
		if !ok {
			return api.ErrInvalidValue
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.TracerouteResponse:
			msg.State = gnoisystem.TracerouteResponse_State(st)
		default:
			return fmt.Errorf("option State: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func StateDEFAULT() func(msg proto.Message) error {
	return State("DEFAULT")
}

func StateNONE() func(msg proto.Message) error {
	return State("NONE")
}

func StateUNKNOWN() func(msg proto.Message) error {
	return State("UNKNOWN")
}

func StateICMP() func(msg proto.Message) error {
	return State("ICMP")
}

func StateHOST_UNREACHABLE() func(msg proto.Message) error {
	return State("HOST_UNREACHABLE")
}

func StateNETWORK_UNREACHABLE() func(msg proto.Message) error {
	return State("NETWORK_UNREACHABLE")
}

func StatePROTOCOL_UNREACHABLE() func(msg proto.Message) error {
	return State("PROTOCOL_UNREACHABLE")
}

func StateSOURCE_ROUTE_FAILED() func(msg proto.Message) error {
	return State("SOURCE_ROUTE_FAILED")
}

func StateFRAGMENTATION_NEEDED() func(msg proto.Message) error {
	return State("FRAGMENTATION_NEEDED")
}

func StatePROHIBITED() func(msg proto.Message) error {
	return State("PROHIBITED")
}

func StatePRECEDENCE_VIOLATION() func(msg proto.Message) error {
	return State("PRECEDENCE_VIOLATION")
}

func StatePRECEDENCE_CUTOFF() func(msg proto.Message) error {
	return State("PRECEDENCE_CUTOFF")
}

func ICMPCode(i int32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option ICMPCode: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.TracerouteResponse:
			msg.IcmpCode = i
		default:
			return fmt.Errorf("option ICMPCode: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func MPLS(m map[string]string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option MPLS: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.TracerouteResponse:
			if msg.Mpls == nil {
				msg.Mpls = make(map[string]string)
			}
			for k, v := range m {
				msg.Mpls[k] = v
			}
		default:
			return fmt.Errorf("option MPLS: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func MPLSKV(v ...string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option MPLSKV: %w", api.ErrInvalidMsgType)
		}
		if len(v)%2 != 0 {
			return fmt.Errorf("missing key or value: %w", api.ErrInvalidValue)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.TracerouteResponse:
			if msg.Mpls == nil {
				msg.Mpls = make(map[string]string)
			}
			for i := 0; i < len(v); i += 2 {
				msg.Mpls[v[i]] = v[i+1]
			}
		default:
			return fmt.Errorf("option MPLSKV: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func ASPath(i ...int32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option ASPath: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.TracerouteResponse:
			if msg.AsPath == nil {
				msg.AsPath = make([]int32, 0, len(i))
			}
			msg.AsPath = append(msg.AsPath, i...)
		default:
			return fmt.Errorf("option ASPath: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func L4Protocol(p string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option L4Protocol: %w", api.ErrInvalidMsgType)
		}
		l4p, ok := gnoisystem.TracerouteRequest_L4Protocol_value[strings.ToUpper(p)]
		if !ok {
			return fmt.Errorf("option L4Protocol: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.TracerouteRequest:
			msg.L4Protocol = gnoisystem.TracerouteRequest_L4Protocol(l4p)
		default:
			return fmt.Errorf("option L4Protocol: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func L4ProtocolICMP() func(msg proto.Message) error {
	return L4Protocol("ICMP")
}

func L4ProtocolTCP() func(msg proto.Message) error {
	return L4Protocol("TCP")
}

func L4ProtocolUDP() func(msg proto.Message) error {
	return L4Protocol("UDP")
}

func L4ProtocolCustom(i int32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option L4ProtocolCustom: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.TracerouteRequest:
			msg.L4Protocol = gnoisystem.TracerouteRequest_L4Protocol(i)
		default:
			return fmt.Errorf("option L4ProtocolCustom: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func DoNotLookupAsn(b bool) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option DoNotLookupAsn: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.TracerouteRequest:
			msg.DoNotLookupAsn = b
		default:
			return fmt.Errorf("option DoNotLookupAsn: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func PID(p uint32) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option PID: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.KillProcessRequest:
			msg.Pid = p
		default:
			return fmt.Errorf("option PID: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func ProcessName(n string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option ProcessName: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.KillProcessRequest:
			msg.Name = n
		default:
			return fmt.Errorf("option ProcessName: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func Signal(sig string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Signal: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.KillProcessRequest:
			msg.Signal = gnoisystem.KillProcessRequest_Signal(gnoisystem.KillProcessRequest_Signal_value["SIGNAL_"+sig])
		default:
			return fmt.Errorf("option Signal: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func ProcessRestart(b bool) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option ProcessRestart: %w", api.ErrInvalidMsgType)
		}
		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.KillProcessRequest:
			msg.Restart = b
		default:
			return fmt.Errorf("option ProcessRestart: %w", api.ErrInvalidMsgType)
		}
		return nil
	}
}

func PackageFile(n string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option PackageFile: %w", api.ErrInvalidMsgType)
		}

		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.SetPackageRequest:
			switch m := msg.GetRequest().(type) {
			case *gnoisystem.SetPackageRequest_Package:
				m.Package.Filename = n
			default:
				return api.ErrInvalidMsgType
			}
		default:
			return fmt.Errorf("option PackageFile: %w", api.ErrInvalidMsgType)
		}

		return nil
	}
}

func Version(v string) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Version: %w", api.ErrInvalidMsgType)
		}

		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.SetPackageRequest:
			switch m := msg.GetRequest().(type) {
			case *gnoisystem.SetPackageRequest_Package:
				m.Package.Version = v
			default:
				return api.ErrInvalidMsgType
			}
		default:
			return fmt.Errorf("option Version: %w", api.ErrInvalidMsgType)
		}

		return nil
	}
}

func Activate(b bool) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Activate: %w", api.ErrInvalidMsgType)
		}

		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.SetPackageRequest:
			switch m := msg.GetRequest().(type) {
			case *gnoisystem.SetPackageRequest_Package:
				m.Package.Activate = b
			default:
				return api.ErrInvalidMsgType
			}
		default:
			return fmt.Errorf("option Version: %w", api.ErrInvalidMsgType)
		}

		return nil
	}
}

func Hash(method string, b []byte) func(msg proto.Message) error {
	return func(msg proto.Message) error {
		if msg == nil {
			return fmt.Errorf("option Hash: %w", api.ErrInvalidMsgType)
		}

		ht, ok := types.HashType_HashMethod_value[strings.ToUpper(method)]
		if !ok {
			return api.ErrInvalidValue
		}

		switch msg := msg.ProtoReflect().Interface().(type) {
		case *gnoisystem.SetPackageRequest:
			switch m := msg.GetRequest().(type) {
			case *gnoisystem.SetPackageRequest_Hash:
				m.Hash = &types.HashType{
					Method: types.HashType_HashMethod(ht),
					Hash:   b,
				}
			default:
				return fmt.Errorf("option Hash: %w", api.ErrInvalidMsgType)
			}

		default:
			return fmt.Errorf("option Hash: %w", api.ErrInvalidMsgType)
		}

		return nil
	}
}
