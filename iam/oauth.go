package iam

import (
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"strings"

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

func init() {
	err := parseArgs()
	if err != nil {
		log.Fatalf("failed to parse args: %s\n", err)
	}
}

func parseArgs() error {
	err := gotenv.Load() // read from .env file
	if err != nil && !strings.HasPrefix(err.Error(), "open .env:") {
		return err
	}

	tenantID = os.Getenv("AZ_TENANT_ID")
	clientID = os.Getenv("AZ_CLIENT_ID")
	clientSecret = os.Getenv("AZ_CLIENT_SECRET")

	if !(len(tenantID) > 0) || !(len(clientID) > 0) || !(len(clientSecret) > 0) {
		return errors.New("tenant id, client id, and client secret must be specified via env var or flags")
	}

	oauthConfig, err = adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, tenantID)

	return err
}

// ClientID gets the client ID
func ClientID() string {
	return clientID
}

// TenantID gets the client ID
func TenantID() string {
	return tenantID
}

// GetResourceManagementToken gets an OAuth token for managing resources using the specified grant type.
func GetResourceManagementToken(grantType OAuthGrantType) (adal.OAuthTokenProvider, error) {
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
