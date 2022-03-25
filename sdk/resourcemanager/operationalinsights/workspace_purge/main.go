// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

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
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	workspace, err := createWorkspace(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("operational insights workspace:", *workspace.ID)

	purge, err := purgeWorkspace(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("purge workspace:", *purge.OperationID, *purge.XMSStatusLocation)

	purgeStatus, err := purgeStatusWorkspace(ctx, cred, *purge.OperationID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("purge status workspace:", *purgeStatus.Status)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createWorkspace(ctx context.Context, cred azcore.TokenCredential) (*armoperationalinsights.Workspace, error) {
	workspacesClient := armoperationalinsights.NewWorkspacesClient(subscriptionID, cred, nil)

	pollerResp, err := workspacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		workspaceName,
		armoperationalinsights.Workspace{
			Location:   to.StringPtr(location),
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

func purgeWorkspace(ctx context.Context, cred azcore.TokenCredential) (*armoperationalinsights.WorkspacePurgeClientPurgeResult, error) {
	workspacePurgeClient := armoperationalinsights.NewWorkspacePurgeClient(subscriptionID, cred, nil)

	resp, err := workspacePurgeClient.Purge(
		ctx,
		resourceGroupName,
		workspaceName,
		armoperationalinsights.WorkspacePurgeBody{
			Filters: []*armoperationalinsights.WorkspacePurgeBodyFilters{},
			Table:   to.StringPtr(""),
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.WorkspacePurgeClientPurgeResult, nil
}

func purgeStatusWorkspace(ctx context.Context, cred azcore.TokenCredential, purgeID string) (*armoperationalinsights.WorkspacePurgeStatusResponse, error) {
	workspacePurgeClient := armoperationalinsights.NewWorkspacePurgeClient(subscriptionID, cred, nil)

	resp, err := workspacePurgeClient.GetPurgeStatus(ctx, resourceGroupName, workspaceName, purgeID, nil)
	if err != nil {
		return nil, err
	}
	return &resp.WorkspacePurgeStatusResponse, nil
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
