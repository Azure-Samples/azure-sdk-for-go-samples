package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/authorization/armauthorization"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resources/armresources"
)

var (
	subscriptionID     string
	ObjectID           string
	location           = "westus"
	resourceGroupName  = "sample-resource-group"
	BuiltInRole        = "sample-built-role"
	roleAssignmentName = "sample-role-assignment"
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	ObjectID = os.Getenv("AZURE_OBJECT_ID")
	if len(ObjectID) == 0 {
		log.Fatal("AZURE_OBJECT_ID is not set.")
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

	roleAssignment, err := createRoleAssignment(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("role assignment:", *roleAssignment.ID)

	roleAssignment, err = getRoleAssignment(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get role assignment:", *roleAssignment.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createRoleAssignment(ctx context.Context, connection *arm.Connection) (*armauthorization.RoleAssignment, error) {

	roleClient := armauthorization.NewRoleAssignmentsClient(connection, subscriptionID)
	resp, err := roleClient.Create(
		ctx,
		BuiltInRole,
		roleAssignmentName,
		armauthorization.RoleAssignmentCreateParameters{
			Properties: &armauthorization.RoleAssignmentProperties{
				PrincipalID:      to.StringPtr(ObjectID),
				RoleDefinitionID: to.StringPtr(""),
			},
		}, nil)

	if err != nil {
		return nil, err
	}
	return &resp.RoleAssignment, err
}

func getRoleAssignment(ctx context.Context, connection *arm.Connection) (*armauthorization.RoleAssignment, error) {
	roleClient := armauthorization.NewRoleAssignmentsClient(connection, subscriptionID)

	resp, err := roleClient.Get(
		ctx,
		BuiltInRole,
		roleAssignmentName,
		nil,
	)

	if err != nil {
		return nil, err
	}
	return &resp.RoleAssignment, err
}

func validateRoleAssignment(ctx context.Context, connection *arm.Connection) (*armauthorization.ValidationResponse, error) {
	roleClient := armauthorization.NewRoleAssignmentsClient(connection, subscriptionID)
	resp, err := roleClient.Validate(
		ctx,
		BuiltInRole,
		roleAssignmentName,
		armauthorization.RoleAssignmentCreateParameters{
			Properties: &armauthorization.RoleAssignmentProperties{
				PrincipalID:      to.StringPtr(ObjectID),
				RoleDefinitionID: to.StringPtr(""), // "/subscriptions/" + SUBSCRIPTION_ID + "/providers/Microsoft.Authorization/roleDefinitions/" + ROLE_DEFINITION
			},
		}, nil)

	if err != nil {
		return nil, err
	}
	return &resp.ValidationResponse, err
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
