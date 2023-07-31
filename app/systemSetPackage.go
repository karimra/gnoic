package app

import (
	"context"
	"crypto/sha512"
	"errors"
	"fmt"
	"io"
	"os"

	gsystem "github.com/karimra/gnoic/api/system"
	"github.com/openconfig/gnoi/system"
	gnoisystem "github.com/openconfig/gnoi/system"

	"github.com/karimra/gnoic/api"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
)

type setPackageResponse struct {
	TargetError
}

func (a *App) InitSystemSetPackageFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.SystemSetPackageFile, "pkg", "", "file to put on the target(s)")
	cmd.Flags().StringVar(&a.Config.SystemSetPackageDstFile, "dst", "", "path and filename on the target(s)")
	cmd.Flags().StringVar(&a.Config.SystemSetPackageVersion, "version", "", "package version")
	cmd.Flags().BoolVar(&a.Config.SystemSetPackageActivate, "activate", false, "make package active")
	cmd.Flags().StringVar(&a.Config.SystemSetPackageRemoteFile, "remote", "", "path to the package for remote download")
	cmd.Flags().Uint64Var(&a.Config.SystemSetPackageChunkSize, "content-chunk-size", defaultChunkSize, "max chunk size to transfer the package")
	//
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) PreRunESetPackage(cmd *cobra.Command, args []string) error {
	a.Config.SetLocalFlagsFromFile(cmd)

	if a.Config.SystemSetPackageFile == "" {
		return errors.New("missing --pkg flag")
	}

	if a.Config.SystemSetPackageVersion == "" {
		return errors.New("missing --version flag")
	}

	if a.Config.SystemSetPackageDstFile == "" {
		return errors.New("missing --dst flag")
	}

	if a.Config.SystemSetPackageRemoteFile != "" {
		return errors.New("remote is not implemented")
	}

	return nil
}

func (a *App) RunESystemSetPackage(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *setPackageResponse, numTargets)

	a.wg.Add(numTargets)

	for _, t := range targets {
		go func(t *api.Target) {
			defer a.wg.Done()

			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()

			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &setPackageResponse{
					TargetError: TargetError{
						TargetName: t.Config.Name,
						Err:        err,
					},
				}
				return
			}

			defer t.Close()

			err = a.SystemSetPackage(ctx, t)
			responseChan <- &setPackageResponse{
				TargetError: TargetError{
					TargetName: t.Config.Name,
					Err:        err,
				},
			}
		}(t)
	}

	a.wg.Wait()

	close(responseChan)

	errs := make([]error, 0, numTargets)

	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q SetPackage failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
		a.Logger.Infof("%q package %s sent successfully", rsp.TargetName, a.Config.SystemSetPackageFile)
	}

	return a.handleErrs(errs)
}

func (a *App) SystemSetPackage(ctx context.Context, t *api.Target) error {
	sysc := gnoisystem.NewSystemClient(t.Conn())
	sysSetPackageClient, err := sysc.SetPackage(ctx)
	if err != nil {
		return err
	}

	a.Logger.Infof("target %q: starting SetPackage stream", t.Config.Name)

	filename := a.Config.SystemSetPackageFile
	_, err = os.Stat(filename)
	if err != nil {
		return fmt.Errorf("file %q stat err: %v", filename, err)
	}

	err = a.sendSysPackageFile(filename, a.Config.SystemSetPackageDstFile, sysSetPackageClient, t)
	if err != nil {
		return err
	}

	return nil
}

func (a *App) sendSysPackageFile(fileName, remoteFile string, sysClient system.System_SetPackageClient, t *api.Target) error {
	f, err := os.Open(fileName)
	if err != nil {
		a.Logger.Errorf("failed opening file %q: %v", fileName, err)

		return err
	}

	defer f.Close()

	a.Logger.Infof("%q sending file=%q", t.Config.Address, fileName)

	req, err := gsystem.NewSetPackagePackageRequest(
		gsystem.PackageFile(remoteFile),
		gsystem.Version(a.Config.SystemSetPackageVersion),
		gsystem.Activate(a.Config.SystemSetPackageActivate),
	)
	if err != nil {
		return err
	}

	a.printMsg(t.Config.Name, req)
	err = sysClient.Send(req)
	if err != nil {
		return err
	}

	h := sha512.New()

	for {
		b := make([]byte, a.Config.SystemSetPackageChunkSize)

		n, err := f.Read(b)
		if err != nil && err != io.EOF {
			return err
		}
		if err == io.EOF || n == 0 {
			break
		}

		h.Write(b[:n])

		a.Logger.Debugf("%q file=%q, writing %d byte(s)", t.Config.Address, fileName, n)

		reqContents := &gnoisystem.SetPackageRequest{
			Request: &gnoisystem.SetPackageRequest_Contents{
				Contents: b[:n],
			},
		}

		err = sysClient.Send(reqContents)
		if err != nil {
			return err
		}
	}

	// send hash
	a.Logger.Infof("%q sending file=%q hash", t.Config.Address, fileName)

	reqHash, err := gsystem.NewSetPackageHashRequest(
		gsystem.Hash("SHA512", h.Sum(nil)),
	)
	if err != nil {
		return err
	}

	a.printMsg(t.Config.Name, reqHash)

	err = sysClient.Send(reqHash)
	if err != nil {
		return err
	}

	rsp, err := sysClient.CloseAndRecv()
	a.printMsg(t.Config.Name, rsp)
	return err
}
