package app

import (
	"fmt"
	"os"
	"sync"
	"time"

	"github.com/karimra/gnoic/api"
	"github.com/karimra/gnoic/config"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc"
	"google.golang.org/grpc/encoding/gzip"
	"google.golang.org/grpc/grpclog"
	"google.golang.org/protobuf/encoding/prototext"
	"google.golang.org/protobuf/proto"
)

type App struct {
	RootCmd *cobra.Command

	wg      *sync.WaitGroup
	Config  *config.Config
	m       *sync.Mutex
	Targets map[string]*api.Target
	Logger  *log.Entry
	// print mutex
	pm *sync.Mutex
}

type TargetResponse interface {
	Target() string
	Response() any
}

func New() *App {
	logger := log.New()
	a := &App{
		RootCmd: new(cobra.Command),
		wg:      new(sync.WaitGroup),
		Config:  config.New(),
		m:       new(sync.Mutex),
		Targets: make(map[string]*api.Target),
		Logger:  log.NewEntry(logger),
		pm:      new(sync.Mutex),
	}
	return a
}

func (a *App) InitGlobalFlags() {
	a.RootCmd.ResetFlags()

	a.RootCmd.PersistentFlags().StringVar(&a.Config.CfgFile, "config", "", "config file (default is $HOME/gnoic.yaml)")
	a.RootCmd.PersistentFlags().StringSliceVarP(&a.Config.GlobalFlags.Address, "address", "a", []string{}, "comma separated gNOI targets addresses")
	a.RootCmd.PersistentFlags().StringVarP(&a.Config.GlobalFlags.Username, "username", "u", "", "username")
	a.RootCmd.PersistentFlags().StringVarP(&a.Config.GlobalFlags.Password, "password", "p", "", "password")
	a.RootCmd.PersistentFlags().StringVarP(&a.Config.GlobalFlags.Port, "port", "", defaultGrpcPort, "gRPC port")
	a.RootCmd.PersistentFlags().BoolVarP(&a.Config.GlobalFlags.Insecure, "insecure", "", false, "insecure connection")
	a.RootCmd.PersistentFlags().StringVarP(&a.Config.GlobalFlags.TLSCa, "tls-ca", "", "", "tls certificate authority")
	a.RootCmd.PersistentFlags().StringVarP(&a.Config.GlobalFlags.TLSCert, "tls-cert", "", "", "tls certificate")
	a.RootCmd.PersistentFlags().StringVarP(&a.Config.GlobalFlags.TLSKey, "tls-key", "", "", "tls key")
	a.RootCmd.PersistentFlags().DurationVarP(&a.Config.GlobalFlags.Timeout, "timeout", "", 10*time.Second, "grpc timeout, valid formats: 10s, 1m30s, 1h")
	a.RootCmd.PersistentFlags().BoolVarP(&a.Config.GlobalFlags.Debug, "debug", "d", false, "debug mode")
	a.RootCmd.PersistentFlags().BoolVarP(&a.Config.GlobalFlags.SkipVerify, "skip-verify", "", false, "skip verify tls connection")
	a.RootCmd.PersistentFlags().BoolVarP(&a.Config.GlobalFlags.ProxyFromEnv, "proxy-from-env", "", false, "use proxy from environment")
	a.RootCmd.PersistentFlags().StringVarP(&a.Config.GlobalFlags.Format, "format", "", "text", "output format, one of: text, json")
	// a.RootCmd.PersistentFlags().StringVarP(&a.Config.GlobalFlags.LogFile, "log-file", "", "", "log file path")
	// a.RootCmd.PersistentFlags().BoolVarP(&a.Config.GlobalFlags.Log, "log", "", false, "write log messages to stderr")
	a.RootCmd.PersistentFlags().BoolVarP(&a.Config.GlobalFlags.PrintProto, "print-proto", "", false, "print request(s)/responses(s) in prototext format")
	// a.RootCmd.PersistentFlags().DurationVarP(&a.Config.GlobalFlags.Retry, "retry", "", defaultRetryTimer, "retry timer for RPCs")
	a.RootCmd.PersistentFlags().StringVarP(&a.Config.GlobalFlags.TLSMinVersion, "tls-min-version", "", "", fmt.Sprintf("minimum TLS supported version, one of %q", tlsVersions))
	a.RootCmd.PersistentFlags().StringVarP(&a.Config.GlobalFlags.TLSMaxVersion, "tls-max-version", "", "", fmt.Sprintf("maximum TLS supported version, one of %q", tlsVersions))
	a.RootCmd.PersistentFlags().StringVarP(&a.Config.GlobalFlags.TLSVersion, "tls-version", "", "", fmt.Sprintf("set TLS version. Overwrites --tls-min-version and --tls-max-version, one of %q", tlsVersions))
	a.RootCmd.PersistentFlags().BoolVarP(&a.Config.GlobalFlags.Gzip, "gzip", "", false, "enable gzip compression on gRPC connections")

	a.RootCmd.PersistentFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(flag.Name, flag)
	})
}

func (a *App) PreRun(cmd *cobra.Command, args []string) error {
	// init logger
	a.Config.SetLogger()
	if a.Config.Debug {
		a.Logger.Logger.SetLevel(log.DebugLevel)
		grpclog.SetLogger(a.Logger) //lint:ignore SA1019 .
	}
	a.Config.SetPersistantFlagsFromFile(a.RootCmd)
	return nil
}

func (a *App) createBaseDialOpts() []grpc.DialOption {
	opts := []grpc.DialOption{grpc.WithBlock()}
	if !a.Config.ProxyFromEnv {
		opts = append(opts, grpc.WithNoProxy())
	}
	if a.Config.Gzip {
		opts = append(opts, grpc.WithDefaultCallOptions(grpc.UseCompressor(gzip.Name)))
	}
	return opts
}

func (a *App) printProtoMsg(targetName string, m proto.Message) {
	if !a.Config.PrintProto {
		return
	}
	a.pm.Lock()
	defer a.pm.Unlock()
	fmt.Fprintf(os.Stdout, "%q:\n%s\n%s\n",
		targetName,
		m.ProtoReflect().Descriptor().FullName(),
		prototext.Format(m))
}
