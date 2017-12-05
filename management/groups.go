package management

import (
	"github.com/Azure/azure-sdk-for-go/profiles/preview/resources/mgmt/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

func CreateGroup() (resources.Group, error) {
	groupsClient := resources.NewGroupsClient(subscriptionId)
	groupsClient.Authorizer = token

	return groupsClient.CreateOrUpdate(
		resourceGroupName,
		resources.Group{
			Location: to.StringPtr(location),
		})
}

func DeleteGroup() (<-chan autorest.Response, <-chan error) {
	groupsClient := resources.NewGroupsClient(subscriptionId)
	groupsClient.Authorizer = token

	return groupsClient.Delete(resourceGroupName, nil)
}
