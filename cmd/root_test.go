package cmd

import (
	"sort"
	"strings"
	"testing"

	"github.com/spf13/cobra"
)

// expectedTree mirrors the command tree documented in the README.
var expectedTree = map[string][]string{
	"gnoic":         {"cert", "completion", "factory-reset", "file", "healthz", "help", "os", "server", "services", "system", "tree", "version"},
	"cert":          {"can-generate-csr", "create-ca", "generate-csr", "get-certs", "install", "load", "load-ca", "revoke", "rotate"},
	"factory-reset": {"start"},
	"file":          {"get", "put", "remove", "stat", "transfer"},
	"healthz":       {"ack", "artifact", "check", "get", "list"},
	"os":            {"activate", "install", "verify"},
	"system":        {"cancel-reboot", "kill-process", "ping", "reboot", "reboot-status", "set-package", "switch-control-processor", "time", "traceroute"},
	"version":       {"upgrade"},
}

func subCommandNames(c *cobra.Command) []string {
	names := make([]string, 0, len(c.Commands()))
	for _, sc := range c.Commands() {
		names = append(names, sc.Name())
	}
	sort.Strings(names)
	return names
}

// findCommand looks a command up by name, or by its full path when name
// contains spaces (e.g. "os install").
func findCommand(root *cobra.Command, name string) *cobra.Command {
	if root.Name() == name || root.CommandPath() == "gnoic "+name {
		return root
	}
	for _, sc := range root.Commands() {
		if c := findCommand(sc, name); c != nil {
			return c
		}
	}
	return nil
}

func TestCommandTree(t *testing.T) {
	root := newRootCmd()
	// cobra adds completion/help lazily during execution
	root.InitDefaultCompletionCmd()
	root.InitDefaultHelpCmd()

	for parent, want := range expectedTree {
		c := findCommand(root, parent)
		if c == nil {
			t.Errorf("command %q not found", parent)
			continue
		}
		got := subCommandNames(c)
		if strings.Join(got, ",") != strings.Join(want, ",") {
			t.Errorf("%s subcommands:\n got %v\nwant %v", parent, got, want)
		}
	}
}

func TestLeafCommandsAreRunnable(t *testing.T) {
	root := newRootCmd()
	var walk func(c *cobra.Command)
	walk = func(c *cobra.Command) {
		if !c.HasSubCommands() {
			if c.RunE == nil && c.Run == nil {
				t.Errorf("leaf command %q has no Run/RunE", c.CommandPath())
			}
			if c.Short == "" {
				t.Errorf("leaf command %q has no Short description", c.CommandPath())
			}
		}
		for _, sc := range c.Commands() {
			walk(sc)
		}
	}
	walk(root)
}

func TestGlobalFlags(t *testing.T) {
	root := newRootCmd()
	for _, name := range []string{
		"config", "address", "username", "password", "port", "insecure", "tls-ca", "tls-cert", "tls-key",
		"timeout", "debug", "skip-verify", "proxy-from-env", "format", "print-proto",
		"tls-min-version", "tls-max-version", "tls-version", "gzip",
	} {
		if root.PersistentFlags().Lookup(name) == nil {
			t.Errorf("global flag --%s not registered", name)
		}
	}
	if f := root.PersistentFlags().Lookup("port"); f != nil && f.DefValue != "57400" {
		t.Errorf("--port default = %q, want 57400", f.DefValue)
	}
	if f := root.PersistentFlags().Lookup("timeout"); f != nil && f.DefValue != "10s" {
		t.Errorf("--timeout default = %q, want 10s", f.DefValue)
	}
	if f := root.PersistentFlags().ShorthandLookup("a"); f == nil || f.Name != "address" {
		t.Errorf("-a shorthand not bound to --address")
	}
}

func TestLocalFlagsPerCommand(t *testing.T) {
	root := newRootCmd()
	tests := map[string][]string{
		"get":          {"file", "dst", "target-prefix"},
		"put":          {"file", "dst", "chunk-size", "permission", "hash-method"},
		"stat":         {"path", "humanize", "recursive"},
		"remove":       {"path"},
		"transfer":     {"local", "remote", "source-address"},
		"ping":         {"destination", "source", "count", "interval", "wait", "size", "do-not-fragment", "do-not-resolve", "protocol", "ns"},
		"kill-process": {"pid", "name", "signal", "restart"},
		"set-package":  {"pkg", "dst", "version", "activate", "remote", "content-chunk-size"},
		"create-ca":    {"org", "org-unit", "country", "state", "locality", "validity", "key-size", "email", "common-name", "key-out", "cert-out"},
		"rotate":       {"id", "key-type", "cert-type", "min-key-size", "common-name", "validity", "print-csr", "gen-csr"},
		"os install":   {"version", "standby", "pkg", "content-chunk-size"},
		"cert install": {"id", "key-type", "cert-type", "min-key-size", "common-name", "validity", "print-csr", "gen-csr"},
		"server":       {"file", "file-hash"},
		"tree":         {"flat", "details"},
		"start":        {"factory-os", "zero-fill"},
		"upgrade":      {"use-pkg"},
	}
	for name, flags := range tests {
		c := findCommand(root, name)
		if c == nil {
			t.Errorf("command %q not found", name)
			continue
		}
		for _, f := range flags {
			if c.Flags().Lookup(f) == nil {
				t.Errorf("%s: flag --%s not registered", c.CommandPath(), f)
			}
		}
	}
	// cert subcommands inherit --ca-cert / --ca-key from the cert command
	if c := findCommand(root, "cert"); c != nil {
		for _, f := range []string{"ca-cert", "ca-key"} {
			if c.PersistentFlags().Lookup(f) == nil {
				t.Errorf("cert: persistent flag --%s not registered", f)
			}
		}
	}
}

func TestPreRunValidation(t *testing.T) {
	// commands with required flags must fail fast with a clear error before
	// attempting to connect to any target.
	tests := []struct {
		args    []string
		wantErr string
	}{
		{[]string{"file", "put", "-a", "127.0.0.1"}, "missing --file flag"},
		{[]string{"file", "transfer", "-a", "127.0.0.1"}, "missing local file path"},
		{[]string{"file", "transfer", "-a", "127.0.0.1", "--local", "/x"}, "missing remote file path"},
		{[]string{"system", "ping", "-a", "127.0.0.1"}, "flag --destination is required"},
		{[]string{"system", "ping", "-a", "127.0.0.1", "--destination", "1.1.1.1", "--protocol", "IPX"}, "unknown protocol"},
		{[]string{"system", "kill-process", "-a", "127.0.0.1"}, "specify --name or --pid"},
		{[]string{"system", "kill-process", "-a", "127.0.0.1", "--pid", "1", "--signal", "FOO"}, "unknown kill signal"},
		{[]string{"system", "set-package", "-a", "127.0.0.1"}, "missing --pkg flag"},
		{[]string{"os", "install", "-a", "127.0.0.1"}, "missing --version flag"},
		{[]string{"os", "install", "-a", "127.0.0.1", "--version", "1"}, "missing --pkg flag"},
	}
	for _, tt := range tests {
		t.Run(strings.Join(tt.args, " "), func(t *testing.T) {
			root := newRootCmd()
			root.SilenceErrors = true
			root.SilenceUsage = true
			root.SetArgs(tt.args)
			err := root.Execute()
			if err == nil {
				t.Fatalf("expected error containing %q", tt.wantErr)
			}
			if !strings.Contains(err.Error(), tt.wantErr) {
				t.Errorf("error = %q, want it to contain %q", err.Error(), tt.wantErr)
			}
		})
	}
}

func TestTreeCommandRuns(t *testing.T) {
	for _, args := range [][]string{{"tree"}, {"tree", "--flat"}, {"tree", "--details"}} {
		root := newRootCmd()
		root.SetArgs(args)
		if err := root.Execute(); err != nil {
			t.Errorf("%v: %v", args, err)
		}
	}
}

func TestVersionCommandRuns(t *testing.T) {
	root := newRootCmd()
	root.SetArgs([]string{"version"})
	if err := root.Execute(); err != nil {
		t.Errorf("version: %v", err)
	}
}
