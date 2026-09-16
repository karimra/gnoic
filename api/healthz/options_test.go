package healthz

import (
	"errors"
	"testing"
	"time"

	gnoihealthz "github.com/openconfig/gnoi/healthz"
	"github.com/openconfig/gnoi/types"
	"google.golang.org/protobuf/types/known/emptypb"

	"github.com/karimra/gnoic/api"
)

func TestOptions_InvalidMessage(t *testing.T) {
	opts := map[string]HealthzOption{
		"Path":                Path("/a"),
		"ComponentStatus":     ComponentStatus(),
		"Status":              Status("STATUS_HEALTHY"),
		"ArtifactHeader":      ArtifactHeader(),
		"ID":                  ID("id"),
		"Name":                Name("n"),
		"SysPath":             SysPath("/p"),
		"MimeType":            MimeType("text/plain"),
		"Size":                Size(1),
		"Hash":                Hash("MD5", nil),
		"Acknowledged":        Acknowledged(true),
		"Created":             Created(time.Now()),
		"Expires":             Expires(time.Now()),
		"IncludeAcknowledged": IncludeAcknowledged(true),
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

func TestOptions_InvalidValue(t *testing.T) {
	if err := Status("SICK")(&gnoihealthz.ComponentStatus{}); !errors.Is(err, api.ErrInvalidValue) {
		t.Errorf("unknown status: %v", err)
	}
	// NOTE: Status_HEALTHY()/Status_UNHEALTHY()/Status_UNSPECIFIED() currently
	// look up names without the STATUS_ prefix and therefore always fail; this
	// is tracked as a bug and covered here only to document current behaviour.
	if err := Status_HEALTHY()(&gnoihealthz.ComponentStatus{}); !errors.Is(err, api.ErrInvalidValue) {
		t.Errorf("Status_HEALTHY: %v", err)
	}
	if err := Hash("CRC", nil)(&gnoihealthz.FileArtifactType{}); !errors.Is(err, api.ErrInvalidValue) {
		t.Errorf("unknown hash method: %v", err)
	}
	// malformed xpath surfaces from Path()
	if _, err := NewGetRequest(Path("a[k")); err == nil {
		t.Error("expected error for malformed path")
	}
}

func TestRequests(t *testing.T) {
	get, err := NewGetRequest(Path("/components/component[name=cpu0]"))
	if err != nil {
		t.Fatal(err)
	}
	if len(get.GetPath().GetElem()) != 2 || get.GetPath().GetElem()[1].GetKey()["name"] != "cpu0" {
		t.Errorf("GetRequest path = %v", get.GetPath())
	}
	ack, err := NewAcknowledgeRequest(Path("/a"), ID("ev-1"))
	if err != nil {
		t.Fatal(err)
	}
	if ack.GetId() != "ev-1" || len(ack.GetPath().GetElem()) != 1 {
		t.Errorf("AcknowledgeRequest = %v", ack)
	}
	list, err := NewListRequest(Path("/a"), IncludeAcknowledged(true))
	if err != nil {
		t.Fatal(err)
	}
	if !list.GetIncludeAcknowledged() || len(list.GetPath().GetElem()) != 1 {
		t.Errorf("ListRequest = %v", list)
	}
	check, err := NewCheckRequest(Path("/a"), ID("ev-2"))
	if err != nil {
		t.Fatal(err)
	}
	if check.GetEventId() != "ev-2" {
		t.Errorf("CheckRequest = %v", check)
	}
	art, err := NewArtifactRequest(ID("art-1"))
	if err != nil {
		t.Fatal(err)
	}
	if art.GetId() != "art-1" {
		t.Errorf("ArtifactRequest = %v", art)
	}
	if _, err := NewArtifactRequest(Path("/a")); err == nil {
		t.Error("expected error: ArtifactRequest has no path")
	}
}

func TestResponses(t *testing.T) {
	created := time.Unix(100, 0)
	expires := time.Unix(200, 0)
	rsp, err := NewGetResponse(
		ComponentStatus(
			ID("c-1"), Status("STATUS_UNHEALTHY"), Acknowledged(true), Created(created), Expires(expires),
			ArtifactHeader(ID("art-1")),
			ComponentStatus(ID("c-2"), Status("status_healthy")),
		),
	)
	if err != nil {
		t.Fatal(err)
	}
	c := rsp.GetComponent()
	if c.GetId() != "c-1" || c.GetStatus() != gnoihealthz.Status_STATUS_UNHEALTHY || !c.GetAcknowledged() ||
		!c.GetCreated().AsTime().Equal(created) || !c.GetExpires().AsTime().Equal(expires) {
		t.Errorf("ComponentStatus = %v", c)
	}
	if len(c.GetArtifacts()) != 1 || c.GetArtifacts()[0].GetId() != "art-1" {
		t.Errorf("Artifacts = %v", c.GetArtifacts())
	}
	if len(c.GetSubcomponents()) != 1 || c.GetSubcomponents()[0].GetStatus() != gnoihealthz.Status_STATUS_HEALTHY {
		t.Errorf("Subcomponents = %v", c.GetSubcomponents())
	}
	// nested errors propagate
	if _, err := NewGetResponse(ComponentStatus(Name("x"))); !errors.Is(err, api.ErrInvalidMsgType) {
		t.Errorf("nested invalid option: %v", err)
	}

	list, err := NewListResponse(ComponentStatus(ID("a")), ComponentStatus(ID("b")))
	if err != nil {
		t.Fatal(err)
	}
	if len(list.GetStatuses()) != 2 || list.GetStatuses()[1].GetId() != "b" {
		t.Errorf("ListResponse = %v", list)
	}
	ack, err := NewAcknowledgeResponse(ComponentStatus(ID("a")))
	if err != nil || ack.GetStatus().GetId() != "a" {
		t.Errorf("AcknowledgeResponse = %v, %v", ack, err)
	}
	check, err := NewCheckResponse(ComponentStatus(ID("a")))
	if err != nil || check.GetStatus().GetId() != "a" {
		t.Errorf("CheckResponse = %v, %v", check, err)
	}
	art, err := NewArtifactResponse(ArtifactHeader(ID("art-1")))
	if err != nil || art.GetHeader().GetId() != "art-1" {
		t.Errorf("ArtifactResponse = %v, %v", art, err)
	}
	if _, err := NewArtifactResponse(ArtifactHeader(Name("x"))); !errors.Is(err, api.ErrInvalidMsgType) {
		t.Errorf("nested invalid artifact header option: %v", err)
	}
}

func TestFileArtifactType(t *testing.T) {
	f := new(gnoihealthz.FileArtifactType)
	if err := apply(f, Name("core.gz"), SysPath("/var/core"), MimeType("application/gzip"), Size(1024), Hash("sha256", []byte{1})); err != nil {
		t.Fatal(err)
	}
	if f.GetName() != "core.gz" || f.GetPath() != "/var/core" || f.GetMimetype() != "application/gzip" || f.GetSize() != 1024 ||
		f.GetHash().GetMethod() != types.HashType_SHA256 {
		t.Errorf("FileArtifactType = %v", f)
	}
}
