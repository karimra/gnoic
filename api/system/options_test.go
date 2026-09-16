package system

import (
	"errors"
	"testing"
	"time"

	gnoisystem "github.com/openconfig/gnoi/system"
	"github.com/openconfig/gnoi/types"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/karimra/gnoic/api"
)

func TestOptions_InvalidMessage(t *testing.T) {
	opts := map[string]SystemOption{
		"Destination":      Destination("d"),
		"DestinationName":  DestinationName("d"),
		"Source":           Source("s"),
		"Count":            Count(1),
		"Interval":         Interval(1),
		"Wait":             Wait(1),
		"Size":             Size(1),
		"DoNotFragment":    DoNotFragment(true),
		"DoNotResolve":     DoNotResolve(true),
		"NetworkInstance":  NetworkInstance("ns"),
		"L3Protocol":       L3Protocol("IPV4"),
		"L3ProtocolIPv4":   L3ProtocolIPv4(),
		"L3ProtocolIPv6":   L3ProtocolIPv6(),
		"L3ProtocolUNSPEC": L3ProtocolUNSPECIFIED(),
		"L3ProtocolCustom": L3ProtocolCustom(1),
		"Time":             Time(1),
		"CurrentTime":      CurrentTime(1),
		"CurrentTimeNow":   CurrentTimeNow(),
		"Sent":             Sent(1),
		"Received":         Received(1),
		"MinTime":          MinTime(1),
		"AvgTime":          AvgTime(1),
		"MaxTime":          MaxTime(1),
		"StdDev":           StdDev(1),
		"Bytes":            Bytes(1),
		"Sequence":         Sequence(1),
		"TTL":              TTL(1),
		"InitialTTL":       InitialTTL(1),
		"Hops":             Hops(1),
		"Hop":              Hop(1),
		"Address":          Address("a"),
		"Name":             Name("n"),
		"RTT":              RTT(1),
		"State":            State("ICMP"),
		"StateDEFAULT":     StateDEFAULT(),
		"ICMPCode":         ICMPCode(1),
		"MPLS":             MPLS(map[string]string{"a": "b"}),
		"MPLSKV":           MPLSKV("a", "b"),
		"ASPath":           ASPath(1),
		"L4Protocol":       L4Protocol("TCP"),
		"L4ProtocolICMP":   L4ProtocolICMP(),
		"L4ProtocolTCP":    L4ProtocolTCP(),
		"L4ProtocolUDP":    L4ProtocolUDP(),
		"L4ProtocolCustom": L4ProtocolCustom(1),
		"DoNotLookupAsn":   DoNotLookupAsn(true),
		"PID":              PID(1),
		"ProcessName":      ProcessName("p"),
		"Signal":           Signal("TERM"),
		"ProcessRestart":   ProcessRestart(true),
		"PackageFile":      PackageFile("f"),
		"Version":          Version("v"),
		"Activate":         Activate(true),
		"Hash":             Hash("MD5", nil),
	}
	for name, o := range opts {
		t.Run(name, func(t *testing.T) {
			if err := o(nil); !errors.Is(err, api.ErrInvalidMsgType) {
				t.Errorf("nil message: got %v, want ErrInvalidMsgType", err)
			}
			if err := o(&emptypb.Empty{}); !errors.Is(err, api.ErrInvalidMsgType) {
				t.Errorf("wrong message type: got %v, want ErrInvalidMsgType", err)
			}
		})
	}
}

func TestNewSystemPingRequest(t *testing.T) {
	req, err := NewSystemPingRequest(
		Destination("1.1.1.1"),
		Source("2.2.2.2"),
		Count(4),
		Interval(int64(time.Second)),
		Wait(int64(2*time.Second)),
		Size(64),
		DoNotFragment(true),
		DoNotResolve(true),
		L3ProtocolIPv4(),
		NetworkInstance("mgmt"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if req.GetDestination() != "1.1.1.1" || req.GetSource() != "2.2.2.2" || req.GetCount() != 4 ||
		req.GetInterval() != int64(time.Second) || req.GetWait() != int64(2*time.Second) || req.GetSize() != 64 ||
		!req.GetDoNotFragment() || !req.GetDoNotResolve() || req.GetL3Protocol() != types.L3Protocol_IPV4 ||
		req.GetNetworkInstance() != "mgmt" {
		t.Errorf("unexpected request: %v", req)
	}
	// L3 protocol variants
	for _, tt := range []struct {
		opt  SystemOption
		want types.L3Protocol
	}{
		{L3ProtocolIPv6(), types.L3Protocol_IPV6},
		{L3ProtocolUNSPECIFIED(), types.L3Protocol_UNSPECIFIED},
		{L3Protocol("ipv4"), types.L3Protocol_IPV4},
		{L3ProtocolCustom(7), types.L3Protocol(7)},
	} {
		r, err := NewSystemPingRequest(tt.opt)
		if err != nil {
			t.Fatal(err)
		}
		if r.GetL3Protocol() != tt.want {
			t.Errorf("L3Protocol = %v, want %v", r.GetL3Protocol(), tt.want)
		}
	}
	if _, err := NewSystemPingRequest(L3Protocol("IPX")); err == nil {
		t.Error("expected error for unknown L3 protocol")
	}
	if _, err := NewSystemPingRequest(Hops(1)); err == nil {
		t.Error("expected error for inapplicable option")
	}
}

func TestNewSystemPingResponse(t *testing.T) {
	rsp, err := NewSystemPingResponse(
		Source("1.1.1.1"), Time(10), Sent(5), Received(4), MinTime(1), AvgTime(2), MaxTime(3), StdDev(1), Bytes(64), Sequence(2), TTL(60),
	)
	if err != nil {
		t.Fatal(err)
	}
	if rsp.GetSource() != "1.1.1.1" || rsp.GetTime() != 10 || rsp.GetSent() != 5 || rsp.GetReceived() != 4 ||
		rsp.GetMinTime() != 1 || rsp.GetAvgTime() != 2 || rsp.GetMaxTime() != 3 || rsp.GetStdDev() != 1 ||
		rsp.GetBytes() != 64 || rsp.GetSequence() != 2 || rsp.GetTtl() != 60 {
		t.Errorf("unexpected response: %v", rsp)
	}
}

func TestNewSystemTracerouteRequest(t *testing.T) {
	req, err := NewSystemTracerouteRequest(
		Destination("1.1.1.1"), Source("2.2.2.2"), InitialTTL(2), TTL(30), Wait(1), DoNotFragment(true), DoNotResolve(true),
		L3ProtocolIPv6(), L4ProtocolUDP(), DoNotLookupAsn(true), NetworkInstance("ns"),
	)
	if err != nil {
		t.Fatal(err)
	}
	if req.GetDestination() != "1.1.1.1" || req.GetSource() != "2.2.2.2" || req.GetInitialTtl() != 2 || req.GetMaxTtl() != 30 ||
		req.GetWait() != 1 || !req.GetDoNotFragment() || !req.GetDoNotResolve() || req.GetL3Protocol() != types.L3Protocol_IPV6 ||
		req.GetL4Protocol() != gnoisystem.TracerouteRequest_UDP || !req.GetDoNotLookupAsn() || req.GetNetworkInstance() != "ns" {
		t.Errorf("unexpected request: %v", req)
	}
	for _, tt := range []struct {
		opt  SystemOption
		want gnoisystem.TracerouteRequest_L4Protocol
	}{
		{L4ProtocolICMP(), gnoisystem.TracerouteRequest_ICMP},
		{L4ProtocolTCP(), gnoisystem.TracerouteRequest_TCP},
		{L4Protocol("udp"), gnoisystem.TracerouteRequest_UDP},
		{L4ProtocolCustom(9), gnoisystem.TracerouteRequest_L4Protocol(9)},
	} {
		r, err := NewSystemTracerouteRequest(tt.opt)
		if err != nil {
			t.Fatal(err)
		}
		if r.GetL4Protocol() != tt.want {
			t.Errorf("L4Protocol = %v, want %v", r.GetL4Protocol(), tt.want)
		}
	}
	if _, err := NewSystemTracerouteRequest(L4Protocol("SCTP")); err == nil {
		t.Error("expected error for unknown L4 protocol")
	}
}

func TestNewSystemTracerouteResponse(t *testing.T) {
	rsp, err := NewSystemTracerouteResponse(
		Destination("1.1.1.1"), DestinationName("one"), Hops(3), Size(64), Hop(1), Address("10.0.0.1"), Name("r1"),
		RTT(100), StateICMP(), ICMPCode(3), MPLS(map[string]string{"label": "100"}), MPLSKV("exp", "0", "ttl", "64"), ASPath(65000, 65001),
	)
	if err != nil {
		t.Fatal(err)
	}
	if rsp.GetDestinationAddress() != "1.1.1.1" || rsp.GetDestinationName() != "one" || rsp.GetHops() != 3 || rsp.GetPacketSize() != 64 ||
		rsp.GetHop() != 1 || rsp.GetAddress() != "10.0.0.1" || rsp.GetName() != "r1" || rsp.GetRtt() != 100 ||
		rsp.GetState() != gnoisystem.TracerouteResponse_ICMP || rsp.GetIcmpCode() != 3 {
		t.Errorf("unexpected response: %v", rsp)
	}
	if len(rsp.GetMpls()) != 3 || rsp.GetMpls()["label"] != "100" || rsp.GetMpls()["exp"] != "0" || rsp.GetMpls()["ttl"] != "64" {
		t.Errorf("Mpls = %v", rsp.GetMpls())
	}
	if len(rsp.GetAsPath()) != 2 || rsp.GetAsPath()[0] != 65000 || rsp.GetAsPath()[1] != 65001 {
		t.Errorf("AsPath = %v", rsp.GetAsPath())
	}
	if err := MPLSKV("odd")(rsp); !errors.Is(err, api.ErrInvalidValue) {
		t.Errorf("MPLSKV with odd number of args: %v", err)
	}
	if err := State("BOGUS")(rsp); !errors.Is(err, api.ErrInvalidValue) {
		t.Errorf("State with unknown value: %v", err)
	}
	for _, tt := range []struct {
		opt  SystemOption
		want gnoisystem.TracerouteResponse_State
	}{
		{StateDEFAULT(), gnoisystem.TracerouteResponse_DEFAULT},
		{StateNONE(), gnoisystem.TracerouteResponse_NONE},
		{StateUNKNOWN(), gnoisystem.TracerouteResponse_UNKNOWN},
		{StateHOST_UNREACHABLE(), gnoisystem.TracerouteResponse_HOST_UNREACHABLE},
		{StateNETWORK_UNREACHABLE(), gnoisystem.TracerouteResponse_NETWORK_UNREACHABLE},
		{StatePROTOCOL_UNREACHABLE(), gnoisystem.TracerouteResponse_PROTOCOL_UNREACHABLE},
		{StateSOURCE_ROUTE_FAILED(), gnoisystem.TracerouteResponse_SOURCE_ROUTE_FAILED},
		{StateFRAGMENTATION_NEEDED(), gnoisystem.TracerouteResponse_FRAGMENTATION_NEEDED},
		{StatePROHIBITED(), gnoisystem.TracerouteResponse_PROHIBITED},
		{StatePRECEDENCE_VIOLATION(), gnoisystem.TracerouteResponse_PRECEDENCE_VIOLATION},
		{StatePRECEDENCE_CUTOFF(), gnoisystem.TracerouteResponse_PRECEDENCE_CUTOFF},
	} {
		r, err := NewSystemTracerouteResponse(tt.opt)
		if err != nil {
			t.Fatal(err)
		}
		if r.GetState() != tt.want {
			t.Errorf("State = %v, want %v", r.GetState(), tt.want)
		}
	}
}

func TestNewSystemTimeResponse(t *testing.T) {
	if NewSystemTimeRequest() == nil {
		t.Fatal("nil time request")
	}
	rsp, err := NewSystemTimeResponse(CurrentTime(42))
	if err != nil {
		t.Fatal(err)
	}
	if rsp.GetTime() != 42 {
		t.Errorf("Time = %d", rsp.GetTime())
	}
	if _, err := NewSystemTimeResponse(CurrentTimeNow()); err != nil {
		t.Errorf("CurrentTimeNow: %v", err)
	}
}

func TestNewSystemKillProcessRequest(t *testing.T) {
	req, err := NewSystemKillProcessRequest(PID(12), ProcessName("sshd"), Signal("HUP"), ProcessRestart(true))
	if err != nil {
		t.Fatal(err)
	}
	if req.GetPid() != 12 || req.GetName() != "sshd" || req.GetSignal() != gnoisystem.KillProcessRequest_SIGNAL_HUP || !req.GetRestart() {
		t.Errorf("unexpected request: %v", req)
	}
	for sig, want := range map[string]gnoisystem.KillProcessRequest_Signal{
		"TERM": gnoisystem.KillProcessRequest_SIGNAL_TERM,
		"KILL": gnoisystem.KillProcessRequest_SIGNAL_KILL,
		"ABRT": gnoisystem.KillProcessRequest_SIGNAL_ABRT,
	} {
		r, _ := NewSystemKillProcessRequest(Signal(sig))
		if r.GetSignal() != want {
			t.Errorf("Signal(%s) = %v, want %v", sig, r.GetSignal(), want)
		}
	}
}

func TestNewSetPackageRequests(t *testing.T) {
	pkg, err := NewSetPackagePackageRequest(PackageFile("/flash/img.bin"), Version("1.2.3"), Activate(true))
	if err != nil {
		t.Fatal(err)
	}
	if pkg.GetPackage().GetFilename() != "/flash/img.bin" || pkg.GetPackage().GetVersion() != "1.2.3" || !pkg.GetPackage().GetActivate() {
		t.Errorf("Package = %v", pkg.GetPackage())
	}
	h, err := NewSetPackageHashRequest(Hash("sha512", []byte{9}))
	if err != nil {
		t.Fatal(err)
	}
	if h.GetHash().GetMethod() != types.HashType_SHA512 || len(h.GetHash().GetHash()) != 1 {
		t.Errorf("Hash = %v", h.GetHash())
	}
	// options applied to the wrong oneof variant fail
	if _, err := NewSetPackageHashRequest(PackageFile("x")); !errors.Is(err, api.ErrInvalidMsgType) {
		t.Errorf("PackageFile on hash request: %v", err)
	}
	if _, err := NewSetPackageHashRequest(Version("x")); !errors.Is(err, api.ErrInvalidMsgType) {
		t.Errorf("Version on hash request: %v", err)
	}
	if _, err := NewSetPackageHashRequest(Activate(true)); !errors.Is(err, api.ErrInvalidMsgType) {
		t.Errorf("Activate on hash request: %v", err)
	}
	if _, err := NewSetPackagePackageRequest(Hash("MD5", nil)); !errors.Is(err, api.ErrInvalidMsgType) {
		t.Errorf("Hash on package request: %v", err)
	}
	if _, err := NewSetPackageHashRequest(Hash("nope", nil)); !errors.Is(err, api.ErrInvalidValue) {
		t.Errorf("unknown hash method: %v", err)
	}
}
