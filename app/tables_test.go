package app

import (
	"strings"
	"testing"
	"time"

	"github.com/openconfig/gnoi/file"
	"github.com/openconfig/gnoi/system"
)

func TestSystemTimeTable(t *testing.T) {
	ts := uint64(time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC).UnixNano())
	out, err := systemTimeTable([]*systemTimeResponse{
		{TargetError: TargetError{TargetName: "r2"}, rsp: &system.TimeResponse{Time: ts}},
		{TargetError: TargetError{TargetName: "r1"}, rsp: &system.TimeResponse{Time: ts}},
	})
	if err != nil {
		t.Fatal(err)
	}
	for _, want := range []string{"Target Name", "Time", "Timestamp", "r1", "r2", "1704164645000000000"} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q:\n%s", want, out)
		}
	}
	// sorted by target name
	if strings.Index(out, "r1") > strings.Index(out, "r2") {
		t.Errorf("rows not sorted by target name:\n%s", out)
	}
}

func TestStatTable(t *testing.T) {
	a := newTestApp(t)
	mod := uint64(time.Date(2024, 1, 2, 3, 4, 5, 0, time.UTC).UnixNano())
	rsps := []*fileStatResponse{
		{
			TargetError: TargetError{TargetName: "r2"},
			rsp: []*fileStatInfo{
				{StatInfo: &file.StatInfo{Path: "/cf3/b.txt", LastModified: mod, Permissions: 644, Umask: 22, Size: 2048}},
				{StatInfo: &file.StatInfo{Path: "/cf3/a", LastModified: mod, Permissions: 755, Umask: 22, Size: 4096}, IsDir: true},
			},
		},
		{
			TargetError: TargetError{TargetName: "r1"},
			rsp: []*fileStatInfo{
				{StatInfo: &file.StatInfo{Path: "/cf3/c.bin", LastModified: mod, Permissions: 600, Umask: 22, Size: 1}},
			},
		},
	}

	out := a.statTable(rsps)
	for _, want := range []string{
		"Target Name", "Path", "LastModified", "Perm", "Umask", "Size",
		"-rw-r--r--", "drwxr-xr-x", "-rw-------", "----w--w-",
		"2024-01-02T03:04:05Z", "2048", "4096",
	} {
		if !strings.Contains(out, want) {
			t.Errorf("table missing %q:\n%s", want, out)
		}
	}
	// targets sorted, and files sorted within a target
	if strings.Index(out, "/cf3/c.bin") > strings.Index(out, "/cf3/b.txt") {
		t.Errorf("targets not sorted (r1 before r2):\n%s", out)
	}
	if strings.Index(out, "/cf3/a") > strings.Index(out, "/cf3/b.txt") {
		t.Errorf("paths not sorted within target:\n%s", out)
	}

	a.Config.FileStatHumanize = true
	out = a.statTable(rsps)
	for _, want := range []string{"2.0 kB", "4.1 kB", "ago"} {
		if !strings.Contains(out, want) {
			t.Errorf("humanized table missing %q:\n%s", want, out)
		}
	}
}
