package resources

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"

	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

func getGroupsClient() resources.GroupsClient {
	token, _ := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	groupsClient := resources.NewGroupsClient(helpers.SubscriptionID())
	groupsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return groupsClient
}

// CreateGroup creates a new resource group named by env var
func CreateGroup(groupName string) (resources.Group, error) {
	groupsClient := getGroupsClient()

	return groupsClient.CreateOrUpdate(
		context.Background(),
		groupName,
		resources.Group{
			Location: to.StringPtr(helpers.Location()),
		})
}

// DeleteGroup removes the resource group named by env var
func DeleteGroup(groupName string) (resources.GroupsDeleteFuture, error) {
	groupsClient := getGroupsClient()
	return groupsClient.Delete(context.Background(), groupName)
}
