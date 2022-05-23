package app

import (
	"bufio"
	"context"
	"errors"
	"fmt"
	"io"
	"os"
	"time"

	"github.com/karimra/gnoic/api"
	gos "github.com/karimra/gnoic/api/os"
	gnoios "github.com/openconfig/gnoi/os"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
)

type osInstallResponse struct {
	TargetError
	// rsp *gnoios.InstallResponse
}

func (a *App) InitOSInstallFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.OsInstallVersion, "version", "", "package version")
	cmd.Flags().BoolVar(&a.Config.OsInstallStandbySupervisor, "standby", false, "install on standby supervisor")
	cmd.Flags().StringVar(&a.Config.OsInstallPackage, "pkg", "", "path to the os package file to install")
	cmd.Flags().Uint64Var(&a.Config.OsInstallContentSize, "content-chunk-size", 1024*1024, "max chunk size to transfer the package")
	//
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) PreRunEOSInstall(cmd *cobra.Command, args []string) error {
	if a.Config.OsInstallVersion == "" {
		return errors.New("missing --version flag")
	}
	if a.Config.OsInstallPackage == "" {
		return errors.New("missing --pkg flag")
	}
	_, err := os.Stat(a.Config.OsInstallPackage)
	return err
}

func (a *App) RunEOSInstall(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *osInstallResponse, numTargets)

	a.wg.Add(numTargets)
	for _, t := range targets {
		go func(t *api.Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &osInstallResponse{
					TargetError: TargetError{
						TargetName: t.Config.Name,
						Err:        err,
					},
				}
				return
			}
			defer t.Close()
			a.Logger.Infof("starting install RPC")
			err = a.OsInstall(ctx, t)
			responseChan <- &osInstallResponse{
				TargetError: TargetError{
					TargetName: t.Config.Name,
					Err:        err,
				},
			}
		}(t)
	}
	a.wg.Wait()
	close(responseChan)
	for rsp := range responseChan {
		if rsp.Err != nil {
			fmt.Printf("%+v\n", rsp)
		}
	}
	return nil
}

func (a *App) OsInstall(ctx context.Context, t *api.Target) error {
	// start stream
	osc := gnoios.NewOSClient(t.Conn())
	osInstallClient, err := osc.Install(ctx)
	if err != nil {
		return err
	}
	a.Logger.Infof("target %q: starting Install stream", t.Config.Name)
	req, err := gos.NewOSInstallTransferRequest(
		gos.Version(a.Config.OsInstallVersion),
		gos.StandbySupervisor(a.Config.OsInstallStandbySupervisor),
	)
	if err != nil {
		return err
	}
	a.printMsg(t.Config.Name, req)
	err = osInstallClient.Send(req)
	if err != nil {
		return err
	}
RCV:
	a.Logger.Debugf("target %q: OS Install stream rcv...", t.Config.Name)
	rsp, err := osInstallClient.Recv()
	if err != nil {
		a.Logger.Debugf("target %q: OS Install stream rcv err: %v", t.Config.Name, err)
		return err
	}
	a.Logger.Debugf("target %q: OS Install stream got: %+v", t.Config.Name, rsp)
	a.printMsg(t.Config.Name, rsp)
	switch rsp := rsp.GetResponse().(type) {
	case *gnoios.InstallResponse_TransferReady:
		err = a.osInstallTransferContent(ctx, t, osInstallClient)
		if err != nil {
			return err
		}
		a.Logger.Debugf("target %q: sent transfer end...", t.Config.Name)
		goto RCV
	case *gnoios.InstallResponse_Validated:
		a.Logger.Debugf("target %q: Validated %v", t.Config.Name, rsp.Validated.String())
		return nil
	case *gnoios.InstallResponse_InstallError:
		a.Logger.Errorf("target %q Install RPC failed: %v: %v", t.Config.Name, rsp.InstallError.GetType(), rsp.InstallError.GetDetail())
		return fmt.Errorf("%v: %v", rsp.InstallError.GetType(), rsp.InstallError.GetDetail())
	case *gnoios.InstallResponse_SyncProgress:
		a.Logger.Debugf("target %q: SyncProgress %v", t.Config.Name, rsp.SyncProgress.String())
		time.Sleep(time.Second)
		goto RCV
	case *gnoios.InstallResponse_TransferProgress:
		a.Logger.Infof("target %q: TransferProgress %v", t.Config.Name, rsp.TransferProgress.String())
		goto RCV
	}
	return nil
}

func (a *App) osInstallTransferContent(ctx context.Context, t *api.Target, osic gnoios.OS_InstallClient) error {
	// read file
	pkg, err := os.Open(a.Config.OsInstallPackage)
	if err != nil {
		return err
	}
	defer pkg.Close()
	errCh := make(chan error)
	doneCh := make(chan struct{})

	go func() {
		defer a.Logger.Infof("target %q: TransferContent done...", t.Config.Name)
		for {
			select {
			case <-doneCh:
				return
			default:
				rsp, err := osic.Recv()
				if err != nil {
					errCh <- err
					return
				}
				a.printMsg(t.Config.Name, rsp)
				switch rsp := rsp.GetResponse().(type) {
				case *gnoios.InstallResponse_InstallError:
					a.Logger.Errorf("target %q Install Content Transfer RPC failed: %v: %v", t.Config.Name, rsp.InstallError.GetType(), rsp.InstallError.GetDetail())
					errCh <- fmt.Errorf("%v: %v", rsp.InstallError.GetType(), rsp.InstallError.GetDetail())
					return
				case *gnoios.InstallResponse_TransferProgress:
					a.Logger.Infof("target %q: TransferProgress %v", t.Config.Name, rsp.TransferProgress.String())
				}
			}
		}
	}()

	r := bufio.NewReader(pkg)
	buf := make([]byte, 0, a.Config.OsInstallContentSize)
OUTER:
	for {
		select {
		case err := <-errCh:
			return err
		default:
			n, err := r.Read(buf[:cap(buf)])
			if err != nil {
				if err == io.EOF {
					a.Logger.Debugf("target %q: file read EOF", t.Config.Name)
					break OUTER
				}
				a.Logger.Errorf("target %q: file read err: %v", t.Config.Name, err)
				close(doneCh)
				return err
			}
			a.Logger.Debugf("target %q: read %d bytes from file", t.Config.Name, n)
			buf = buf[:n]
			a.Logger.Debugf("target %q: sending %d bytes", t.Config.Name, n)
			err = osic.Send(&gnoios.InstallRequest{
				Request: &gnoios.InstallRequest_TransferContent{
					TransferContent: buf,
				},
			})
			if err != nil {
				return err
			}
		}
	}
	close(doneCh)
	a.Logger.Infof("target %q: sending TransferEnd", t.Config.Name)
	return osic.Send(gos.NewOSInstallTransferEnd())
}
