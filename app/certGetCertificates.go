package app

import (
	"bytes"
	"context"
	"crypto/x509"
	"encoding/pem"
	"fmt"
	"os"
	"path/filepath"
	"sort"
	"strconv"
	"strings"
	"time"

	"github.com/olekukonko/tablewriter"
	"github.com/openconfig/gnoi/cert"
	"github.com/spf13/cobra"
	"github.com/spf13/pflag"
	"google.golang.org/grpc/metadata"
)

type getCertificatesResponse struct {
	TargetError
	rsp *cert.GetCertificatesResponse
}

func (a *App) InitCertGetCertificatesFlags(cmd *cobra.Command) {
	cmd.ResetFlags()
	//
	cmd.Flags().BoolVar(&a.Config.CertGetCertificatesDetails, "details", false, "print retrieved certificates details")
	cmd.Flags().StringSliceVar(&a.Config.CertGetCertificatesID, "id", []string{}, "certificate ID to be displayed")
	cmd.Flags().BoolVar(&a.Config.CertGetCertificatesSave, "save", false, "save retrieved certificates locally")
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
					TargetError: TargetError{
						TargetName: t.Config.Address,
						Err:        err,
					},
				}
				return
			}
			a.Logger.Debugf("%q gRPC client created", t.Config.Address)
			rsp, err := a.CertGetCertificates(ctx, t)
			responseChan <- &getCertificatesResponse{
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
	result := make([]*getCertificatesResponse, 0, numTargets)

	for rsp := range responseChan {
		if rsp.Err != nil {
			wErr := fmt.Errorf("%q Cert GetCertificates failed: %v", rsp.TargetName, rsp.Err)
			a.Logger.Error(wErr)
			errs = append(errs, wErr)
			continue
		}
		if a.Config.CertGetCertificatesSave {
			a.saveCerts(rsp)
		}
		result = append(result, rsp)
	}
	//
	if a.Config.CertGetCertificatesDetails {
		if len(result) == 0 {
			a.Logger.Warn("no certificates found")
			return nil
		}
		for _, rsp := range result {
			if rsp.rsp == nil || len(rsp.rsp.GetCertificateInfo()) == 0 {
				a.Logger.Warnf("%q no certificates found", rsp.TargetName)
				continue
			}
			for _, certInfo := range rsp.rsp.GetCertificateInfo() {
				// check name is in list
				if !sInListNotEmpty(certInfo.CertificateId, a.Config.CertGetCertificatesID) {
					continue
				}
				block, _ := pem.Decode(certInfo.Certificate.Certificate)
				cert, err := x509.ParseCertificate(block.Bytes)
				if err != nil {
					errs = append(errs, err)
					continue
				}

				certString, err := CertificateText(cert, false)
				if err != nil {
					errs = append(errs, err)
					continue
				}
				fmt.Printf("%q CertificateID: %s\n", rsp.TargetName, certInfo.CertificateId)
				fmt.Printf("%q Certificate Type: %s\n", rsp.TargetName, certInfo.Certificate.Type.String())
				fmt.Printf("%q Modification Time: %s\n", rsp.TargetName, time.Unix(0, certInfo.ModificationTime))
				fmt.Printf("%q %s\n", rsp.TargetName, certString)
			}
		}
	} else {
		rs, err := a.certTable(result)
		if err != nil {
			errs = append(errs, err)
		} else {
			fmt.Print(rs)
		}
	}
	return a.handleErrs(errs)
}

func (a *App) certTable(rsps []*getCertificatesResponse) (string, error) {
	tabData := make([][]string, 0)
	for _, rsp := range rsps {
		for _, certInfo := range rsp.rsp.GetCertificateInfo() {
			if !sInListNotEmpty(certInfo.CertificateId, a.Config.CertGetCertificatesID) {
				continue
			}
			block, _ := pem.Decode(certInfo.Certificate.Certificate)
			cert, err := x509.ParseCertificate(block.Bytes)
			if err != nil {
				return "", err
			}
			ipAddrs := make([]string, 0, len(cert.IPAddresses))
			for _, ipAddr := range cert.IPAddresses {
				ipAddrs = append(ipAddrs, ipAddr.String())
			}
			tabData = append(tabData, []string{
				rsp.TargetName,
				certInfo.CertificateId,
				time.Unix(0, certInfo.ModificationTime).Format(time.RFC3339),
				certInfo.GetCertificate().GetType().String(),
				strconv.Itoa(cert.Version),
				cert.Subject.ToRDNSequence().String(),
				// cert.Issuer.ToRDNSequence().String(),
				cert.NotBefore.Format(time.RFC3339),
				cert.NotAfter.Format(time.RFC3339),
				strings.Join(ipAddrs, ", "),
			})
		}
	}
	sort.Slice(tabData, func(i, j int) bool {
		return tabData[i][0] < tabData[j][0]
	})
	b := new(bytes.Buffer)
	table := tablewriter.NewWriter(b)
	table.SetHeader([]string{"Target Name", "ID", "Modification Time", "Type", "Version", "Subject", "Valid From", "Valid Until", "IP Addrs"})
	formatTable(table)
	table.AppendBulk(tabData)
	table.Render()
	return b.String(), nil
}

func (a *App) CertGetCertificates(ctx context.Context, t *Target) (*cert.GetCertificatesResponse, error) {
	return cert.NewCertificateManagementClient(t.client).GetCertificates(ctx, new(cert.GetCertificatesRequest))
}

func (a *App) saveCerts(rsp *getCertificatesResponse) {
	_, err := os.Stat(rsp.TargetName)
	if os.IsNotExist(err) {
		os.Mkdir(rsp.TargetName, os.ModeDir)
	}
	if rsp.rsp == nil || len(rsp.rsp.GetCertificateInfo()) == 0 {
		a.Logger.Warnf("%q no certificates found", rsp.TargetName)
		return
	}
	for _, certInfo := range rsp.rsp.GetCertificateInfo() {
		// check name is in list
		if !sInListNotEmpty(certInfo.CertificateId, a.Config.CertGetCertificatesID) {
			continue
		}
		f, err := os.Create(filepath.Join(rsp.TargetName, certInfo.CertificateId+".pem"))
		if err != nil {
			a.Logger.Warnf("%q cert=%q failed to create file: %v", rsp.TargetName, certInfo.CertificateId, err)
			continue
		}
		_, err = f.Write(certInfo.GetCertificate().GetCertificate())
		if err != nil {
			a.Logger.Warnf("%q cert=%q failed to write certificate file: %v", rsp.TargetName, certInfo.CertificateId, err)
		}
		f.Close()
	}
}
