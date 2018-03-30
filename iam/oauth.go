// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package iam

import (
	"log"
	"net/http"
	"net/url"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/azure/auth"
)

const (
	samplesAppID  = "bee3737f-b06f-444f-b3c3-5b0f3fce46ea"
	azCLIclientID = "04b07795-8ddb-461a-bbee-02f9e1bf7b46"
)

var (
	// for service principal and device
	clientID      string
	oauthConfig   *adal.OAuthConfig
	armAuthorizer autorest.Authorizer
	batchToken    adal.OAuthTokenProvider
	graphToken    adal.OAuthTokenProvider

	// for service principal
	subscriptionID string
	tenantID       string
	clientSecret   string
	// UseCLIclientID sets if the Azure CLI client iD should be used on device authentication
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

func init() {
	err := parseArgs()
	if err != nil {
		log.Fatalf("failed to parse args: %s\n", err)
	}
}

func parseArgs() error {
	return helpers.LoadEnvVars()
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
		var token adal.OAuthTokenProvider
		token, err = getDeviceToken(azure.PublicCloud.ResourceManagerEndpoint)
		a = autorest.NewBearerAuthorizer(token)
	default:
		log.Fatalln("invalid token type specified")
	}

	if err == nil {
		armAuthorizer = a
	}
	return
}

const batchManagementEndpoint = "https://batch.core.windows.net/"

// GetBatchToken gets an OAuth token for Azure batch using the specified grant type.
func GetBatchToken(grantType OAuthGrantType) (adal.OAuthTokenProvider, error) {
	if batchToken != nil {
		return batchToken, nil
	}

	token, err := getToken(grantType, batchManagementEndpoint)
	if err == nil {
		batchToken = token
	}

	return token, err
}

// GetGraphToken gets an OAuth token for the graphrbac API using the specified grant type.
func GetGraphToken(grantType OAuthGrantType) (adal.OAuthTokenProvider, error) {
	if graphToken != nil {
		return graphToken, nil
	}

	token, err := getToken(grantType, azure.PublicCloud.GraphEndpoint)
	if err == nil {
		graphToken = token
	}

	return token, err
}

func getToken(grantType OAuthGrantType, endpoint string) (token adal.OAuthTokenProvider, err error) {
	switch grantType {
	case OAuthGrantTypeServicePrincipal:
		token, err = getServicePrincipalToken(endpoint)
	case OAuthGrantTypeDeviceFlow:
		token, err = getDeviceToken(endpoint)
	default:
		log.Fatalln("invalid token type specified")
	}
	return
}

func getServicePrincipalToken(endpoint string) (adal.OAuthTokenProvider, error) {
	return adal.NewServicePrincipalToken(
		*oauthConfig,
		clientID,
		clientSecret,
		endpoint)
}

func getDeviceToken(endpoint string) (adal.OAuthTokenProvider, error) {
	sender := &http.Client{}
	cliID := samplesAppID
	if UseCLIclientID {
		cliID = azCLIclientID
	}
	code, err := adal.InitiateDeviceAuth(
		sender,
		*oauthConfig,
		cliID, // clientID
		endpoint)
	if err != nil {
		log.Fatalf("%s: %v\n", "failed to initiate device auth", err)
	}

	log.Println(*code.Message)
	return adal.WaitForUserCompletion(sender, code)
}

// GetKeyvaultToken gets an authorizer for the keyvault dataplane
func GetKeyvaultToken(grantType OAuthGrantType) (authorizer autorest.Authorizer, err error) {
	config, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, tenantID)
	updatedAuthorizeEndpoint, err := url.Parse("https://login.windows.net/" + tenantID + "/oauth2/token")
	config.AuthorizeEndpoint = *updatedAuthorizeEndpoint
	if err != nil {
		return
	}

	switch grantType {
	case OAuthGrantTypeServicePrincipal:
		spt, err := adal.NewServicePrincipalToken(
			*config,
			clientID,
			clientSecret,
			"https://vault.azure.net")

		if err != nil {
			return authorizer, err
		}
		authorizer = autorest.NewBearerAuthorizer(spt)
	case OAuthGrantTypeDeviceFlow:
		sender := &http.Client{}

		code, err := adal.InitiateDeviceAuth(
			sender,
			*config,
			samplesAppID, // clientID
			"https://vault.azure.net")
		if err != nil {
			log.Fatalf("%s: %v\n", "failed to initiate device auth", err)
		}

		log.Println(*code.Message)
		spt, err := adal.WaitForUserCompletion(sender, code)
		if err != nil {
			return authorizer, err
		}
		authorizer = autorest.NewBearerAuthorizer(spt)
	default:
		log.Fatalln("invalid token type specified")
	}

	return
}
