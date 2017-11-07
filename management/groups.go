package management

import (
	"github.com/Azure/azure-sdk-for-go/profiles/preview/resources/mgmt/resources"
	"github.com/joshgav/az-go/common"
	"log"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

// create a new resource group named by env var
func CreateGroup() (resources.Group, error) {
	token, err := common.GetResourceManagementToken(common.OAuthGrantTypeServicePrincipal)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get auth token", err)
	}

	groupsClient := resources.NewGroupsClient(subscriptionId)
	groupsClient.Authorizer = autorest.NewBearerAuthorizer(token)

	return groupsClient.CreateOrUpdate(
		resourceGroupName,
		resources.Group{
			Location: to.StringPtr(location)})
}

func DeleteGroup() (<-chan autorest.Response, <-chan error) {
	token, err := common.GetResourceManagementToken(common.OAuthGrantTypeServicePrincipal)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get auth token", err)
	}

	groupsClient := resources.NewGroupsClient(subscriptionId)
	groupsClient.Authorizer = autorest.NewBearerAuthorizer(token)

	return groupsClient.Delete(resourceGroupName, nil)
}
