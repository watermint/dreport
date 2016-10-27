package report

import "github.com/watermint/dreport/integration"

type Report interface {
	ReportName() string
	ReportDescription() string
	RequiredPermissions() []string
	Report(context *integration.ReportContext) error
}
