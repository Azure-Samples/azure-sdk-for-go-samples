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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID              string
	TenantID                    string
	location                    = "westus"
	resourceGroupName           = "sample-resource-group"
	proximityPlacementGroupName = "sample-proximity-placement"
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	TenantID = os.Getenv("AZURE_TENANT_ID")
	if len(TenantID) == 0 {
		log.Fatal("AZURE_TENANT_ID is not set.")
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

	proximityPlacement, err := createProximityPlacement(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("proximity placement group:", *proximityPlacement.ID)

	proximityPlacement, err = getProximityPlacement(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get proximity placement group:", *proximityPlacement.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createProximityPlacement(ctx context.Context, cred azcore.TokenCredential) (*armcompute.ProximityPlacementGroup, error) {
	proximityPlacementGroupClient := armcompute.NewProximityPlacementGroupsClient(subscriptionID, cred, nil)

	resp, err := proximityPlacementGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		proximityPlacementGroupName,
		armcompute.ProximityPlacementGroup{
			Location: to.StringPtr(location),
			Properties: &armcompute.ProximityPlacementGroupProperties{
				ProximityPlacementGroupType: armcompute.ProximityPlacementGroupTypeStandard.ToPtr(),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.ProximityPlacementGroup, nil
}

func getProximityPlacement(ctx context.Context, cred azcore.TokenCredential) (*armcompute.ProximityPlacementGroup, error) {
	proximityPlacementGroupClient := armcompute.NewProximityPlacementGroupsClient(subscriptionID, cred, nil)

	resp, err := proximityPlacementGroupClient.Get(ctx, resourceGroupName, proximityPlacementGroupName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.ProximityPlacementGroup, nil
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
