package main

import (
	"errors"
	"flag"
	"fmt"
	"github.com/cihub/seelog"
	"github.com/dropbox/dropbox-sdk-go-unofficial"
	"github.com/watermint/dreport/auth"
	"github.com/watermint/dreport/integration"
	"github.com/watermint/dreport/report"
	"log"
	"os"
	"strings"
)

var (
	AppVersion                    string = "dev"
	DropboxBusinessInfoAppKey     string
	DropboxBusinessInfoAppSecret  string
	DropboxBusinessFileAppKey     string
	DropboxBusinessFileAppSecret  string
	DropboxBusinessAuditAppKey    string
	DropboxBusinessAuditAppSecret string
)

const (
	seeLogXmlTemplate = `
	<seelog type="adaptive" mininterval="200000000" maxinterval="1000000000" critmsgcount="5">
	<formats>
    		<format id="detail" format="date:%%Date(2006-01-02T15:04:05Z07:00)%%tloc:%%File:%%FuncShort:%%Line%%tlevel:%%Level%%tmsg:%%Msg%%n" />
    		<format id="short" format="%%Date(2006-01-02T15:04:05Z07:00) [%%LEV] %%Msg%%n" />
	</formats>
	<outputs formatid="detail">
		<filter levels="info,warn,error,critical">
        		<console formatid="short" />
    		</filter>
    	</outputs>
	</seelog>
	`
)

func Authorise(ctx *integration.ExecutionContext, report report.Report) error {
	permissions := report.RequiredPermissions()
	seelog.Infof("Report requires following permission(s): %s\n", strings.Join(permissions, ","))
	seelog.Flush()

	for _, p := range permissions {
		switch p {
		case auth.PERMISSION_INFO:
			a := auth.DropboxAuthenticator{
				Permission: "Team Information",
				AppName:    ctx.AppName,
				AppKey:     ctx.TeamInfoAppKey,
				AppSecret:  ctx.TeamInfoAppSecret,
			}
			t, err := a.Authorise()
			if err != nil {
				seelog.Errorf("Unable to acquire token for '%s'", a.Permission)
				return err
			}
			ctx.TeamInfoToken = t

		case auth.PERMISSION_FILE:
			a := auth.DropboxAuthenticator{
				Permission: "Team file access",
				AppName:    ctx.AppName,
				AppKey:     ctx.TeamFileAppKey,
				AppSecret:  ctx.TeamFileAppSecret,
			}
			t, err := a.Authorise()
			if err != nil {
				seelog.Errorf("Unable to acquire token for '%s'", a.Permission)
				return err
			}
			ctx.TeamFileToken = t

		case auth.PERMISSION_AUDIT:
			a := auth.DropboxAuthenticator{
				Permission: "Team auditing",
				AppName:    ctx.AppName,
				AppKey:     ctx.TeamAuditAppKey,
				AppSecret:  ctx.TeamAuditAppSecret,
			}
			t, err := a.Authorise()
			if err != nil {
				seelog.Errorf("Unable to acquire token for '%s'", a.Permission)
				return err
			}
			ctx.TeamAuditToken = t

		}

	}

	return nil
}

func Revoke(ctx *integration.ExecutionContext) {
	if ctx.TeamInfoToken != "" {
		seelog.Info("Clean up token: Team Information")
		client := dropbox.Client(ctx.TeamInfoToken, dropbox.Options{})
		client.TokenRevoke()
	}
	if ctx.TeamFileToken != "" {
		seelog.Info("Clean up token: Team file access")
		client := dropbox.Client(ctx.TeamFileToken, dropbox.Options{})
		client.TokenRevoke()
	}
}

type Commands struct {
	SupportedReports []report.Report

	Report     report.Report
	ReportFile string
}

var (
	descReportName = "Report type name"
	descReportFile = "Output file path"
	descProxy      = "HTTP(S) proxy (hostname:port)"
)

func (o *Commands) Update() error {
	reportName := flag.String("report", "", descReportName)
	reportFile := flag.String("out", "", descReportFile)
	proxy := flag.String("proxy", "", descProxy)

	flag.Parse()

	if *reportFile == "" {
		flag.Usage()
		o.ShowSupportedReports()
		return errors.New("Required option: Output file path")
	}

	r, err := o.FindReport(*reportName)
	if r == nil || err != nil {
		if *reportName != "" {
			seelog.Errorf("Unsupported Report type: '%s'", *reportName)
		}

		flag.Usage()
		o.ShowSupportedReports()
		return err
	}
	o.ConfigureProxy(*proxy)

	o.Report = r
	o.ReportFile = *reportFile

	return nil
}

func (o *Commands) ShowSupportedReports() {
	fmt.Fprintln(os.Stderr, "")
	fmt.Fprintln(os.Stderr, "Supported Report types: ")
	fmt.Fprintln(os.Stderr, "")
	for _, r := range o.SupportedReports {
		fmt.Fprintln(os.Stderr, r.ReportName())
		fmt.Fprintf(os.Stderr, "    - %s\n", r.ReportDescription())
	}
}

func (o *Commands) FindReport(reportName string) (report.Report, error) {
	for _, r := range o.SupportedReports {
		if reportName == r.ReportName() {
			return r, nil
		}
	}
	return nil, errors.New("Unsupported Report type")
}

func (o *Commands) ConfigureProxy(proxy string) {
	if proxy != "" {
		seelog.Info("Explicit proxy configuration: HTTP_PROXY[%s]", proxy)
		seelog.Info("Explicit proxy configuration: HTTPS_PROXY[%s]", proxy)
		os.Setenv("HTTP_PROXY", proxy)
		os.Setenv("HTTPS_PROXY", proxy)
	}
}

func ConfigLogger() {
	logger, err := seelog.LoggerFromConfigAsString(fmt.Sprintf(seeLogXmlTemplate))
	if err != nil {
		log.Fatalln("Failed to load logger", err.Error())
	}
	seelog.ReplaceLogger(logger)
}

func main() {
	ConfigLogger()

	defer seelog.Flush()
	seelog.Info("dreport version: " + AppVersion)

	reports := []report.Report{
		&report.ReportMemberProfile{},
		&report.ReportQuotaUsage{},
		&report.ReportMemberSessions{},
	}
	cmd := Commands{
		SupportedReports: reports,
	}

	if err := cmd.Update(); err != nil {
		return
	}

	ctx := &integration.ExecutionContext{
		AppName:            "dreport",
		TeamInfoAppKey:     DropboxBusinessInfoAppKey,
		TeamInfoAppSecret:  DropboxBusinessInfoAppSecret,
		TeamFileAppKey:     DropboxBusinessFileAppKey,
		TeamFileAppSecret:  DropboxBusinessFileAppSecret,
		TeamAuditAppKey:    DropboxBusinessAuditAppKey,
		TeamAuditAppSecret: DropboxBusinessAuditAppSecret,
		OutputFile:         cmd.ReportFile,
	}

	if err := Authorise(ctx, cmd.Report); err != nil {
		seelog.Error("Unable to acquire enough authorisations.")
		return
	}
	defer Revoke(ctx)

	seelog.Info("Start report: ", cmd.Report.ReportName())
	if err := cmd.Report.Report(ctx); err != nil {
		seelog.Error(err)
	}
}
