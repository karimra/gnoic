package app

import (
	"context"
	"fmt"
	"os"
	"path/filepath"

	"github.com/karimra/gnoic/api"
	gcert "github.com/karimra/gnoic/api/cert"
	"github.com/openconfig/gnoi/cert"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
)

type certGenCSRResponse struct {
	TargetError
	rsp *cert.GenerateCSRResponse
}

func (a *App) InitCertGenerateCSRFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().StringVar(&a.Config.CertGenerateCSRCertificateID, "id", "", "Certificate ID")
	cmd.Flags().StringVar(&a.Config.CertGenerateCSRKeyType, "key-type", "KT_RSA", "Key Type")
	cmd.Flags().StringVar(&a.Config.CertGenerateCSRCertificateType, "cert-type", "CT_X509", "Certificate Type")
	cmd.Flags().Uint32Var(&a.Config.CertGenerateCSRMinKeySize, "min-key-size", 1024, "Minimum Key Size")
	cmd.Flags().StringVar(&a.Config.CertGenerateCSRCommonName, "common-name", "", "CSR common name")
	cmd.Flags().StringVar(&a.Config.CertGenerateCSRCountry, "country", "", "CSR country")
	cmd.Flags().StringVar(&a.Config.CertGenerateCSRState, "state", "", "CSR state")
	cmd.Flags().StringVar(&a.Config.CertGenerateCSRCity, "city", "", "CSR city")
	cmd.Flags().StringVar(&a.Config.CertGenerateCSROrg, "org", "", "CSR organization")
	cmd.Flags().StringVar(&a.Config.CertGenerateCSROrgUnit, "org-unit", "", "CSR organization unit")
	cmd.Flags().StringVar(&a.Config.CertGenerateCSRIPAddress, "ip-address", "", "CSR IP address")
	cmd.Flags().StringVar(&a.Config.CertGenerateCSREmailID, "email-id", "", "CSR email ID")
	//
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunEGenerateCSR(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *certGenCSRResponse, numTargets)
	a.wg.Add(numTargets)
	for _, t := range targets {
		go a.certGenerateCSRRequest(cmd.Context(), t, responseChan)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	result := make([]*certGenCSRResponse, 0, numTargets)

	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q Cert CanGenerateCSR failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
		result = append(result, rsp)
	}
	for _, rsp := range result {
		err = a.saveCSR(rsp)
		if err != nil {
			errs = append(errs, err)
		}
	}
	return a.handleErrs(errs)
}

func (a *App) CertGenerateCSR(ctx context.Context, t *api.Target) (*cert.GenerateCSRResponse, error) {
	req, err := gcert.NewCertGenerateCSRRequest(
		gcert.CertificateID(a.Config.CertGenerateCSRCertificateID),
		gcert.CSRParams(
			gcert.CertificateType(a.Config.CertGenerateCSRCertificateType),
			gcert.MinKeySize(a.Config.CertGenerateCSRMinKeySize),
			gcert.KeyType(a.Config.CertGenerateCSRKeyType),
			gcert.CommonName(a.Config.CertGenerateCSRCommonName),
			gcert.Country(a.Config.CertGenerateCSRCountry),
			gcert.State(a.Config.CertGenerateCSRState),
			gcert.City(a.Config.CertGenerateCSRCity),
			gcert.Org(a.Config.CertGenerateCSROrg),
			gcert.OrgUnit(a.Config.CertGenerateCSROrgUnit),
			gcert.IPAddress(a.Config.CertGenerateCSRIPAddress),
			gcert.EmailID(a.Config.CertGenerateCSREmailID),
		),
	)
	if err != nil {
		return nil, err
	}
	a.printProtoMsg(t.Config.Name, req)
	certClient := t.CertClient()
	resp, err := certClient.GenerateCSR(ctx, req)
	if err != nil {
		return nil, err
	}
	a.printProtoMsg(t.Config.Name, resp)
	return resp, nil
}

func (a *App) saveCSR(rsp *certGenCSRResponse) error {
	certId := a.Config.CertGenerateCSRCertificateID

	if rsp.rsp == nil || rsp.rsp.GetCsr().GetCsr() == nil {
		return fmt.Errorf("%q cert=%q failed to get CSR from response", rsp.TargetName, certId)
	}
	_, err := os.Stat(rsp.TargetName)
	if os.IsNotExist(err) {
		os.MkdirAll(rsp.TargetName, 0755)
	}
	f, err := os.Create(filepath.Join(rsp.TargetName, certId+".csr"))
	if err != nil {
		a.Logger.Warnf("%q cert=%q failed to create file: %v", rsp.TargetName, certId, err)
		return err
	}
	defer f.Close()
	_, err = f.Write(rsp.rsp.GetCsr().GetCsr())
	if err != nil {
		a.Logger.Warnf("%q cert=%q failed to write certificate file: %v", rsp.TargetName, certId, err)
		return err
	}
	return nil
}

func (a *App) certGenerateCSRRequest(ctx context.Context, t *api.Target, rspCh chan<- *certGenCSRResponse) {
	defer a.wg.Done()
	ctx = t.AppendMetadata(ctx)
	ctx, cancel := context.WithCancel(ctx)
	defer cancel()

	err := t.CreateGrpcClient(ctx, a.createBaseDialOpts()...)
	if err != nil {
		rspCh <- &certGenCSRResponse{
			TargetError: TargetError{
				TargetName: t.Config.Address,
				Err:        err,
			},
		}
		return
	}
	defer t.Close()
	rsp, err := a.CertGenerateCSR(ctx, t)
	rspCh <- &certGenCSRResponse{
		TargetError: TargetError{
			TargetName: t.Config.Address,
			Err:        err,
		},
		rsp: rsp,
	}
}
