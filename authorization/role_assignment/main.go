package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID     string
	objectID           string
	scope              string
	location           = "westus"
	resourceGroupName  = "sample-resource-group"
	roleAssignmentName = "sample-role-assignment"
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	objectID = os.Getenv("AZURE_OBJECT_ID")
	if len(objectID) == 0 {
		log.Fatal("AZURE_OBJECT_ID is not set.")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)
	scope = *resourceGroup.ID

	roleAssignment, err := createRoleAssignment(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("role assignment:", *roleAssignment.ID)

	roleAssignment, err = getRoleAssignment(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get role assignment:", *roleAssignment.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createRoleAssignment(ctx context.Context, cred azcore.TokenCredential) (*armauthorization.RoleAssignment, error) {
	roleClient := armauthorization.NewRoleAssignmentsClient(subscriptionID, cred, nil)
	resp, err := roleClient.Create(
		ctx,
		scope,
		roleAssignmentName,
		armauthorization.RoleAssignmentCreateParameters{
			Properties: &armauthorization.RoleAssignmentProperties{
				PrincipalID:      to.StringPtr(objectID),
				RoleDefinitionID: to.StringPtr(""),
			},
		}, nil)

	if err != nil {
		return nil, err
	}
	return &resp.RoleAssignment, err
}

func getRoleAssignment(ctx context.Context, cred azcore.TokenCredential) (*armauthorization.RoleAssignment, error) {
	roleClient := armauthorization.NewRoleAssignmentsClient(subscriptionID, cred, nil)

	resp, err := roleClient.Get(
		ctx,
		scope,
		roleAssignmentName,
		nil,
	)

	if err != nil {
		return nil, err
	}
	return &resp.RoleAssignment, err
}

func validateRoleAssignment(ctx context.Context, cred azcore.TokenCredential) (*armauthorization.ValidationResponse, error) {
	roleClient := armauthorization.NewRoleAssignmentsClient(subscriptionID, cred, nil)
	resp, err := roleClient.Validate(
		ctx,
		scope,
		roleAssignmentName,
		armauthorization.RoleAssignmentCreateParameters{
			Properties: &armauthorization.RoleAssignmentProperties{
				PrincipalID:      to.StringPtr(objectID),
				RoleDefinitionID: to.StringPtr(""), // "/subscriptions/" + SUBSCRIPTION_ID + "/providers/Microsoft.Authorization/roleDefinitions/" + ROLE_DEFINITION
			},
		}, nil)

	if err != nil {
		return nil, err
	}
	return &resp.ValidationResponse, err
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

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

func cleanup(ctx context.Context, cred azcore.TokenCredential) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

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
