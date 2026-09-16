package app

import (
	"strings"
	"testing"
	"time"

	"github.com/openconfig/gnoi/healthz"
	"github.com/openconfig/gnoi/types"
	"google.golang.org/protobuf/types/known/anypb"
	"google.golang.org/protobuf/types/known/timestamppb"
)

func TestArtifactType(t *testing.T) {
	tests := []struct {
		hdr  *healthz.ArtifactHeader
		want string
	}{
		{&healthz.ArtifactHeader{ArtifactType: &healthz.ArtifactHeader_File{}}, "file"},
		{&healthz.ArtifactHeader{ArtifactType: &healthz.ArtifactHeader_Custom{}}, "custom"},
		{&healthz.ArtifactHeader{ArtifactType: &healthz.ArtifactHeader_Proto{}}, "proto"},
		{&healthz.ArtifactHeader{}, ""},
	}
	for _, tt := range tests {
		if got := artifactType(tt.hdr); got != tt.want {
			t.Errorf("artifactType(%v) = %q, want %q", tt.hdr, got, tt.want)
		}
	}
}

func TestPrintArtifactType(t *testing.T) {
	tests := []struct {
		name string
		hdr  *healthz.ArtifactHeader
		want []string
	}{
		{
			name: "file",
			hdr: &healthz.ArtifactHeader{
				Id: "a1",
				ArtifactType: &healthz.ArtifactHeader_File{File: &healthz.FileArtifactType{
					Name: "core.gz", Path: "/var/core", Mimetype: "application/gzip", Size: 42,
					Hash: &types.HashType{Method: types.HashType_MD5, Hash: []byte{0xab}},
				}},
			},
			want: []string{"id       : a1", "name     : core.gz", "path     : /var/core", "mimeType : application/gzip", "size     : 42", "MD5(ab)"},
		},
		{
			name: "custom",
			hdr: &healthz.ArtifactHeader{
				Id:           "a2",
				ArtifactType: &healthz.ArtifactHeader_Custom{Custom: &anypb.Any{TypeUrl: "type.googleapis.com/x", Value: []byte{1, 2}}},
			},
			want: []string{"id       : a2", "typeURL : type.googleapis.com/x", "value   : 0102"},
		},
		{
			name: "proto",
			hdr: &healthz.ArtifactHeader{
				Id:           "a3",
				ArtifactType: &healthz.ArtifactHeader_Proto{Proto: &healthz.ProtoArtifactType{}},
			},
			want: []string{"id       : a3", "proto :"},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			out := string(printArtifactType("", tt.hdr))
			for _, w := range tt.want {
				if !strings.Contains(out, w) {
					t.Errorf("output missing %q:\n%s", w, out)
				}
			}
		})
	}
}

func TestHealthzGetTree(t *testing.T) {
	a := newTestApp(t)
	created := time.Unix(1_700_000_000, 0).UTC()
	comp := &healthz.ComponentStatus{
		Path:         &types.Path{Elem: []*types.PathElem{{Name: "components"}, {Name: "component", Key: map[string]string{"name": "chassis"}}}},
		Status:       healthz.Status_STATUS_UNHEALTHY,
		Id:           "ev-1",
		Acknowledged: true,
		Created:      timestamppb.New(created),
		Expires:      timestamppb.New(created.Add(time.Hour)),
		Artifacts: []*healthz.ArtifactHeader{
			{Id: "art-1", ArtifactType: &healthz.ArtifactHeader_File{File: &healthz.FileArtifactType{Name: "f"}}},
		},
		Subcomponents: []*healthz.ComponentStatus{
			{
				Path:   &types.Path{Elem: []*types.PathElem{{Name: "components"}, {Name: "component", Key: map[string]string{"name": "cpu0"}}}},
				Status: healthz.Status_STATUS_HEALTHY,
				Id:     "ev-2",
			},
		},
	}
	out := a.healthzGetTree(comp, "  ")
	for _, want := range []string{
		"  path     : components/component[name=chassis]",
		"  status   : STATUS_UNHEALTHY",
		"  id       : ev-1",
		"  acked    : true",
		"  created  : 2023-11-14 22:13:20 +0000 UTC",
		"  artifict :",
		"    - id       : art-1",
		"  subcomponents:",
		"    path     : components/component[name=cpu0]",
		"    status   : STATUS_HEALTHY",
		"    id       : ev-2",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("tree output missing %q:\n%s", want, out)
		}
	}
	// no artifacts / subcomponents sections when empty
	leaf := a.healthzGetTree(&healthz.ComponentStatus{Id: "x"}, "")
	if strings.Contains(leaf, "artifict") || strings.Contains(leaf, "subcomponents") {
		t.Errorf("unexpected sections in leaf output:\n%s", leaf)
	}
}
