package app

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/openconfig/gnoi/file"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
)

type fileStatResponse struct {
	targetName string
	rsp        *file.StatResponse
	err        error
}

func (a *App) InitFileStatFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.FileStatFile, "file", "", "file to get from the target(s)")
	//
	cmd.Flags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunEFileStat(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *fileStatResponse, numTargets)

	a.wg.Add(numTargets)
	for _, t := range targets {
		go func(t *Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = a.CreateGrpcClient(ctx, t, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &fileStatResponse{
					targetName: t.Config.Address,
					err:        err,
				}
				return
			}
			rsp, err := a.FileStat(ctx, t)
			responseChan <- &fileStatResponse{
				targetName: t.Config.Address,
				rsp:        rsp,
				err:        err,
			}
		}(t)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	result := make([]*fileStatResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.err != nil {
			a.Logger.Errorf("%q file Stat failed: %v", rsp.targetName, rsp.err)
			errs = append(errs, rsp.err)
			continue
		}
		result = append(result, rsp)
	}

	for _, err := range errs {
		a.Logger.Errorf("err: %v", err)
	}

	fmt.Print(statTable(result))

	//
	if len(errs) > 0 {
		return fmt.Errorf("there was %d errors", len(errs))
	}
	a.Logger.Debug("done...")
	return nil
}

func (a *App) FileStat(ctx context.Context, t *Target) (*file.StatResponse, error) {
	fileClient := file.NewFileClient(t.client)
	return fileClient.Stat(ctx, &file.StatRequest{
		Path: a.Config.FileStatFile,
	})
}

func statTable(r []*fileStatResponse) string {
	tabData := make([][]string, 0)
	for _, rsp := range r {
		for _, si := range rsp.rsp.GetStats() {

			tabData = append(tabData, []string{
				rsp.targetName,
				si.GetPath(),
				time.Unix(0, int64(si.GetLastModified())).String(),
				strconv.Itoa(int(si.GetPermissions())),
				strconv.Itoa(int(si.GetUmask())),
				strconv.Itoa(int(si.GetSize())),
			})
		}
	}
	sort.Slice(tabData, func(i, j int) bool {
		return tabData[i][0] < tabData[j][0]
	})
	b := new(bytes.Buffer)
	table := tablewriter.NewWriter(b)
	table.SetHeader([]string{"Target Name", "Path", "LastModified", "Permissions", "Umask", "Size"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.AppendBulk(tabData)
	table.Render()
	return b.String()
}
