package common

import (
	"fmt"
	"log"
	"net/http"
	"os/user"
	"path/filepath"

	"github.com/subosito/gotenv"

	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
)

const (
	samplesAppId = "bee3737f-b06f-444f-b3c3-5b0f3fce46ea"
)

var (
	// for service principal and device
	clientId    string
	oauthConfig *adal.OAuthConfig
	armToken    adal.OAuthTokenProvider

	// for service principal
	subscriptionId string
	tenantId       string
	clientSecret   string
)

// token types
type OAuthGrantType int

const (
	OAuthGrantTypeServicePrincipal OAuthGrantType = iota
	OAuthGrantTypeDeviceFlow
)

type ClientCredentialType int

const (
	ClientCredentialTypeSecret ClientCredentialType = iota
	ClientCredentialTypeCertificate
)

func init() {
	gotenv.Load() // read from .env file

	subscriptionId = GetEnvVarOrFail("AZURE_SUBSCRIPTION_ID")
	tenantId = GetEnvVarOrFail("AZURE_TENANT_ID")
	clientId = GetEnvVarOrFail("AZURE_CLIENT_ID")
	clientSecret = GetEnvVarOrFail("AZURE_CLIENT_SECRET")

	var err error
	oauthConfig, err = adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, tenantId)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get OAuth config", err)
	}
}

// get OAuth token for managing resources (ARM)
func GetResourceManagementToken(grantType OAuthGrantType) (adal.OAuthTokenProvider, error) {
	// TODO(joshgav): if cached token is available retrieve that
	if armToken != nil {
		return armToken, nil
	}

	var err error
	var token adal.OAuthTokenProvider

	switch grantType {
	case OAuthGrantTypeServicePrincipal:
		token, err = getServicePrincipalToken()
	case OAuthGrantTypeDeviceFlow:
		token, err = getDeviceToken()
	default:
		log.Fatalln("invalid token type specified")
	}
	if err == nil {
		armToken = token
		//TODO(joshgav): cache token to fs
	}
	return token, err
}

func getServicePrincipalToken() (adal.OAuthTokenProvider, error) {
	return adal.NewServicePrincipalToken(
		*oauthConfig,
		clientId,
		clientSecret,
		azure.PublicCloud.ResourceManagerEndpoint)
}

func getDeviceToken() (adal.OAuthTokenProvider, error) {
	sender := &http.Client{}

	code, err := adal.InitiateDeviceAuth(
		sender,
		*oauthConfig,
		samplesAppId, // clientID
		azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		log.Fatalf("%s: %v\n", "failed to initiate device auth", err)
	}

	fmt.Println(*code.Message)
	return adal.WaitForUserCompletion(sender, code)
}

func getTokenCachePath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("%s: %v", "failed to get current user", err)
	}
	return filepath.Join(usr.HomeDir, ".azure", "armToken.json")
}
