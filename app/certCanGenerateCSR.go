package app

import (
	"bytes"
	"context"
	"fmt"
	"sort"

	"github.com/olekukonko/tablewriter"
	"github.com/openconfig/gnoi/cert"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
)

type certCGCSRResponse struct {
	TargetError
	can bool
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
		go func(t *Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = a.CreateGrpcClient(ctx, t, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &certCGCSRResponse{
					TargetError: TargetError{
						TargetName: t.Config.Address,
						Err:        err,
					},
				}
				return
			}

			can, err := a.CertCanGenerateCSR(ctx, t)
			responseChan <- &certCGCSRResponse{
				TargetError: TargetError{
					TargetName: t.Config.Address,
					Err:        err,
				},
				can: can,
			}
		}(t)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	result := make([]*certCGCSRResponse, 0, numTargets)
	for rsp := range responseChan {
		if rsp.Err != nil {
			a.Logger.Errorf("%q Cert CanGenerateCSR failed: %v", rsp.TargetName, rsp.Err)
			errs = append(errs, rsp.Err)
			continue
		}
		result = append(result, rsp)
	}
	fmt.Print(certCGCSRTable(result))
	return a.handleErrs(errs)
}

func (a *App) CertCanGenerateCSR(ctx context.Context, t *Target) (bool, error) {
	certClient := cert.NewCertificateManagementClient(t.client)
	resp, err := certClient.CanGenerateCSR(ctx,
		&cert.CanGenerateCSRRequest{
			KeyType:         cert.KeyType(cert.KeyType_value[a.Config.CertCanGenerateCSRKeyType]),
			CertificateType: cert.CertificateType(cert.CertificateType_value[a.Config.CertCanGenerateCSRCertificateType]),
			KeySize:         a.Config.CertCanGenerateCSRKeySize,
		})
	if err != nil {
		return false, err
	}
	a.Logger.Infof("%q key-type=%s, cert-type=%s, key-size=%d: can_generate: %v",
		t.Config.Address,
		a.Config.CertCanGenerateCSRKeyType,
		a.Config.CertCanGenerateCSRCertificateType,
		a.Config.CertCanGenerateCSRKeySize,
		resp.GetCanGenerate())
	return resp.GetCanGenerate(), nil
}

func certCGCSRTable(rsps []*certCGCSRResponse) string {
	tabData := make([][]string, 0, len(rsps))
	for _, rsp := range rsps {
		tabData = append(tabData, []string{
			rsp.TargetName,
			fmt.Sprintf("%t", rsp.can),
		})
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
