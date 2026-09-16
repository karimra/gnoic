package os

import (
	"errors"
	"testing"

	gnoios "github.com/openconfig/gnoi/os"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/karimra/gnoic/api"
)

func TestOptions_InvalidMessage(t *testing.T) {
	opts := map[string]OsOption{
		"Version":               Version("v"),
		"Description":           Description("d"),
		"StandbySupervisor":     StandbySupervisor(true),
		"PackageSize":           PackageSize(1),
		"NoReboot":              NoReboot(true),
		"ErrorType":             ErrorType(1),
		"ErrorDetail":           ErrorDetail("d"),
		"BytesReceived":         BytesReceived(1),
		"PercentageTransferred": PercentageTransferred(1),
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
	// verify-related options only reject nil messages
	for name, o := range map[string]OsOption{
		"ActivationFailMsg":     ActivationFailMsg("m"),
		"VerifyStandbyState":    VerifyStandbyStateUNAVAILABLE(),
		"VerifyStandbyResponse": VerifyStandbyResponse(),
		"StandbyResponseID":     StandbyResponseID("id"),
	} {
		t.Run(name, func(t *testing.T) {
			if err := o(nil); !errors.Is(err, api.ErrInvalidMsgType) {
				t.Errorf("nil message: got %v, want ErrInvalidMsgType", err)
			}
		})
	}
}

func TestInstallRequests(t *testing.T) {
	req, err := NewOSInstallTransferRequest(Version("1.0"), StandbySupervisor(true), PackageSize(1234))
	if err != nil {
		t.Fatal(err)
	}
	tr := req.GetTransferRequest()
	if tr.GetVersion() != "1.0" || !tr.GetStandbySupervisor() || tr.GetPackageSize() != 1234 {
		t.Errorf("TransferRequest = %v", tr)
	}
	if _, err := NewOSInstallTransferRequest(NoReboot(true)); err == nil {
		t.Error("expected error for inapplicable option")
	}
	if NewOSInstallTransferEnd().GetTransferEnd() == nil {
		t.Error("TransferEnd missing")
	}
	if _, err := NewOSInstallTransferContent(); err != nil {
		t.Error(err)
	}
}

func TestInstallResponses(t *testing.T) {
	if NewOSInstallTransferReadyResponse().GetTransferReady() == nil {
		t.Error("TransferReady missing")
	}
	p, err := NewOSInstallTransferProgressResponse(BytesReceived(10))
	if err != nil || p.GetTransferProgress().GetBytesReceived() != 10 {
		t.Errorf("TransferProgress = %v, %v", p, err)
	}
	if _, err := NewOSInstallTransferProgressResponse(Version("x")); err == nil {
		t.Error("expected error")
	}
	s, err := NewOSInstallSyncProgressResponse(PercentageTransferred(50))
	if err != nil || s.GetSyncProgress().GetPercentageTransferred() != 50 {
		t.Errorf("SyncProgress = %v, %v", s, err)
	}
	if _, err := NewOSInstallSyncProgressResponse(Version("x")); err == nil {
		t.Error("expected error")
	}
	v, err := NewOSInstallValidatedResponse(Version("1.0"), Description("ok"))
	if err != nil || v.GetValidated().GetVersion() != "1.0" || v.GetValidated().GetDescription() != "ok" {
		t.Errorf("Validated = %v, %v", v, err)
	}
	if _, err := NewOSInstallValidatedResponse(NoReboot(true)); err == nil {
		t.Error("expected error")
	}
	e, err := NewOSInstallInstallErrorResponse(ErrorType(int32(gnoios.InstallError_INCOMPATIBLE)), ErrorDetail("bad"))
	if err != nil || e.GetInstallError().GetType() != gnoios.InstallError_INCOMPATIBLE || e.GetInstallError().GetDetail() != "bad" {
		t.Errorf("InstallError = %v, %v", e, err)
	}
	if _, err := NewOSInstallInstallErrorResponse(Version("x")); err == nil {
		t.Error("expected error")
	}
}

func TestActivate(t *testing.T) {
	req, err := NewActivateRequest(Version("2.0"), StandbySupervisor(true), NoReboot(true))
	if err != nil {
		t.Fatal(err)
	}
	if req.GetVersion() != "2.0" || !req.GetStandbySupervisor() || !req.GetNoReboot() {
		t.Errorf("ActivateRequest = %v", req)
	}
	if _, err := NewActivateRequest(PackageSize(1)); err == nil {
		t.Error("expected error")
	}
	if NewActivateOKResponse().GetActivateOk() == nil {
		t.Error("ActivateOk missing")
	}
	e, err := NewActivateErrorResponse(ErrorType(int32(gnoios.ActivateError_NON_EXISTENT_VERSION)), ErrorDetail("nope"))
	if err != nil || e.GetActivateError().GetType() != gnoios.ActivateError_NON_EXISTENT_VERSION || e.GetActivateError().GetDetail() != "nope" {
		t.Errorf("ActivateError = %v, %v", e, err)
	}
	if _, err := NewActivateErrorResponse(Version("x")); err == nil {
		t.Error("expected error")
	}
}

func TestVerify(t *testing.T) {
	if NewOSVerifyRequest() == nil {
		t.Fatal("nil verify request")
	}
	rsp, err := NewOSVerifyResponse(Version("3.0"), ActivationFailMsg("failed"), VerifyStandbyStateUNSUPPORTED())
	if err != nil {
		t.Fatal(err)
	}
	if rsp.GetVersion() != "3.0" || rsp.GetActivationFailMessage() != "failed" {
		t.Errorf("VerifyResponse = %v", rsp)
	}
	if rsp.GetVerifyStandby().GetStandbyState().GetState() != gnoios.StandbyState_UNSUPPORTED {
		t.Errorf("VerifyStandby = %v", rsp.GetVerifyStandby())
	}
	for _, tt := range []struct {
		opt  OsOption
		want gnoios.StandbyState_State
	}{
		{VerifyStandbyStateNON_EXISTENT(), gnoios.StandbyState_NON_EXISTENT},
		{VerifyStandbyStateUNAVAILABLE(), gnoios.StandbyState_UNAVAILABLE},
	} {
		r, _ := NewOSVerifyResponse(tt.opt)
		if r.GetVerifyStandby().GetStandbyState().GetState() != tt.want {
			t.Errorf("StandbyState = %v, want %v", r.GetVerifyStandby().GetStandbyState().GetState(), tt.want)
		}
	}
	rsp2, err := NewOSVerifyResponse(VerifyStandbyResponse(StandbyResponseID("sup-1"), Version("3.0"), ActivationFailMsg("m")))
	if err != nil {
		t.Fatal(err)
	}
	sr := rsp2.GetVerifyStandby().GetVerifyResponse()
	if sr.GetId() != "sup-1" || sr.GetVersion() != "3.0" || sr.GetActivationFailMessage() != "m" {
		t.Errorf("StandbyResponse = %v", sr)
	}
	if _, err := NewOSVerifyResponse(VerifyStandbyResponse(PackageSize(1))); err == nil {
		t.Error("expected nested error")
	}
}
