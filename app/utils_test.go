package app

import (
	"errors"
	"strings"
	"testing"
	"time"
)

func TestSInListNotEmpty(t *testing.T) {
	tests := []struct {
		s    string
		l    []string
		want bool
	}{
		{"a", nil, true},
		{"a", []string{}, true},
		{"a", []string{"a"}, true},
		{"a", []string{"b", "a"}, true},
		{"a", []string{"b"}, false},
	}
	for _, tt := range tests {
		if got := sInListNotEmpty(tt.s, tt.l); got != tt.want {
			t.Errorf("sInListNotEmpty(%q, %v) = %v, want %v", tt.s, tt.l, got, tt.want)
		}
	}
}

func TestHandleErrs(t *testing.T) {
	a := newTestApp(t)
	if err := a.handleErrs(nil); err != nil {
		t.Errorf("handleErrs(nil) = %v", err)
	}
	if err := a.handleErrs([]error{}); err != nil {
		t.Errorf("handleErrs(empty) = %v", err)
	}
	err := a.handleErrs([]error{errors.New("a"), errors.New("b")})
	if err == nil || !strings.Contains(err.Error(), "2 error(s)") {
		t.Errorf("handleErrs(2) = %v", err)
	}
}

func TestFormatDurationMS(t *testing.T) {
	tests := map[int64]string{
		0:                              "0.000",
		int64(time.Millisecond):        "1.000",
		int64(1500 * time.Microsecond): "1.500",
		int64(time.Second):             "1000.000",
		123456:                         "0.123",
	}
	for in, want := range tests {
		if got := formatDurationMS(in); got != want {
			t.Errorf("formatDurationMS(%d) = %q, want %q", in, got, want)
		}
	}
}
