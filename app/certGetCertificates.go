package app

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"sort"
	"strconv"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/openconfig/gnoi/cert"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
)

type getCertificatesResponse struct {
	targetName string
	rsp        *cert.GetCertificatesResponse
	err        error
}

func (a *App) InitCertGetCertificatesFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().BoolVar(&a.Config.CertGetCertificatesDetails, "details", false, "print certificates details")
	//
	cmd.LocalFlags().VisitAll(func(flag *pflag.Flag) {
		a.Config.FileConfig.BindPFlag(fmt.Sprintf("%s-%s", cmd.Name(), flag.Name), flag)
	})
}

func (a *App) RunECertGetCertificates(cmd *cobra.Command, args []string) error {
	targets, err := a.GetTargets()
	if err != nil {
		return err
	}

	numTargets := len(targets)
	responseChan := make(chan *getCertificatesResponse, numTargets)

	a.wg.Add(numTargets)
	for _, t := range targets {
		go func(t *Target) {
			defer a.wg.Done()
			ctx, cancel := context.WithCancel(a.ctx)
			defer cancel()
			ctx = metadata.AppendToOutgoingContext(ctx, "username", *t.Config.Username, "password", *t.Config.Password)

			err = a.CreateGrpcClient(ctx, t, a.createBaseDialOpts()...)
			if err != nil {
				responseChan <- &getCertificatesResponse{
					targetName: t.Config.Address,
					rsp:        nil,
					err:        err,
				}
				return
			}
			a.Logger.Debugf("%q gRPC client created", t.Config.Address)
			rsp, err := a.CertGetCertificates(ctx, t)
			responseChan <- &getCertificatesResponse{
				targetName: t.Config.Address,
				rsp:        rsp,
				err:        err,
			}
		}(t)
	}
	a.wg.Wait()
	close(responseChan)

	errs := make([]error, 0, numTargets)
	result := make([]*getCertificatesResponse, 0, numTargets)

	for rsp := range responseChan {
		if rsp.err != nil {
			a.Logger.Errorf("%q get certificates failed: %v", rsp.targetName, err)
			errs = append(errs, err)
			continue
		}
		result = append(result, rsp)
	}
	//
	for _, err := range errs {
		a.Logger.Errorf("err: %v", err)
	}
	if a.Config.CertGetCertificatesDetails {
		if len(result) == 0 {
			fmt.Printf("no certificates found\n")
			return nil
		}
		for _, rsp := range result {
			if rsp.rsp == nil || len(rsp.rsp.GetCertificateInfo()) == 0 {
				fmt.Printf("%q no certificates found\n", rsp.targetName)
				continue
			}
			for _, certInfo := range rsp.rsp.CertificateInfo {
				block, _ := pem.Decode(certInfo.Certificate.Certificate)
				cert, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					return err
				}

				fmt.Printf("%q CertificateID: %s\n", rsp.targetName, certInfo.CertificateId)
				fmt.Printf("%q Certificate Type: %s\n", rsp.targetName, certInfo.Certificate.Type.String())
				fmt.Printf("%q Modification Time: %s\n", rsp.targetName, time.Unix(0, certInfo.ModificationTime))

				certString, err := CertificateText(cert)
				if err != nil {
					return err
				}
				fmt.Printf("%q %s\n", rsp.targetName, certString)
			}
		}
	} else {
		rs, err := certTable(result)
		if err != nil {
			return err
		}
		fmt.Print(rs)
	}
	numErrors := len(errs)
	if numErrors > 0 {
		return fmt.Errorf("there was %d error(s)", numErrors)
	}
	a.Logger.Debug("done...")
	return nil
}

func certTable(rsps []*getCertificatesResponse) (string, error) {
	tabData := make([][]string, 0)
	for _, rsp := range rsps {
		for _, certInfo := range rsp.rsp.GetCertificateInfo() {
			block, _ := pem.Decode(certInfo.Certificate.Certificate)
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return "", err
			}
			tabData = append(tabData, []string{
				rsp.targetName,
				certInfo.CertificateId,
				time.Unix(0, certInfo.ModificationTime).String(),
				certInfo.GetCertificate().GetType().String(),
				strconv.Itoa(cert.Version),
				cert.Subject.ToRDNSequence().String(),
				cert.Issuer.ToRDNSequence().String(),
			})
		}
	}
	sort.Slice(tabData, func(i, j int) bool {
		return tabData[i][0] < tabData[j][0]
	})
	b := new(bytes.Buffer)
	table := tablewriter.NewWriter(b)
	table.SetHeader([]string{"Target Name", "ID", "Modification Time", "Type", "Version", "Subject", "Issuer"})
	table.SetAlignment(tablewriter.ALIGN_LEFT)
	table.SetAutoFormatHeaders(false)
	table.SetAutoWrapText(false)
	table.AppendBulk(tabData)
	table.Render()
	return b.String(), nil
}

func (a *App) CertGetCertificates(ctx context.Context, t *Target) (*cert.GetCertificatesResponse, error) {
	certClient := cert.NewCertificateManagementClient(t.client)
	return certClient.GetCertificates(ctx, new(cert.GetCertificatesRequest))
}
