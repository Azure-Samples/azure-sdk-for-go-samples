package resources

import (
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

// CreateGroup creates a new resource group named by env var
func CreateGroup() (resources.Group, error) {
	token, err := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get auth token", err)
	}

	groupsClient := resources.NewGroupsClient(helpers.SubscriptionID)
	groupsClient.Authorizer = autorest.NewBearerAuthorizer(token)

	return groupsClient.CreateOrUpdate(
		helpers.ResourceGroupName,
		resources.Group{
			Location: to.StringPtr(helpers.Location)})
}

// DeleteGroup removes the resource group named by env var
func DeleteGroup() (<-chan autorest.Response, <-chan error) {
	token, err := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get auth token", err)
	}

	groupsClient := resources.NewGroupsClient(helpers.SubscriptionID)
	groupsClient.Authorizer = autorest.NewBearerAuthorizer(token)

	return groupsClient.Delete(helpers.ResourceGroupName, nil)
}
