package app

import (
	"bytes"
	"context"
	"fmt"
	"strings"

	"github.com/karimra/gnoic/api"
	"github.com/olekukonko/tablewriter"
	"github.com/openconfig/gnoi/common"
	"github.com/openconfig/gnoi/file"
	"github.com/openconfig/gnoi/types"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type fileTransferResponse struct {
	TargetError
	rsp *file.TransferToRemoteResponse
}

func (a *App) InitFileTransferFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.FileTransferLocal, "local", "", "path to local target file to be transferred")
	cmd.Flags().StringVar(&a.Config.FileTransferRemote, "remote", "", "remote path to transfer the local file to")
	cmd.Flags().StringVar(&a.Config.FileTransferSourceAddress, "source-address", "", "source address used to initiate connections from the target")
	//
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunEFileTransfer(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *fileTransferResponse, numTargets)
	a.wg.Add(numTargets)
	for _, t := range targets {
		go a.fileTransferRequest(cmd.Context(), t, responseChan)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	result := make([]*fileTransferResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q File Transfer failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
		result = append(result, rsp)
	}
	for _, r := range result {
		a.printProtoMsg(r.TargetName, r.rsp)
	}
	fmt.Print(a.transferTable(result))
	return a.handleErrs(errs)
}

func (a *App) fileTransferRequest(ctx context.Context, t *api.Target, rspCh chan<- *fileTransferResponse) {
	defer a.wg.Done()
	ctx = t.AppendMetadata(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
	if err != nil {
		rspCh <- &fileTransferResponse{
			TargetError: TargetError{
				TargetName: t.Config.Address,
				Err:        err,
			},
		}
		return
	}
	defer t.Close()
	rspCh <- a.FileTransfer(ctx, t)
}

func (a *App) FileTransfer(ctx context.Context, t *api.Target) *fileTransferResponse {
	rd, err := a.transferFileRemoteDownload()
	if err != nil {
		return &fileTransferResponse{
			TargetError: TargetError{
				TargetName: t.Config.Name,
				Err:        err,
			},
		}
	}
	req := &file.TransferToRemoteRequest{
		LocalPath:      a.Config.FileTransferLocal,
		RemoteDownload: rd,
	}
	fileClient := t.FileClient()
	a.Logger.Infof("sending file transfer request: %v to target %q", req, t.Config.Name)
	rsp, err := fileClient.TransferToRemote(ctx, req)
	return &fileTransferResponse{
		TargetError: TargetError{
			TargetName: t.Config.Name,
			Err:        err,
		},
		rsp: rsp,
	}
}

func (a *App) transferTable(r []*fileTransferResponse) string {
	targetTabData := make([][]string, 0)
	for _, rsp := range r {
		targetTabData = append(targetTabData, []string{rsp.TargetName, rsp.rsp.Hash.Method.String(), fmt.Sprintf("%x", rsp.rsp.Hash.Hash)})
	}
	b := new(bytes.Buffer)
	table := tablewriter.NewWriter(b)
	table.SetHeader([]string{"Target Name", "Hash Method", "Hash"})
	formatTable(table)
	table.AppendBulk(targetTabData)
	table.Render()
	return b.String()
}

// scp://user:pass@server.com:/path/to/file
// sftp://user:pass@server.com:/path/to/file
// http(s)://user:pass@server.com/path/to/file
func (a *App) transferFileRemoteDownload() (*common.RemoteDownload, error) {
	path := a.Config.FileTransferRemote
	rd := new(common.RemoteDownload)
	switch {
	case strings.HasPrefix(path, "https://"):
		rd.Protocol = common.RemoteDownload_HTTPS
		path = strings.TrimPrefix(path, "https://")
	case strings.HasPrefix(path, "http://"):
		rd.Protocol = common.RemoteDownload_HTTP
		path = strings.TrimPrefix(path, "http://")
	case strings.HasPrefix(path, "scp://"):
		rd.Protocol = common.RemoteDownload_SCP
		path = strings.TrimPrefix(path, "scp://")
	case strings.HasPrefix(path, "sftp://"):
		rd.Protocol = common.RemoteDownload_SFTP
		path = strings.TrimPrefix(path, "sftp://")
	default:
		return nil, fmt.Errorf("missing protocol in remote path %q", path)
	}
	if strings.Contains(path, "@") {
		rd.Credentials = &types.Credentials{}
		pathElems := strings.SplitN(path, "@", 2)
		credentials := strings.SplitN(pathElems[0], ":", 2)
		if len(credentials) != 2 {
			return nil, fmt.Errorf("failed to parse remote path credentials %q", a.Config.FileTransferRemote)
		}
		rd.Credentials.Username = credentials[0]
		rd.Credentials.Password = &types.Credentials_Cleartext{Cleartext: credentials[1]}
		path = strings.TrimPrefix(path, fmt.Sprintf("%s:%s@", credentials[0], credentials[1]))
	}
	rd.Path = path
	rd.SourceAddress = a.Config.FileTransferSourceAddress
	return rd, nil
}
