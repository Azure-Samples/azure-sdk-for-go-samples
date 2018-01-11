package resources

import (
	"context"
	"fmt"
	"log"
	"strings"
	"sync"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/resources/mgmt/2017-05-10/resources"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

func getGroupsClient() resources.GroupsClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	groupsClient := resources.NewGroupsClient(helpers.SubscriptionID())
	groupsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return groupsClient
}

// CreateGroup creates a new resource group named by env var
func CreateGroup(ctx context.Context, groupName string) (resources.Group, error) {
	groupsClient := getGroupsClient()
	log.Println(fmt.Sprintf("creating resource group on location: %v", helpers.Location()))
	return groupsClient.CreateOrUpdate(
		ctx,
		groupName,
		resources.Group{
			Location: to.StringPtr(helpers.Location()),
		})
}

// DeleteGroup removes the resource group named by env var
func DeleteGroup(ctx context.Context, groupName string) (result resources.GroupsDeleteFuture, err error) {
	groupsClient := getGroupsClient()
	return groupsClient.Delete(ctx, groupName)
}

// ListGroups gets an interator that gets all resource groups in the subscription
func ListGroups(ctx context.Context) (resources.GroupListResultIterator, error) {
	groupsClient := getGroupsClient()
	return groupsClient.ListComplete(ctx, "", nil)
}

// DeleteAllGroupsWithPrefix deletes all rescource groups that start with a certain prefix
func DeleteAllGroupsWithPrefix(ctx context.Context, prefix string) (futures []resources.GroupsDeleteFuture, groups []string) {
	if helpers.KeepResources() {
		log.Println("keeping resource groups")
		return
	}
	for list, err := ListGroups(ctx); list.NotDone(); err = list.Next() {
		if err != nil {
			log.Fatalf("got error: %s", err)
		}
		rgName := *list.Value().Name
		if strings.HasPrefix(rgName, prefix) {
			fmt.Printf("deleting group '%s'\n", rgName)
			future, err := DeleteGroup(ctx, rgName)
			if err != nil {
				log.Fatalf("got error: %s", err)
			}
			futures = append(futures, future)
			groups = append(groups, rgName)
		}
	}
	return
}

// WaitForDeleteCompletion concurrently waits for delete group operations to finish
func WaitForDeleteCompletion(ctx context.Context, wg *sync.WaitGroup, futures []resources.GroupsDeleteFuture, groups []string) {
	for i, f := range futures {
		wg.Add(1)
		go func(ctx context.Context, future resources.GroupsDeleteFuture, rg string) {
			err := future.WaitForCompletion(ctx, getGroupsClient().Client)
			if err != nil {
				log.Fatalf("got error: %s", err)
			} else {
				fmt.Printf("finished deleting group '%s'\n", rg)
			}
			wg.Done()
		}(ctx, f, groups[i])
	}
}
