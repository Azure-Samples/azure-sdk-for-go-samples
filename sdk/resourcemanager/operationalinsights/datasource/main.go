// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	workspaceName     = "sample-workspace"
	dataSourceName    = "sample-data-source"
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

	dataSource, err := createDatasource(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("data source:", *dataSource.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createWorkspace(ctx context.Context, cred azcore.TokenCredential) (*armoperationalinsights.Workspace, error) {
	workspacesClient, err := armoperationalinsights.NewWorkspacesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := workspacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		workspaceName,
		armoperationalinsights.Workspace{
			Location:   to.Ptr(location),
			Properties: &armoperationalinsights.WorkspaceProperties{},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Workspace, nil
}

func createDatasource(ctx context.Context, cred azcore.TokenCredential) (*armoperationalinsights.DataSource, error) {
	sourcesClient, err := armoperationalinsights.NewDataSourcesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := sourcesClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		workspaceName,
		dataSourceName,
		armoperationalinsights.DataSource{
			Kind:       to.Ptr(armoperationalinsights.DataSourceKindAzureActivityLog),
			Properties: map[string]interface{}{},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.DataSource, nil
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.Ptr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, cred azcore.TokenCredential) error {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}
	return nil
}
