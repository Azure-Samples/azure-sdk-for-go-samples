package resources

import (
  "log"
  "github.com/joshgav/az-go/util"

  "github.com/subosito/gotenv"

  "github.com/Azure/azure-sdk-for-go/arm/resources/resources"
  "github.com/Azure/go-autorest/autorest"
  "github.com/Azure/go-autorest/autorest/azure"
  "github.com/Azure/go-autorest/autorest/adal"
  "github.com/Azure/go-autorest/autorest/to"
)

var (
  SubscriptionId string
  TenantId string
  ResourceGroupName string
  Location string
  Location2 string = "westus2"
  // keep these private and share the token
  clientId string = ""
  clientSecret string = ""
  Token *adal.ServicePrincipalToken
)

func init() {
  gotenv.Load() // read from .env file

	SubscriptionId    = util.GetEnvVarOrFail("AZURE_SUBSCRIPTION_ID")
	TenantId          = util.GetEnvVarOrFail("AZURE_TENANT_ID")
	ResourceGroupName = util.GetEnvVarOrFail("AZURE_RG_NAME")
  Location          = util.GetEnvVarOrFail("AZURE_LOCATION")
	clientId          = util.GetEnvVarOrFail("AZURE_CLIENT_ID")
	clientSecret      = util.GetEnvVarOrFail("AZURE_CLIENT_SECRET")

  oauthConfig, err := adal.NewOAuthConfig(azure.PublicCloud.ActiveDirectoryEndpoint, TenantId)
  util.OnErrorFail(err, "oauth configuration failed")

  Token, err = adal.NewServicePrincipalToken(
    *oauthConfig,
    clientId,
    clientSecret,
    azure.PublicCloud.ResourceManagerEndpoint)
  util.OnErrorFail(err, "failed to get token")
}

func CreateGroup() (resources.Group, error) {
  groupsClient := resources.NewGroupsClient(SubscriptionId)
  groupsClient.Authorizer = autorest.NewBearerAuthorizer(Token)

  return groupsClient.CreateOrUpdate(
    ResourceGroupName,
    resources.Group{
      Location: to.StringPtr(Location)})
}

func DeleteGroup() error {
  groupsClient := resources.NewGroupsClient(SubscriptionId)
  groupsClient.Authorizer = autorest.NewBearerAuthorizer(Token)

  response, errC := groupsClient.Delete(ResourceGroupName, nil)
  err := <-errC
  log.Println(<-response)
  return err
}

