// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID      string
	location            = "westus"
	resourceGroupName   = "sample-resource-group"
	availabilitySetName = "sample-availability-sets"
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

	availabilitySets, err := createAvailabilitySet(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("availability set:", *availabilitySets.ID)

	availabilitySetList, err := listAvailabilitySet(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	for i, v := range availabilitySetList {
		log.Println(i, *v.ID)
	}

	availabilitySetSizesList, err := listAvailabilitySizes(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	for i, v := range availabilitySetSizesList {
		log.Println(i, v.Name)
	}

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createAvailabilitySet(ctx context.Context, cred azcore.TokenCredential) (*armcompute.AvailabilitySet, error) {
	availabilitySetsClient, err := armcompute.NewAvailabilitySetsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := availabilitySetsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		availabilitySetName,
		armcompute.AvailabilitySet{
			Location: to.Ptr(location),
			Properties: &armcompute.AvailabilitySetProperties{
				PlatformFaultDomainCount:  to.Ptr[int32](1),
				PlatformUpdateDomainCount: to.Ptr[int32](1),
			},
			SKU: &armcompute.SKU{
				Name: to.Ptr("Aligned"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.AvailabilitySet, nil
}

func listAvailabilitySet(ctx context.Context, cred azcore.TokenCredential) ([]*armcompute.AvailabilitySet, error) {
	availabilitySetsClient, err := armcompute.NewAvailabilitySetsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pager := availabilitySetsClient.List(resourceGroupName, nil)
	availabilitySets := make([]*armcompute.AvailabilitySet, 0)
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		availabilitySets = append(availabilitySets, nextResult.Value...)
	}

	return availabilitySets, nil
}

func listAvailabilitySizes(ctx context.Context, cred azcore.TokenCredential) ([]*armcompute.VirtualMachineSize, error) {
	availabilitySetsClient, err := armcompute.NewAvailabilitySetsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pager := availabilitySetsClient.ListAvailableSizes(resourceGroupName, availabilitySetName, nil)
	availabilitySizes := make([]*armcompute.VirtualMachineSize, 0)
	for pager.More() {
		nextResult, err := pager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		availabilitySizes = append(availabilitySizes, nextResult.Value...)
	}

	return availabilitySizes, nil
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

	_, err = pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return err
	}
	return nil
}
