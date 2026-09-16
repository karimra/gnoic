package api

import (
	"context"
	"crypto/tls"
	"errors"
	"net"
	"path/filepath"
	"testing"
	"time"

	"github.com/karimra/gnoic/config"
	"google.golang.org/grpc"
)

func TestNewTarget_Defaults(t *testing.T) {
	tg, err := NewTarget(Address("10.0.0.1:57400"))
	if err != nil {
		t.Fatal(err)
	}
	if tg.Config.Name != "10.0.0.1:57400" {
		t.Errorf("Name = %q, want address", tg.Config.Name)
	}
	if tg.Config.Timeout != defaultTargetTimeout {
		t.Errorf("Timeout = %v, want %v", tg.Config.Timeout, defaultTargetTimeout)
	}
	if tg.Config.Insecure == nil || *tg.Config.Insecure {
		t.Errorf("Insecure = %v, want false", tg.Config.Insecure)
	}
	if tg.Config.SkipVerify == nil || *tg.Config.SkipVerify {
		t.Errorf("SkipVerify = %v, want false", tg.Config.SkipVerify)
	}
	if tg.client != nil {
		t.Error("client should be nil before CreateGrpcClient")
	}
	if err := tg.Close(); err != nil {
		t.Errorf("Close() on unconnected target: %v", err)
	}
}

func TestNewTarget_MissingAddress(t *testing.T) {
	if _, err := NewTarget(Name("r1")); err == nil {
		t.Fatal("expected error for missing address")
	}
}

func TestNewTarget_OptionError(t *testing.T) {
	boom := errors.New("boom")
	_, err := NewTarget(Address("a:1"), func(*Target) error { return boom })
	if !errors.Is(err, boom) {
		t.Fatalf("got %v, want %v", err, boom)
	}
}

func TestNewTarget_Options(t *testing.T) {
	tlsCfg := &tls.Config{ServerName: "x"}
	tg, err := NewTarget(
		Name("r1"),
		Address("10.0.0.1:57400"),
		Address("10.0.0.2:57400"),
		Username("u"),
		Password("p"),
		Timeout(3*time.Second),
		Insecure(true),
		SkipVerify(true),
		TLSCA("/ca"),
		TLSCert("/cert"),
		TLSKey("/key"),
		TLSMinVersion("1.2"),
		TLSMaxVersion("1.3"),
		TLSVersion("1.3"),
		TLSConfig(tlsCfg),
		Gzip(true),
	)
	if err != nil {
		t.Fatal(err)
	}
	c := tg.Config
	if c.Name != "r1" {
		t.Errorf("Name = %q", c.Name)
	}
	if c.Address != "10.0.0.1:57400,10.0.0.2:57400" {
		t.Errorf("Address = %q", c.Address)
	}
	if *c.Username != "u" || *c.Password != "p" {
		t.Errorf("credentials = %q/%q", *c.Username, *c.Password)
	}
	if c.Timeout != 3*time.Second {
		t.Errorf("Timeout = %v", c.Timeout)
	}
	if !*c.Insecure || !*c.SkipVerify || !*c.Gzip {
		t.Errorf("bool options not applied: %s", c)
	}
	if *c.TLSCA != "/ca" || *c.TLSCert != "/cert" || *c.TLSKey != "/key" {
		t.Errorf("TLS file options not applied: %s", c)
	}
	if c.TLSMinVersion != "1.2" || c.TLSMaxVersion != "1.3" || c.TLSVersion != "1.3" {
		t.Errorf("TLS version options not applied: %s", c)
	}
	// TLSConfig() is consumed by DialOpts when the target is not insecure
	c.Insecure = nil
	tg2, _ := NewTarget(Address("a:1"), TLSConfig(tlsCfg))
	opts, err := tg2.Config.DialOpts()
	if err != nil || len(opts) != 1 {
		t.Errorf("DialOpts() with preset TLS config = %v, %v", opts, err)
	}
}

func TestNewTarget_MultipleAddressesName(t *testing.T) {
	tg, err := NewTarget(Address("a:1"), Address("b:2"))
	if err != nil {
		t.Fatal(err)
	}
	if tg.Config.Name != "a:1" {
		t.Errorf("Name = %q, want first address", tg.Config.Name)
	}
}

func TestNewTargetFromConfig(t *testing.T) {
	tc := &config.TargetConfig{Name: "r1", Address: "a:1"}
	tg := NewTargetFromConfig(tc)
	if tg.Config != tc {
		t.Error("NewTargetFromConfig should keep the provided config pointer")
	}
}

func TestCreateGrpcClient_UnixSocket(t *testing.T) {
	sock := filepath.Join(t.TempDir(), "s.sock")
	l, err := net.Listen("unix", sock)
	if err != nil {
		t.Fatal(err)
	}
	srv := grpc.NewServer()
	go srv.Serve(l)
	t.Cleanup(srv.Stop)

	tg, err := NewTarget(Address("unix://"+sock), Insecure(true), Timeout(2*time.Second))
	if err != nil {
		t.Fatal(err)
	}
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := tg.CreateGrpcClient(ctx); err != nil {
		t.Fatalf("CreateGrpcClient: %v", err)
	}
	defer tg.Close()
	if tg.Conn() == nil {
		t.Fatal("Conn() is nil after CreateGrpcClient")
	}
	// exercise the custom dialer: a raw connection to the unix socket must succeed
	dial := tg.createDialer(tg.Config.Address)
	conn, err := dial(ctx, "")
	if err != nil {
		t.Fatalf("custom dialer: %v", err)
	}
	conn.Close()

	// typed clients are constructed on top of the connection
	if tg.CertClient() == nil || tg.FileClient() == nil || tg.NewOsClient() == nil || tg.SystemClient() == nil {
		t.Error("service client constructors returned nil")
	}
}

func TestCreateGrpcClient_TLSError(t *testing.T) {
	tg, err := NewTarget(Address("a:1"), TLSCA("/definitely/missing/ca.pem"))
	if err != nil {
		t.Fatal(err)
	}
	if err := tg.CreateGrpcClient(context.Background()); err == nil {
		t.Fatal("expected error from unreadable CA file")
	}
}
