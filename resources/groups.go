package resources

import (
	"log"
	"github.com/joshgav/az-go/common"

	"github.com/Azure/azure-sdk-for-go/arm/resources/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/subosito/gotenv"
)

var (
	SubscriptionId    string
	ResourceGroupName string
	Location          string
)

func init() {
	gotenv.Load() // read from .env file

	SubscriptionId = common.GetEnvVarOrFail("AZURE_SUBSCRIPTION_ID")
	ResourceGroupName = common.GetEnvVarOrFail("AZURE_RG_NAME")
	Location = common.GetEnvVarOrFail("AZURE_LOCATION")
}

// create a new resource group named by env var
func CreateGroup() (resources.Group, error) {
	token, err := common.GetResourceManagementToken(common.OAuthGrantTypeServicePrincipal)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get auth token", err)
	}

	groupsClient := resources.NewGroupsClient(SubscriptionId)
	groupsClient.Authorizer = autorest.NewBearerAuthorizer(token)

	return groupsClient.CreateOrUpdate(
		ResourceGroupName,
		resources.Group{
			Location: to.StringPtr(Location)})
}

func DeleteGroup() error {
	token, err := common.GetResourceManagementToken(common.OAuthGrantTypeServicePrincipal)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get auth token", err)
	}

	groupsClient := resources.NewGroupsClient(SubscriptionId)
	groupsClient.Authorizer = autorest.NewBearerAuthorizer(token)

	response, errC := groupsClient.Delete(ResourceGroupName, nil)
	err = <-errC
	log.Println(<-response)
	return err
}
