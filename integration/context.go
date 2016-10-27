package integration

import "github.com/watermint/dreport/publisher"

type ReportContext struct {
	// Auth Tokens
	TeamInfoToken  string
	TeamFileToken  string
	TeamAuditToken string

	// Output
	ReportOutput publisher.Publisher
}

type ApplicationContext struct {
	AppName string

	TeamInfoAppKey    string
	TeamInfoAppSecret string

	TeamFileAppKey    string
	TeamFileAppSecret string

	TeamAuditAppKey    string
	TeamAuditAppSecret string
}
