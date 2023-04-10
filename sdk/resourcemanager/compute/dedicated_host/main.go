// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/compute/armcompute/v4"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID    string
	TenantID          string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	hostGroupName     = "sample-host-group"
	hostName          = "sample-host"
)

var (
	resourcesClientFactory *armresources.ClientFactory
	computeClientFactory   *armcompute.ClientFactory
)

var (
	resourceGroupClient       *armresources.ResourceGroupsClient
	dedicatedHostGroupsClient *armcompute.DedicatedHostGroupsClient
	dedicatedHostsClient      *armcompute.DedicatedHostsClient
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

	resourcesClientFactory, err = armresources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	resourceGroupClient = resourcesClientFactory.NewResourceGroupsClient()

	computeClientFactory, err = armcompute.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	dedicatedHostGroupsClient = computeClientFactory.NewDedicatedHostGroupsClient()
	dedicatedHostsClient = computeClientFactory.NewDedicatedHostsClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	dedicatedHostGroup, err := createDedicatedHostGroups(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("dedicated host group:", *dedicatedHostGroup.ID)

	dedicatedHost, err := createDedicatedHost(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("dedicated host:", *dedicatedHost.ID)

	dedicatedHostGroup, err = getDedicatedHostGroups(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get dedicated host:", *dedicatedHostGroup.ID)

	dedicatedHost, err = getDedicatedHost(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get dedicated host:", *dedicatedHost.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err := cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createDedicatedHostGroups(ctx context.Context) (*armcompute.DedicatedHostGroup, error) {

	resp, err := dedicatedHostGroupsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		hostGroupName,
		armcompute.DedicatedHostGroup{
			Location: to.Ptr("eastus"),
			Properties: &armcompute.DedicatedHostGroupProperties{
				PlatformFaultDomainCount: to.Ptr[int32](3),
			},
			Zones: []*string{to.Ptr("1")},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.DedicatedHostGroup, nil
}

func getDedicatedHostGroups(ctx context.Context) (*armcompute.DedicatedHostGroup, error) {

	resp, err := dedicatedHostGroupsClient.Get(ctx, resourceGroupName, hostGroupName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.DedicatedHostGroup, nil
}

func createDedicatedHost(ctx context.Context) (*armcompute.DedicatedHost, error) {

	pollerResp, err := dedicatedHostsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		hostGroupName,
		hostName,
		armcompute.DedicatedHost{
			Location: to.Ptr("eastus"),
			Properties: &armcompute.DedicatedHostProperties{
				PlatformFaultDomain: to.Ptr[int32](1),
			},
			SKU: &armcompute.SKU{
				Name: to.Ptr("DSv3-Type1"),
			},
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
	return &resp.DedicatedHost, nil
}

func getDedicatedHost(ctx context.Context) (*armcompute.DedicatedHost, error) {

	resp, err := dedicatedHostsClient.Get(ctx, resourceGroupName, hostGroupName, hostName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.DedicatedHost, nil
}

func createResourceGroup(ctx context.Context) (*armresources.ResourceGroup, error) {

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

func cleanup(ctx context.Context) error {

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
