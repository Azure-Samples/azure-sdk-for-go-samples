// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package iam

import (
	"log"
	"net/url"
	"os"
	"strings"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

const (
	samplesAppID  = "bee3737f-b06f-444f-b3c3-5b0f3fce46ea"
	azCLIclientID = "04b07795-8ddb-461a-bbee-02f9e1bf7b46"
)

var (
	// for service principal and device
	clientID           string
	oauthConfig        *adal.OAuthConfig
	armAuthorizer      autorest.Authorizer
	batchAuthorizer    autorest.Authorizer
	graphAuthorizer    autorest.Authorizer
	keyvaultAuthorizer autorest.Authorizer

	// for service principal
	subscriptionID string
	tenantID       string
	clientSecret   string
	// UseCLIclientID sets if the Azure CLI client ID should be used on device authentication
	UseCLIclientID bool
)

// OAuthGrantType specifies which grant type to use.
type OAuthGrantType int

const (
	// OAuthGrantTypeServicePrincipal for client credentials flow
	OAuthGrantTypeServicePrincipal OAuthGrantType = iota
	// OAuthGrantTypeDeviceFlow for device-auth flow
	OAuthGrantTypeDeviceFlow
)

// ParseArgs picks up shared env vars
// Other packages should use this func after helpers.ParseArgs()
func ParseArgs() error {
	err := helpers.ReadEnvFile()
	if err != nil {
		return err
	}

	tenantID = os.Getenv("AZURE_TENANT_ID")
	clientID = os.Getenv("AZURE_CLIENT_ID")
	clientSecret = os.Getenv("AZURE_CLIENT_SECRET")

	oauthConfig, err = adal.NewOAuthConfig(helpers.Environment().ActiveDirectoryEndpoint, tenantID)
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

// ClientSecret gets the client secret
func ClientSecret() string {
	return clientSecret
}

// AuthGrantType returns what kind of authentication is going to be used: device flow or service principal
func AuthGrantType() OAuthGrantType {
	if helpers.DeviceFlow() {
		return OAuthGrantTypeDeviceFlow
	}
	return OAuthGrantTypeServicePrincipal
}

// GetResourceManagementAuthorizer gets an OAuth token for managing resources using the specified grant type.
func GetResourceManagementAuthorizer(grantType OAuthGrantType) (a autorest.Authorizer, err error) {
	if armAuthorizer != nil {
		return armAuthorizer, nil
	}

	switch grantType {
	case OAuthGrantTypeServicePrincipal:
		a, err = auth.NewAuthorizerFromEnvironment()
	case OAuthGrantTypeDeviceFlow:
		config := auth.NewDeviceFlowConfig(samplesAppID, tenantID)
		a, err = config.Authorizer()
	default:
		log.Fatalln("invalid token type specified")
	}

	if err == nil {
		armAuthorizer = a
	}
	return
}

// GetBatchAuthorizer gets an authorizer for Azure batch using the specified grant type.
func GetBatchAuthorizer(grantType OAuthGrantType) (a autorest.Authorizer, err error) {
	if batchAuthorizer != nil {
		return batchAuthorizer, nil
	}

	a, err = getAuthorizer(grantType, helpers.Environment().BatchManagementEndpoint)
	if err == nil {
		batchAuthorizer = a
	}

	return
}

// GetGraphAuthorizer gets an authorizer for the graphrbac API using the specified grant type.
func GetGraphAuthorizer(grantType OAuthGrantType) (a autorest.Authorizer, err error) {
	if graphAuthorizer != nil {
		return graphAuthorizer, nil
	}

	a, err = getAuthorizer(grantType, helpers.Environment().GraphEndpoint)
	if err == nil {
		graphAuthorizer = a
	}

	return
}

// GetResourceManagementTokenHybrid retrieves auth token for hybrid environment
func GetResourceManagementTokenHybrid(activeDirectoryEndpoint, tokenAudience string) (adal.OAuthTokenProvider, error) {
	var token adal.OAuthTokenProvider
	oauthConfig, err := adal.NewOAuthConfig(activeDirectoryEndpoint, tenantID)
	token, err = adal.NewServicePrincipalToken(
		*oauthConfig,
		clientID,
		clientSecret,
		tokenAudience)

	return token, err
}

func getAuthorizer(grantType OAuthGrantType, endpoint string) (a autorest.Authorizer, err error) {
	switch grantType {
	case OAuthGrantTypeServicePrincipal:
		token, err := adal.NewServicePrincipalToken(*oauthConfig, clientID, clientSecret, endpoint)
		if err != nil {
			return a, err
		}
		a = autorest.NewBearerAuthorizer(token)
	case OAuthGrantTypeDeviceFlow:
		config := auth.NewDeviceFlowConfig(samplesAppID, tenantID)
		config.Resource = endpoint
		a, err = config.Authorizer()
	default:
		log.Fatalln("invalid token type specified")
	}
	return
}

// GetKeyvaultAuthorizer gets an authorizer for the keyvault dataplane
func GetKeyvaultAuthorizer(grantType OAuthGrantType) (a autorest.Authorizer, err error) {
	if keyvaultAuthorizer != nil {
		return keyvaultAuthorizer, nil
	}

	vaultEndpoint := strings.TrimSuffix(helpers.Environment().KeyVaultEndpoint, "/")
	config, err := adal.NewOAuthConfig(helpers.Environment().ActiveDirectoryEndpoint, tenantID)
	updatedAuthorizeEndpoint, err := url.Parse("https://login.windows.net/" + tenantID + "/oauth2/token")
	config.AuthorizeEndpoint = *updatedAuthorizeEndpoint
	if err != nil {
		return
	}

	switch grantType {
	case OAuthGrantTypeServicePrincipal:
		token, err := adal.NewServicePrincipalToken(*config, clientID, clientSecret, vaultEndpoint)
		if err != nil {
			return a, err
		}
		a = autorest.NewBearerAuthorizer(token)
	case OAuthGrantTypeDeviceFlow:
		deviceConfig := auth.NewDeviceFlowConfig(samplesAppID, tenantID)
		deviceConfig.Resource = vaultEndpoint
		deviceConfig.AADEndpoint = updatedAuthorizeEndpoint.String()
		a, err = deviceConfig.Authorizer()
	default:
		log.Fatalln("invalid token type specified")
	}

	if err == nil {
		keyvaultAuthorizer = a
	}

	return
}
