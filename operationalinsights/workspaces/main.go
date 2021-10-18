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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	workspaceName     = "sample-workspace"
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

	workspace, err := createWorkspace(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("operational insights workspace:", *workspace.ID)

	workspace, err = getWorkspace(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("operational insights get workspace:", *workspace.ID)

	workspaces, err := listWorkspace(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	for _, w := range workspaces {
		log.Println(*w.Name, *w.ID)
	}

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createWorkspace(ctx context.Context, conn *arm.Connection) (*armoperationalinsights.Workspace, error) {
	workspacesClient := armoperationalinsights.NewWorkspacesClient(conn, subscriptionID)

	pollerResp, err := workspacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		workspaceName,
		armoperationalinsights.Workspace{
			TrackedResource: armoperationalinsights.TrackedResource{
				Location: to.StringPtr(location),
			},
			Properties: &armoperationalinsights.WorkspaceProperties{},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.Workspace, nil
}

func getWorkspace(ctx context.Context, conn *arm.Connection) (*armoperationalinsights.Workspace, error) {
	workspacesClient := armoperationalinsights.NewWorkspacesClient(conn, subscriptionID)

	resp, err := workspacesClient.Get(ctx, resourceGroupName, workspaceName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Workspace, nil
}

func listWorkspace(ctx context.Context, conn *arm.Connection) ([]*armoperationalinsights.Workspace, error) {
	workspacesClient := armoperationalinsights.NewWorkspacesClient(conn, subscriptionID)

	workspaceResp, err := workspacesClient.ListByResourceGroup(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, err
	}
	workspaces := make([]*armoperationalinsights.Workspace, 0)
	workspaces = append(workspaces, workspaceResp.Value...)
	return workspaces, nil
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
