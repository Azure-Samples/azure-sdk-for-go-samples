package main

import (
    "context"
    "github.com/Azure/azure-sdk-for-go/services/network/mgmt/2017-09-01/network"
    "github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
 	"github.com/Azure/go-autorest/autorest/adal"
	"github.com/Azure/go-autorest/autorest/to"
	"fmt"
	"github.com/Azure/go-autorest/autorest/azure/cli"
	"time"
	"log"
	"strings"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
)

type Config struct {

	// SPN Auth
	TenantID               string
	SubscriptionID         string
	ClientID               string
	ClientSecret           string

	// Bearer Auth
	AccessToken            *adal.Token

	Environment 		  string
}

const (
	resGroup = "GOSDKTEST"
	location = "westus"

)

var (
	ctx = context.Background()
	config Config
	environment = azure.PublicCloud.Name
)


func main() {

	// Read the token from the file created by az cli
	err := config.loadTokensFromAzureCLI()
	if err != nil {
		fmt.Errorf("Error loading the token from the CLI: %+v", err)
	}

	env, envErr := azure.EnvironmentFromName(config.Environment)
	if envErr != nil {
		log.Printf("Error loading the environment for %s : %+v", config.Environment, envErr)
	}

	oauthConfig, oauthErr := adal.NewOAuthConfig(env.ActiveDirectoryEndpoint, config.TenantID)
	if oauthErr != nil {
		log.Printf( "Error getting a new oauthToken: %+v", oauthErr)
	}

	// Get a service principal token from the refresh token
	spt, sptErr := adal.NewServicePrincipalTokenFromManualToken(*oauthConfig, config.ClientID, env.ResourceManagerEndpoint, *config.AccessToken)
	if sptErr != nil {
		fmt.Errorf("Error in getting service principal token: %+v", sptErr)
	}

	// Create the resource group
	groupsClient := resources.NewGroupsClient(config.SubscriptionID)
	groupsClient.Authorizer = autorest.NewBearerAuthorizer(spt)
	groupsClient.CreateOrUpdate(
		ctx,
		resGroup,
		resources.Group{
			Location: to.StringPtr(location)})

	// Create a Vnet and associated subnets
	vnetClient := network.NewVirtualNetworksClient(config.SubscriptionID)
	vnetClient.Authorizer = autorest.NewBearerAuthorizer(spt)
	vnetClient.CreateOrUpdate(
		ctx,
		resGroup,
		"MyVNET",
		network.VirtualNetwork{
			Location: to.StringPtr(location),
			VirtualNetworkPropertiesFormat: &network.VirtualNetworkPropertiesFormat{
				AddressSpace: &network.AddressSpace{
					AddressPrefixes: &[]string{"10.0.0.0/8"},
				},
				Subnets: &[]network.Subnet{
					{
						Name: to.StringPtr("subnet1"),
						SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
							AddressPrefix: to.StringPtr("10.0.0.0/16"),
						},
					},
					{
						Name: to.StringPtr("subnet2"),
						SubnetPropertiesFormat: &network.SubnetPropertiesFormat{
							AddressPrefix: to.StringPtr("10.1.0.0/16"),
						},
					},
				},
			},
		})
}

//
// The following block of code is borrowed from @tombuildsstuff from the work done 
// to make terraform compatible with refresh token from az cli
// 
func (c *Config) loadTokensFromAzureCLI() error {

	profilePath, err := cli.ProfilePath()
	if err != nil {
		fmt.Errorf("Error in loading profile path from the az cli: %+v", err)
	}

	profile, err := cli.LoadProfile(profilePath)
	if err != nil {
		fmt.Errorf("Authorization profile was not found: %+v", err)
	}

	// pull out the TenantID and Subscription ID from the Azure Profile
	for _, subscription := range profile.Subscriptions {
		if subscription.IsDefault {
			c.SubscriptionID = subscription.ID
			c.TenantID = subscription.TenantID
			c.Environment = environment
			break
			}
	}

	foundToken := false
	tokensPath, err := cli.AccessTokensPath()
	if err != nil {
		return fmt.Errorf("Error loading the Tokens Path from the Azure CLI %+v", err)
	}

	tokens, err := cli.LoadTokens(tokensPath)
	if err != nil {
		return fmt.Errorf("Error getting the Azure CLI authorization tokens")
	}

	for _, accessToken := range tokens {
		token, atErr := accessToken.ToADALToken()
		if atErr != nil {
			return fmt.Errorf("Error converting access token to token: %+v", atErr)
		}

		expirationDate, err := cli.ParseExpirationDate(accessToken.ExpiresOn)
		if err != nil {
			return fmt.Errorf("Error parsing the expiration date: %q", accessToken.ExpiresOn)
		}

		if expirationDate.UTC().Before(time.Now().UTC()){
			//log.Printf("Token '%s' has expired", token.AccessToken)
			continue
		}

		if !strings.Contains(accessToken.Resource, "management"){
			log.Printf("Resource '%s' is not a management domain", accessToken.Resource)
		}

		if !strings.HasSuffix(accessToken.Authority, c.TenantID){
			log.Printf("Resource '%s' is not for the correct Tenant", accessToken.Resource)
			continue
		}

		c.ClientID = accessToken.ClientID
		c.AccessToken = &token
		foundToken = true

		break

	}

	if !foundToken {
		return fmt.Errorf("No valid Azure CLI Tokens found, please login with 'az login'")
	}

	return nil
}
