// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	registryName      = "sample2registry"
	replicationName   = "sample2replication"
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

	registry, err := createRegistry(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("registry:", *registry.ID)

	replication, err := createReplication(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("replication:", *replication.ID)

	replication, err = getReplication(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get replication:", *replication.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createRegistry(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.Registry, error) {
	registriesClient, err := armcontainerregistry.NewRegistriesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := registriesClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		armcontainerregistry.Registry{
			Location: to.Ptr(location),
			Tags: map[string]*string{
				"key": to.Ptr("value"),
			},
			SKU: &armcontainerregistry.SKU{
				Name: to.Ptr(armcontainerregistry.SKUNamePremium),
			},
			Properties: &armcontainerregistry.RegistryProperties{
				AdminUserEnabled: to.Ptr(true),
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
	return &resp.Registry, nil
}

func createReplication(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.Replication, error) {
	replicationsClient, err := armcontainerregistry.NewReplicationsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := replicationsClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		replicationName,
		armcontainerregistry.Replication{
			Location: to.Ptr("eastus"),
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
	return &resp.Replication, nil
}

func getReplication(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.Replication, error) {
	replicationsClient, err := armcontainerregistry.NewReplicationsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := replicationsClient.Get(ctx, resourceGroupName, registryName, replicationName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Replication, nil
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
