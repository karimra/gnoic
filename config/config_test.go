package config

import (
	"io"
	"os"
	"path/filepath"
	"testing"
	"time"

	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
)

func newTestConfig() *Config {
	// SetLogger uses the logrus standard logger; keep test output quiet.
	log.StandardLogger().SetOutput(io.Discard)
	c := New()
	c.SetLogger()
	return c
}

// newTestCmdTree builds gnoic -> file -> get with a persistent root flag and a
// local slice flag on the leaf, mirroring how the real command tree is wired.
func newTestCmdTree(c *Config) (root, leaf *cobra.Command) {
	root = &cobra.Command{Use: "gnoic"}
	root.PersistentFlags().StringSliceVarP(&c.Address, "address", "a", nil, "")
	root.PersistentFlags().StringVar(&c.Username, "username", "", "")

	mid := &cobra.Command{Use: "file"}
	leaf = &cobra.Command{Use: "get", Run: func(*cobra.Command, []string) {}}
	leaf.Flags().StringSlice("file", nil, "")
	leaf.Flags().String("dst", "", "")
	leaf.Flags().Bool("target-prefix", false, "")

	mid.AddCommand(leaf)
	root.AddCommand(mid)
	return root, leaf
}

func TestFlagFullName(t *testing.T) {
	c := newTestConfig()
	root, leaf := newTestCmdTree(c)

	tests := []struct {
		cmd  *cobra.Command
		flag string
		want string
	}{
		{root, "address", "address"},
		{leaf, "file", "file-get-file"},
		{leaf.Parent(), "x", "file-x"},
	}
	for _, tt := range tests {
		if got := flagFullName(tt.cmd, tt.flag); got != tt.want {
			t.Errorf("flagFullName(%s, %s) = %q, want %q", tt.cmd.Name(), tt.flag, got, tt.want)
		}
	}
}

func TestSetLocalFlagsFromFile(t *testing.T) {
	c := newTestConfig()
	_, leaf := newTestCmdTree(c)

	c.FileConfig.Set("file-get-file", []interface{}{"a.txt", "b.txt"})
	c.FileConfig.Set("file-get-dst", "/tmp/out")
	c.FileConfig.Set("file-get-target-prefix", true)

	c.SetLocalFlagsFromFile(leaf)

	files, err := leaf.Flags().GetStringSlice("file")
	if err != nil {
		t.Fatal(err)
	}
	if len(files) != 2 || files[0] != "a.txt" || files[1] != "b.txt" {
		t.Errorf("file flag = %v, want [a.txt b.txt]", files)
	}
	if dst, _ := leaf.Flags().GetString("dst"); dst != "/tmp/out" {
		t.Errorf("dst flag = %q, want /tmp/out", dst)
	}
	if tp, _ := leaf.Flags().GetBool("target-prefix"); !tp {
		t.Errorf("target-prefix flag = false, want true")
	}
}

func TestSetLocalFlagsFromFile_ExplicitFlagWins(t *testing.T) {
	c := newTestConfig()
	_, leaf := newTestCmdTree(c)

	if err := leaf.Flags().Set("dst", "/from/cli"); err != nil {
		t.Fatal(err)
	}
	c.FileConfig.Set("file-get-dst", "/from/file")

	c.SetLocalFlagsFromFile(leaf)

	if dst, _ := leaf.Flags().GetString("dst"); dst != "/from/cli" {
		t.Errorf("dst flag = %q, want value set on the command line to be kept", dst)
	}
}

func TestSetPersistantFlagsFromFile(t *testing.T) {
	c := newTestConfig()
	root, _ := newTestCmdTree(c)

	c.FileConfig.Set("username", "admin")
	c.FileConfig.Set("address", []interface{}{"10.0.0.1", "10.0.0.2"})

	// during Execute cobra merges persistent flags into Flags(); emulate that.
	if err := root.ParseFlags(nil); err != nil {
		t.Fatal(err)
	}
	c.SetPersistantFlagsFromFile(root)

	if c.Username != "admin" {
		t.Errorf("Username = %q, want admin", c.Username)
	}
	if len(c.Address) != 2 || c.Address[0] != "10.0.0.1" || c.Address[1] != "10.0.0.2" {
		t.Errorf("Address = %v, want [10.0.0.1 10.0.0.2]", c.Address)
	}
}

func TestLoad_FromExplicitFile(t *testing.T) {
	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "gnoic.yaml")
	content := `
username: admin
password: secret
port: "9339"
targets:
  10.0.0.1:
    username: other
    timeout: 5s
  10.0.0.2:57400:
`
	if err := os.WriteFile(cfgFile, []byte(content), 0600); err != nil {
		t.Fatal(err)
	}

	c := newTestConfig()
	c.CfgFile = cfgFile
	if err := c.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got := c.FileConfig.GetString("username"); got != "admin" {
		t.Errorf("username = %q, want admin", got)
	}
	if got := c.FileConfig.GetString("port"); got != "9339" {
		t.Errorf("port = %q, want 9339", got)
	}
	targets := c.FileConfig.GetStringMap("targets")
	if len(targets) != 2 {
		t.Errorf("targets = %v, want 2 entries", targets)
	}
}

func TestLoad_MissingExplicitFile(t *testing.T) {
	c := newTestConfig()
	c.CfgFile = filepath.Join(t.TempDir(), "does-not-exist.yaml")
	if err := c.Load(); err == nil {
		t.Fatal("Load() expected an error for a missing explicit config file")
	}
}

func TestLoad_EnvOverride(t *testing.T) {
	t.Setenv("GNOIC_USERNAME", "from-env")
	t.Setenv("GNOIC_FILE_GET_DST", "/env/dst")

	dir := t.TempDir()
	cfgFile := filepath.Join(dir, "gnoic.yaml")
	if err := os.WriteFile(cfgFile, []byte("username: from-file\n"), 0600); err != nil {
		t.Fatal(err)
	}
	c := newTestConfig()
	c.CfgFile = cfgFile
	if err := c.Load(); err != nil {
		t.Fatalf("Load() error: %v", err)
	}
	if got := c.FileConfig.GetString("username"); got != "from-env" {
		t.Errorf("username = %q, want env var to override the file", got)
	}
	if got := c.FileConfig.GetString("file-get-dst"); got != "/env/dst" {
		t.Errorf("file-get-dst = %q, want /env/dst", got)
	}
}

func TestSetLogger(t *testing.T) {
	std := log.StandardLogger()
	std.SetOutput(io.Discard)
	prev := std.GetLevel()
	t.Cleanup(func() { std.SetLevel(prev) })

	c := New()
	c.Debug = true
	c.SetLogger()
	if c.LogOutput() == nil {
		t.Fatal("LogOutput() returned nil")
	}
	if !c.logger.Logger.IsLevelEnabled(logDebugLevel()) {
		t.Error("debug level not enabled when Debug=true")
	}
}

func TestNewDefaults(t *testing.T) {
	c := New()
	if c.FileConfig == nil {
		t.Fatal("FileConfig is nil")
	}
	if c.Timeout != 0 || c.Port != "" {
		t.Errorf("unexpected non-zero defaults: timeout=%v port=%q", c.Timeout, c.Port)
	}
	_ = time.Second
}
