package app

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"strings"

	"github.com/openconfig/gnoi/healthz"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/prototext"

	"github.com/karimra/gnoic/api"
	ghealthz "github.com/karimra/gnoic/api/healthz"
	"github.com/karimra/gnoic/utils"
)

type healthzGetResponse struct {
	TargetError
	rsp *healthz.GetResponse
}

func (a *App) InitHealthzFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) InitHealthzGetFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.HealthzGetPath, "path", "", "path to the component to try to fetch healthz state for")
	//
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunEHealthzGet(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *healthzGetResponse, numTargets)

	a.wg.Add(numTargets)
	for _, t := range targets {
		go func(t *api.Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &healthzGetResponse{
					TargetError: TargetError{
						TargetName: t.Config.Address,
						Err:        err,
					},
				}
				return
			}
			defer t.Close()
			responseChan <- a.HealthzGet(ctx, t)
		}(t)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	result := make([]*healthzGetResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q Healthz Get failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
		result = append(result, rsp)
	}

	for _, r := range result {
		fmt.Printf("target %q:\n", r.TargetName)
		a.printMsg(r.TargetName, r.rsp)
		switch a.Config.Format {
		case "json":
			b, err := json.MarshalIndent(r.rsp, "", "  ")
			if err != nil {
				return fmt.Errorf("failed to marshal to JSON: %v", err)
			}
			fmt.Println(string(b))
		default:
			fmt.Println(a.healthzGetTree(r.rsp.GetComponent(), "  "))
		}
	}
	return a.handleErrs(errs)
}

func (a *App) HealthzGet(ctx context.Context, t *api.Target) *healthzGetResponse {
	opts := []ghealthz.HealthzOption{
		ghealthz.Path(a.Config.HealthzGetPath),
	}
	req, err := ghealthz.NewGetRequest(opts...)
	if err != nil {
		return &healthzGetResponse{
			TargetError: TargetError{
				TargetName: t.Config.Name,
				Err:        err,
			},
		}
	}
	a.printMsg(t.Config.Name, req)
	hc := healthz.NewHealthzClient(t.Conn())
	rsp, err := hc.Get(ctx, req)
	return &healthzGetResponse{
		TargetError: TargetError{
			TargetName: t.Config.Name,
			Err:        err,
		},
		rsp: rsp,
	}
}

func (a *App) healthzGetTree(comp *healthz.ComponentStatus, indent string) string {
	b := new(strings.Builder)
	fmt.Fprintf(b, "%spath     : %s\n", indent, utils.PathToXPath(comp.GetPath()))
	fmt.Fprintf(b, "%sstatus   : %s\n", indent, comp.GetStatus().String())
	fmt.Fprintf(b, "%sid       : %s\n", indent, comp.GetId())
	fmt.Fprintf(b, "%sacked    : %t\n", indent, comp.GetAcknowledged())
	fmt.Fprintf(b, "%screated  : %s\n", indent, comp.GetCreated().AsTime())
	fmt.Fprintf(b, "%sexpires  : %s\n", indent, comp.GetExpires().AsTime())

	if len(comp.GetArtifacts()) > 0 {
		fmt.Fprintf(b, "%sartifict :\n", indent)
		for _, art := range comp.GetArtifacts() {
			b.Write(printArtifactType(indent, art))
		}
	}

	if len(comp.GetSubcomponents()) > 0 {
		fmt.Fprintf(b, "%ssubcomponents:\n", indent)
		for _, sc := range comp.GetSubcomponents() {
			b.WriteString(a.healthzGetTree(sc, indent+"  "))
		}
	}
	return b.String()
}
func artifactType(arth *healthz.ArtifactHeader) string {
	switch arth.GetArtifactType().(type) {
	case *healthz.ArtifactHeader_File:
		return "file"
	case *healthz.ArtifactHeader_Custom:
		return "custom"
	case *healthz.ArtifactHeader_Proto:
		return "proto"
	}
	return ""
}

func printArtifactType(indent string, arth *healthz.ArtifactHeader) []byte {
	sb := new(bytes.Buffer)
	fmt.Fprintf(sb, "%s  - id       : %s\n", indent, arth.Id)
	switch arth := arth.GetArtifactType().(type) {
	case *healthz.ArtifactHeader_File:
		fmt.Fprintf(sb, "%s    name     : %s\n", indent, arth.File.GetName())
		fmt.Fprintf(sb, "%s    path     : %s\n", indent, arth.File.GetPath())
		fmt.Fprintf(sb, "%s    mimeType : %s\n", indent, arth.File.GetMimetype())
		fmt.Fprintf(sb, "%s    size     : %d\n", indent, arth.File.GetSize())
		fmt.Fprintf(sb, "%s    hash     : %s(%x)\n", indent, arth.File.GetHash().GetMethod(), arth.File.GetHash().GetHash())
	case *healthz.ArtifactHeader_Custom:
		fmt.Fprintf(sb, "%s    typeURL : %s\n", indent, arth.Custom.GetTypeUrl())
		fmt.Fprintf(sb, "%s    value   : %x\n", indent, arth.Custom.GetValue())
	case *healthz.ArtifactHeader_Proto:
		fmt.Fprintf(sb, "%s    proto :\n%s\n", indent, prototext.Format(arth.Proto))
	}
	return sb.Bytes()
}
