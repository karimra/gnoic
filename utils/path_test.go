package utils

import (
	"errors"
	"testing"

	"github.com/openconfig/gnoi/types"
	"google.golang.org/protobuf/proto"
)

func TestParsePath(t *testing.T) {
	tests := []struct {
		name    string
		in      string
		want    *types.Path
		wantErr error
	}{
		{
			name: "empty",
			in:   "",
			want: &types.Path{},
		},
		{
			name: "root",
			in:   "/",
			want: &types.Path{Elem: []*types.PathElem{}},
		},
		{
			name: "absolute",
			in:   "/a/b/c",
			want: &types.Path{Elem: []*types.PathElem{{Name: "a"}, {Name: "b"}, {Name: "c"}}},
		},
		{
			name: "relative",
			in:   "a/b",
			want: &types.Path{Elem: []*types.PathElem{{Name: "a"}, {Name: "b"}}},
		},
		{
			name: "trailing slash",
			in:   "a/b/",
			want: &types.Path{Elem: []*types.PathElem{{Name: "a"}, {Name: "b"}}},
		},
		{
			name: "single key",
			in:   "/a[k=v]/b",
			want: &types.Path{Elem: []*types.PathElem{{Name: "a", Key: map[string]string{"k": "v"}}, {Name: "b"}}},
		},
		{
			name: "multiple keys",
			in:   "a[k1=v1][k2=v2]",
			want: &types.Path{Elem: []*types.PathElem{{Name: "a", Key: map[string]string{"k1": "v1", "k2": "v2"}}}},
		},
		{
			name: "key value containing slash",
			in:   "a[k=x/y]/b",
			want: &types.Path{Elem: []*types.PathElem{{Name: "a", Key: map[string]string{"k": "x/y"}}, {Name: "b"}}},
		},
		{
			name: "escaped brackets in key value",
			in:   `a[k=v\]x]`,
			want: &types.Path{Elem: []*types.PathElem{{Name: "a", Key: map[string]string{"k": "v]x"}}}},
		},
		{
			name: "origin with absolute path",
			in:   "openconfig:/a/b",
			want: &types.Path{Origin: "openconfig", Elem: []*types.PathElem{{Name: "a"}, {Name: "b"}}},
		},
		{
			name: "origin only",
			in:   "openconfig:",
			want: &types.Path{Origin: "openconfig", Elem: []*types.PathElem{}},
		},
		{
			name: "colon in element is not an origin",
			in:   "/a:b/c",
			want: &types.Path{Elem: []*types.PathElem{{Name: "a:b"}, {Name: "c"}}},
		},
		{
			name: "colon in key value is not an origin",
			in:   "a[k=1:2]",
			want: &types.Path{Elem: []*types.PathElem{{Name: "a", Key: map[string]string{"k": "1:2"}}}},
		},
		{
			name:    "unterminated key",
			in:      "a[k=v",
			wantErr: errMalformedXPath,
		},
		{
			name:    "nested open bracket",
			in:      "a[k=[v]",
			wantErr: errMalformedXPath,
		},
		{
			name:    "stray close bracket",
			in:      "a]",
			wantErr: errMalformedXPath,
		},
		{
			name:    "key without value",
			in:      "a[k]",
			wantErr: errMalformedXPathKey,
		},
		{
			name:    "empty key name",
			in:      "a[=v]",
			wantErr: errMalformedXPathKey,
		},
		{
			name:    "empty key value",
			in:      "a[k=]",
			wantErr: errMalformedXPathKey,
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got, err := ParsePath(tt.in)
			if tt.wantErr != nil {
				if !errors.Is(err, tt.wantErr) {
					t.Fatalf("ParsePath(%q) error = %v, want %v", tt.in, err, tt.wantErr)
				}
				return
			}
			if err != nil {
				t.Fatalf("ParsePath(%q) unexpected error: %v", tt.in, err)
			}
			if !proto.Equal(got, tt.want) {
				t.Fatalf("ParsePath(%q)\n got: %v\nwant: %v", tt.in, got, tt.want)
			}
		})
	}
}

func TestPathToXPath(t *testing.T) {
	tests := []struct {
		name string
		in   *types.Path
		want string
	}{
		{name: "nil", in: nil, want: ""},
		{name: "empty", in: &types.Path{}, want: ""},
		{
			name: "elems",
			in:   &types.Path{Elem: []*types.PathElem{{Name: "a"}, {Name: "b"}}},
			want: "a/b",
		},
		{
			name: "elems with key",
			in:   &types.Path{Elem: []*types.PathElem{{Name: "a", Key: map[string]string{"k": "v"}}, {Name: "b"}}},
			want: "a[k=v]/b",
		},
		{
			name: "origin",
			in:   &types.Path{Origin: "oc", Elem: []*types.PathElem{{Name: "a"}}},
			want: "oc:a",
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if got := PathToXPath(tt.in); got != tt.want {
				t.Fatalf("PathToXPath() = %q, want %q", got, tt.want)
			}
		})
	}
}

// TestPathRoundTrip verifies that formatting a parsed path and parsing it again
// yields the same path (for paths with at most one key per element, since map
// ordering is not stable).
func TestPathRoundTrip(t *testing.T) {
	for _, in := range []string{
		"a/b/c",
		"a[k=v]/b[x=y]/c",
		"oc:a/b",
		"interfaces/interface[name=eth0]/state",
	} {
		p, err := ParsePath(in)
		if err != nil {
			t.Fatalf("ParsePath(%q): %v", in, err)
		}
		xp := PathToXPath(p)
		p2, err := ParsePath(xp)
		if err != nil {
			t.Fatalf("ParsePath(%q): %v", xp, err)
		}
		if !proto.Equal(p, p2) {
			t.Fatalf("round trip mismatch for %q: %v != %v", in, p, p2)
		}
	}
}
