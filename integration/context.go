package integration

type ExecutionContext struct {
	AppName string

	// App keys and Tokens
	TeamInfoAppKey     string
	TeamInfoAppSecret  string
	TeamInfoToken      string
	TeamFileAppKey     string
	TeamFileAppSecret  string
	TeamFileToken      string
	TeamAuditAppKey    string
	TeamAuditAppSecret string
	TeamAuditToken     string

	// Output
	OutputFile string
}
