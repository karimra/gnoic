package app

import (
	"bytes"
	"context"
	"fmt"
	"os"

	"github.com/openconfig/gnoi/healthz"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
	"google.golang.org/protobuf/encoding/prototext"

	"github.com/karimra/gnoic/api"
	ghealthz "github.com/karimra/gnoic/api/healthz"
)

type healthzArtifactResponse struct {
	TargetError
	rsp *healthz.ArtifactResponse
}

func (a *App) InitHealthzArtifactFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.HealthzArtifactID, "id", "", "artifact ID")
	//
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunEHealthzArtifact(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *healthzArtifactResponse, numTargets)

	a.wg.Add(numTargets)
	for _, t := range targets {
		go func(t *api.Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &healthzArtifactResponse{
					TargetError: TargetError{
						TargetName: t.Config.Address,
						Err:        err,
					},
				}
				return
			}
			defer t.Close()
			responseChan <- a.HealthArtifact(ctx, t)
		}(t)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	result := make([]*healthzArtifactResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q Healthz Artifact failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
		result = append(result, rsp)
	}

	for _, r := range result {
		a.printMsg(r.TargetName, r.rsp)
	}
	return a.handleErrs(errs)
}

func (a *App) HealthArtifact(ctx context.Context, t *api.Target) *healthzArtifactResponse {
	opts := []ghealthz.HealthzOption{
		ghealthz.ID(a.Config.HealthzArtifactID),
	}
	req, err := ghealthz.NewArtifactRequest(opts...)
	if err != nil {
		return &healthzArtifactResponse{
			TargetError: TargetError{
				TargetName: t.Config.Name,
				Err:        err,
			},
		}
	}
	a.printMsg(t.Config.Name, req)
	hc := healthz.NewHealthzClient(t.Conn())
	artifactStream, err := hc.Artifact(ctx, req)
	if err != nil {
		return &healthzArtifactResponse{
			TargetError: TargetError{
				TargetName: t.Config.Name,
				Err:        err,
			},
		}
	}
	// rcv header
	rsp, err := artifactStream.Recv()
	if err != nil {
		return &healthzArtifactResponse{
			TargetError: TargetError{
				TargetName: t.Config.Name,
				Err:        err,
			},
		}
	}
	switch content := rsp.GetContents().(type) {
	case *healthz.ArtifactResponse_Header:
		switch content.Header.GetArtifactType().(type) {
		case *healthz.ArtifactHeader_File:
			err = a.handleFileArtifact(t.Config.Name, content, artifactStream)
		case *healthz.ArtifactHeader_Custom:
			err = a.handleCustomArtifact(t.Config.Name, content, artifactStream)
		case *healthz.ArtifactHeader_Proto:
			err = a.handleProtoArtifact(t.Config.Name, content, artifactStream)
		}
		if err != nil {
			return &healthzArtifactResponse{
				TargetError: TargetError{
					TargetName: t.Config.Name,
					Err:        err,
				},
			}
		}
	default:
		return &healthzArtifactResponse{
			TargetError: TargetError{
				TargetName: t.Config.Name,
				Err:        fmt.Errorf("unexpected message type, expecting ArtifactResponse_Header, got %T", rsp.GetContents()),
			},
		}
	}
	return &healthzArtifactResponse{
		TargetError: TargetError{
			TargetName: t.Config.Name,
		},
	}
}

func (a *App) handleFileArtifact(targetName string, h *healthz.ArtifactResponse_Header, stream healthz.Healthz_ArtifactClient) error {
	id := h.Header.GetId()
	log.Infof("%s: received file header for artifactID: %s", targetName, id)
	fmt.Println(prototext.Format(h.Header))

	b := new(bytes.Buffer)
	for {
		rsp, err := stream.Recv()
		if err != nil {
			return err
		}
		switch content := rsp.GetContents().(type) {
		case *healthz.ArtifactResponse_Trailer:
			log.Infof("%s: received trailer for artifactID: %s", targetName, id)
			log.Infof("%s: received %d bytes in total", targetName, b.Len())
			log.Infof("%s: comparing file HASH", targetName)
			err = a.compareFileHash(targetName, b, h.Header.GetFile().GetHash())
			if err != nil {
				return fmt.Errorf("%s: hash err: %v", targetName, err)
			}
			log.Infof("%s: HASH OK", targetName)
			fi, err := os.Create(h.Header.GetFile().GetName())
			if err != nil {
				return err
			}
			defer fi.Close()
			_, err = fi.Write(b.Bytes())
			return err
		case *healthz.ArtifactResponse_Bytes:
			log.Infof("received %d bytes for artifactID: %s", len(content.Bytes), id)
			_, err = b.Write(content.Bytes)
			if err != nil {
				return err
			}
		default:
			return fmt.Errorf("unexpected message type when receiving a File Artifact, got :%T", rsp.GetContents())
		}
	}
}

func (a *App) handleCustomArtifact(targetName string, h *healthz.ArtifactResponse_Header, _ healthz.Healthz_ArtifactClient) error {
	id := h.Header.GetId()
	log.Infof("%s: received custom header for artifactID: %s", targetName, id)
	fmt.Println(prototext.Format(h.Header))
	//
	return nil
}

func (a *App) handleProtoArtifact(targetName string, h *healthz.ArtifactResponse_Header, _ healthz.Healthz_ArtifactClient) error {
	id := h.Header.GetId()
	log.Infof("%s: received proto header for artifactID: %s", targetName, id)
	fmt.Println(prototext.Format(h.Header))
	//
	return nil
}
