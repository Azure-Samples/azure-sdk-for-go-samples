package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/google/uuid"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	roleDefinitionID  string
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}

	conn := arm.NewDefaultConnection(cred, &arm.ConnectionOptions{
		Logging: policy.LogOptions{
			IncludeBody: true,
		},
	})
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	roleDefinitions := listRoleDefinition(ctx, conn)
	for _, rd := range roleDefinitions {
		log.Println(*rd.Name, *rd.ID)
	}

	roleDefinitionID = uuid.New().String() //Replace with your roleDefinitionID
	roleDefinition, err := createRoleDefinition(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("role definition:", *roleDefinition.ID)

	roleDefinition, err = getRoleDefinition(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get role definition:", *roleDefinition.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createRoleDefinition(ctx context.Context, conn *arm.Connection) (*armauthorization.RoleDefinition, error) {
	roleDefinitionsClient := armauthorization.NewRoleDefinitionsClient(conn)

	resp, err := roleDefinitionsClient.CreateOrUpdate(
		ctx,
		"subscriptions/"+subscriptionID+"/resourceGroups/"+resourceGroupName,
		roleDefinitionID,
		armauthorization.RoleDefinition{
			Properties: &armauthorization.RoleDefinitionProperties{
				AssignableScopes: []*string{
					to.StringPtr("subscriptions/" + subscriptionID + "/resourceGroups/" + resourceGroupName),
				},
				Permissions: []*armauthorization.Permission{},
				RoleName:    to.StringPtr("sample"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.RoleDefinition, nil
}

func getRoleDefinition(ctx context.Context, conn *arm.Connection) (*armauthorization.RoleDefinition, error) {
	roleDefinitionsClient := armauthorization.NewRoleDefinitionsClient(conn)

	resp, err := roleDefinitionsClient.Get(
		ctx,
		"subscriptions/"+subscriptionID+"/resourceGroups/"+resourceGroupName,
		roleDefinitionID,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.RoleDefinition, nil
}

func listRoleDefinition(ctx context.Context, conn *arm.Connection) []*armauthorization.RoleDefinition {
	roleDefinitionsClient := armauthorization.NewRoleDefinitionsClient(conn)

	pager := roleDefinitionsClient.List("subscriptions/"+subscriptionID+"/resourceGroups/"+resourceGroupName, nil)

	roleDefinitions := make([]*armauthorization.RoleDefinition, 0)
	for pager.NextPage(ctx) {
		resp := pager.PageResponse()
		roleDefinitions = append(roleDefinitions, resp.Value...)
	}
	return roleDefinitions
}

func createResourceGroup(ctx context.Context, conn *arm.Connection) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.StringPtr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, conn *arm.Connection) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
}
