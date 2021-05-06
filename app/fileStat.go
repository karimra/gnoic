package app

import (
	"bytes"
	"context"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/dustin/go-humanize"
	"github.com/olekukonko/tablewriter"
	"github.com/openconfig/gnoi/file"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/prototext"
)

type fileStatResponse struct {
	TargetError
	rsp *file.StatResponse
}

func (a *App) InitFileStatFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.FileStatFile, "file", "", "file to get from the target(s)")
	cmd.Flags().BoolVar(&a.Config.FileStatHumanize, "humanize", false, "make outputted values human readable")
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
					TargetError: TargetError{
						TargetName: t.Config.Address,
						Err:        err,
					},
				}
				return
			}
			rsp, err := a.FileStat(ctx, t)
			responseChan <- &fileStatResponse{
				TargetError: TargetError{
					TargetName: t.Config.Address,
					Err:        err,
				},
				rsp: rsp,
			}
		}(t)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	result := make([]*fileStatResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q File Stat failed: %v", rsp.TargetName, rsp.Err)
			errs = append(errs, wErr)
			continue
		}
		result = append(result, rsp)
	}

	fmt.Print(a.statTable(result))
	return a.handleErrs(errs)
}

func (a *App) FileStat(ctx context.Context, t *Target) (*file.StatResponse, error) {
	fileClient := file.NewFileClient(t.client)
	r, err := fileClient.Stat(ctx, &file.StatRequest{
		Path: a.Config.FileStatFile,
	})
	if err != nil {
		return nil, err
	}
	a.Logger.Debugf("File Stat Response: %s", prototext.Format(r))
	return r, nil
}

func (a *App) statTable(r []*fileStatResponse) string {
	tabData := make([][]string, 0)
	for _, rsp := range r {
		for _, si := range rsp.rsp.GetStats() {
			if a.Config.FileStatHumanize {
				tabData = append(tabData, []string{
					rsp.TargetName,
					si.GetPath(),
					humanize.Time(time.Unix(0, int64(si.GetLastModified()))),
					"0" + strconv.FormatUint(uint64(si.GetPermissions()), 8),
					"0" + strconv.FormatUint(uint64(si.GetUmask()), 8),
					humanize.Bytes(si.GetSize()),
				})
				continue
			}
			tabData = append(tabData, []string{
				rsp.TargetName,
				si.GetPath(),
				time.Unix(0, int64(si.GetLastModified())).Format(time.RFC3339),
				"0" + strconv.FormatUint(uint64(si.GetPermissions()), 8),
				"0" + strconv.FormatUint(uint64(si.GetUmask()), 8),
				strconv.Itoa(int((si.GetSize()))),
			})
		}
	}
	sort.Slice(tabData, func(i, j int) bool {
		return tabData[i][0] < tabData[j][0]
	})
	b := new(bytes.Buffer)
	table := tablewriter.NewWriter(b)
	table.SetHeader([]string{"Target Name", "Path", "LastModified", "Perm", "Umask", "Size"})
	formatTable(table)
	table.AppendBulk(tabData)
	table.Render()
	return b.String()
}
