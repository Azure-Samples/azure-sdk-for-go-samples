package iam

import (
	"fmt"
	"log"
	"net/http"
	"os/user"
	"path/filepath"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/common"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/subosito/gotenv"
)

const (
	samplesAppID = "bee3737f-b06f-444f-b3c3-5b0f3fce46ea"
)

var (
	// for service principal and device
	clientID    string
	oauthConfig *adal.OAuthConfig
	armToken    adal.OAuthTokenProvider

	// for service principal
	subscriptionID string
	tenantID       string
	clientSecret   string
)

// OAuthGrantType specifies which grant type to use.
type OAuthGrantType int

const (
	// OAuthGrantTypeServicePrincipal for client credentials flow
	OAuthGrantTypeServicePrincipal OAuthGrantType = iota
	// OAuthGrantTypeDeviceFlow for device-auth flow
	OAuthGrantTypeDeviceFlow
)

func GetEnvVars() {
	gotenv.Load() // read from .env file

	tenantID = common.GetEnvVarOrFail("AZURE_TENANT_ID")
	clientID = common.GetEnvVarOrFail("AZURE_CLIENT_ID")
	clientSecret = common.GetEnvVarOrFail("AZURE_CLIENT_SECRET")

	var err error
	oauthConfig, err = adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, tenantID)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get OAuth config", err)
	}
}

// GetResourceManagementToken gets an OAuth token for managing resources using the specified grant type.
func GetResourceManagementToken(grantType OAuthGrantType) (adal.OAuthTokenProvider, error) {
	GetEnvVars()
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
		clientID,
		clientSecret,
		azure.PublicCloud.ResourceManagerEndpoint)
}

func getDeviceToken() (adal.OAuthTokenProvider, error) {
	sender := &http.Client{}

	code, err := adal.InitiateDeviceAuth(
		sender,
		*oauthConfig,
		samplesAppID, // clientID
		azure.PublicCloud.ResourceManagerEndpoint)
	if err != nil {
		log.Fatalf("%s: %v\n", "failed to initiate device auth", err)
	}

	fmt.Println(*code.Message)
	return adal.WaitForUserCompletion(sender, code)
}

// TODO(joshgav): use cached token when available
func getTokenCachePath() string {
	usr, err := user.Current()
	if err != nil {
		log.Fatalf("%s: %v", "failed to get current user", err)
	}
	return filepath.Join(usr.HomeDir, ".azure", "armToken.json")
}

func GetClientID() string {
	return clientID
}

func GetTenantID() string {
	return tenantID
}
