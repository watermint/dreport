package auth

import (
	"fmt"
	"github.com/cihub/seelog"
	"github.com/satori/go.uuid"
	"golang.org/x/oauth2"
)

const (
	PERMISSION_INFO  = "info"
	PERMISSION_FILE  = "file"
	PERMISSION_AUDIT = "audit"
)

type DropboxAuthenticator struct {
	Permission string
	AppName    string
	AppKey     string
	AppSecret  string
}

func (d *DropboxAuthenticator) Authorise() (string, error) {
	state := uuid.NewV4().String()

	tok, err := d.auth(state)
	if err != nil {
		fmt.Errorf("Err: %s\n", err)
		return "", err
	}
	return tok.AccessToken, nil
}

func (d *DropboxAuthenticator) authEndpoint() *oauth2.Endpoint {
	return &oauth2.Endpoint{
		AuthURL:  "https://www.dropbox.com/oauth2/authorize",
		TokenURL: "https://api.dropboxapi.com/oauth2/token",
	}
}

func (d *DropboxAuthenticator) authConfig() *oauth2.Config {
	return &oauth2.Config{
		ClientID:     d.AppKey,
		ClientSecret: d.AppSecret,
		Scopes:       []string{},
		Endpoint:     *d.authEndpoint(),
	}
}

func (d *DropboxAuthenticator) authUrl(cfg *oauth2.Config, state string) string {
	return cfg.AuthCodeURL(
		state,
		oauth2.SetAuthURLParam("response_type", "code"),
	)
}

func (d *DropboxAuthenticator) authExchange(cfg *oauth2.Config, code string) (*oauth2.Token, error) {
	return cfg.Exchange(oauth2.NoContext, code)
}

func (d *DropboxAuthenticator) codeDialogue(state string) string {
	var code string

	fmt.Print("Enter the authorisation code here: ")

	if _, err := fmt.Scan(&code); err != nil {
		fmt.Errorf("%s\n", err)
		return ""
	}
	return code
}

func (d *DropboxAuthenticator) auth(state string) (*oauth2.Token, error) {
	cfg := d.authConfig()
	url := d.authUrl(cfg, state)

	seelog.Flush()
	fmt.Println("=====================")
	fmt.Printf("Authorise application '%s' with '%s' permission.\n", d.AppName, d.Permission)
	fmt.Println("1. Visit the URL for the auth dialog:")
	fmt.Println("")
	fmt.Println(url)
	fmt.Println("")
	fmt.Println("2. Click 'Allow' (you might have to login first)")
	fmt.Println("3. Copy the authorisation code: ")

	code := d.codeDialogue(state)

	return d.authExchange(cfg, code)
}
