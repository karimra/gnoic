package app

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"github.com/karimra/gnoic/api"
	gcert "github.com/karimra/gnoic/api/cert"
	"github.com/olekukonko/tablewriter"
	"github.com/openconfig/gnoi/cert"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type certCGCSRResponse struct {
	TargetError
	rsp *cert.CanGenerateCSRResponse
}

func (r *certCGCSRResponse) Target() string {
	return r.TargetName
}

func (r *certCGCSRResponse) Response() any {
	return r.rsp
}

func (a *App) InitCertCanGenerateCSRFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.CertCanGenerateCSRKeyType, "key-type", "KT_RSA", "Key Type")
	cmd.Flags().StringVar(&a.Config.CertCanGenerateCSRCertificateType, "cert-type", "CT_X509", "Certificate Type")
	cmd.Flags().Uint32Var(&a.Config.CertCanGenerateCSRKeySize, "key-size", 2048, "Key Size")
	//
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunECertCanGenerateCSR(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *certCGCSRResponse, numTargets)

	a.wg.Add(numTargets)
	for _, t := range targets {
		go a.certCanGenerateCSRRequest(cmd.Context(), t, responseChan)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	result := make([]TargetResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q Cert CanGenerateCSR failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, rsp.Err)
			continue
		}
		result = append(result, rsp)
		a.printProtoMsg(rsp.TargetName, rsp.rsp)
	}
	a.printCMDOutput(result, a.certCGCSRTable)
	return a.handleErrs(errs)
}

func (a *App) certCanGenerateCSRRequest(ctx context.Context, t *api.Target, rspCh chan<- *certCGCSRResponse) {
	defer a.wg.Done()
	ctx = t.AppendMetadata(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
	if err != nil {
		rspCh <- &certCGCSRResponse{
			TargetError: TargetError{
				TargetName: t.Config.Address,
				Err:        err,
			},
		}
		return
	}
	defer t.Close()
	rspCh <- a.certCanGenerateCSR(ctx, t)
}

func (a *App) certCanGenerateCSR(ctx context.Context, t *api.Target) *certCGCSRResponse {
	req, err := gcert.NewCertCanGenerateCSRRequest(
		gcert.CertificateType(a.Config.CertCanGenerateCSRCertificateType),
		gcert.KeyType(a.Config.CertCanGenerateCSRKeyType),
		gcert.KeySize(a.Config.CertCanGenerateCSRKeySize),
	)
	if err != nil {
		return &certCGCSRResponse{
			TargetError: TargetError{
				TargetName: t.Config.Name,
				Err:        err,
			},
		}
	}

	a.printProtoMsg(t.Config.Name, req)
	certClient := t.CertClient()
	resp, err := certClient.CanGenerateCSR(ctx, req)
	if err != nil {
		return &certCGCSRResponse{
			TargetError: TargetError{
				TargetName: t.Config.Name,
				Err:        err,
			},
		}
	}

	a.printProtoMsg(t.Config.Name, resp)
	a.Logger.Infof("%q key-type=%s, cert-type=%s, key-size=%d: can_generate: %v",
		t.Config.Address,
		a.Config.CertCanGenerateCSRKeyType,
		a.Config.CertCanGenerateCSRCertificateType,
		a.Config.CertCanGenerateCSRKeySize,
		resp.GetCanGenerate())
	return &certCGCSRResponse{
		TargetError: TargetError{
			TargetName: t.Config.Name,
			Err:        err,
		},
		rsp: resp,
	}
}

func (a *App) certCGCSRTable(rsps []TargetResponse) string {
	tabData := make([][]string, 0, len(rsps))
	sort.Slice(rsps, func(i, j int) bool {
		return rsps[i].Target() < rsps[j].Target()
	})
	for _, rsp := range rsps {
		switch r := rsp.Response().(type) {
		case *cert.CanGenerateCSRResponse:
			tabData = append(tabData, []string{
				rsp.Target(),
				fmt.Sprintf("%t", r.GetCanGenerate()),
			})
		default:
			a.Logger.Printf("%s: unexpected message type: %T", rsp.Target(), rsp.Response())
		}
	}
	sort.Slice(tabData, func(i, j int) bool {
		return tabData[i][0] < tabData[j][0]
	})
	b := new(bytes.Buffer)
	table := tablewriter.NewWriter(b)
	table.SetHeader([]string{"Target Name", "Can Generate CSR"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.AppendBulk(tabData)
	table.Render()
	return b.String()
}
