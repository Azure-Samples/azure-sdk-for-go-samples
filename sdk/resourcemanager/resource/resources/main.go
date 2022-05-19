// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID     string
	location           = "westus"
	resourceGroupName  = "sample-resource-group"
	virtualNetworkName = "sample-virtual-network"
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

	exist, err := checkExistResource(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources is not exist:", exist)

	resources, err := createResource(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("created resources:", *resources.ID)

	genericResource, err := getResource(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get resource:", *genericResource.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

var resourceProviderNamespace = "Microsoft.Network"
var resourceType = "virtualNetworks"
var apiVersion = "2021-02-01"

func checkExistResource(ctx context.Context, cred azcore.TokenCredential) (bool, error) {
	resourceClient, err := armresources.NewClient(subscriptionID, cred, nil)
	if err != nil {
		return false, err
	}

	boolResp, err := resourceClient.CheckExistence(
		ctx,
		resourceGroupName,
		resourceProviderNamespace,
		"/",
		resourceType,
		virtualNetworkName,
		apiVersion,
		nil)
	if err != nil {
		return false, err
	}

	return boolResp.Success, nil
}

func createResource(ctx context.Context, cred azcore.TokenCredential) (*armresources.GenericResource, error) {
	resourceClient, err := armresources.NewClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := resourceClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		resourceProviderNamespace,
		"/",
		resourceType,
		virtualNetworkName,
		apiVersion,
		armresources.GenericResource{
			Location: to.Ptr(location),
			Properties: map[string]interface{}{
				"addressSpace": armnetwork.AddressSpace{
					AddressPrefixes: []*string{
						to.Ptr("10.1.0.0/16"),
					},
				},
			},
		},
		nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}

	return &resp.GenericResource, nil
}

func getResource(ctx context.Context, cred azcore.TokenCredential) (*armresources.GenericResource, error) {
	resourceClient, err := armresources.NewClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := resourceClient.Get(
		ctx,
		resourceGroupName,
		resourceProviderNamespace,
		"/",
		resourceType,
		virtualNetworkName,
		apiVersion,
		nil)
	if err != nil {
		return nil, err
	}

	return &resp.GenericResource, nil
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
