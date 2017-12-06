package resources

import (
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/management"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

func getGroupsClient() resources.GroupsClient {
	groupsClient := resources.NewGroupsClient(management.GetSubID())
	groupsClient.Authorizer = management.GetToken()
	return groupsClient
}

// CreateGroup creates a new resource group named by env var
func CreateGroup() (resources.Group, error) {
	groupsClient := getGroupsClient()
	return groupsClient.CreateOrUpdate(
		management.GetResourceGroup(),
		resources.Group{
			Location: to.StringPtr(management.GetLocation()),
		})
}

// DeleteGroup removes the resource group named by env var
func DeleteGroup() (<-chan autorest.Response, <-chan error) {
	groupsClient := getGroupsClient()
	return groupsClient.Delete(management.GetResourceGroup(), nil)
}

func Cleanup() error {
	if management.KeepResources() {
		log.Println("keeping resources")
		return nil
	}
	log.Println("deleting resources")
	_, errChan := DeleteGroup()
	return <-errChan
}
