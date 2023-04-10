// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
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
	scopeMapName      = "sample-scope-map"
	tokenName         = "sample-token"
)

var (
	resourcesClientFactory         *armresources.ClientFactory
	containerRegistryClientFactory *armcontainerregistry.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	registriesClient    *armcontainerregistry.RegistriesClient
	scopeMapsClient     *armcontainerregistry.ScopeMapsClient
	tokensClient        *armcontainerregistry.TokensClient
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

	resourcesClientFactory, err = armresources.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	resourceGroupClient = resourcesClientFactory.NewResourceGroupsClient()

	containerRegistryClientFactory, err = armcontainerregistry.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	registriesClient = containerRegistryClientFactory.NewRegistriesClient()
	scopeMapsClient = containerRegistryClientFactory.NewScopeMapsClient()
	tokensClient = containerRegistryClientFactory.NewTokensClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	registry, err := createRegistry(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("registry:", *registry.ID)

	scopeMap, err := createScopeMap(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("scope map:", *scopeMap.ID)

	token, err := createToken(ctx, *scopeMap.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("token:", *token.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createRegistry(ctx context.Context) (*armcontainerregistry.Registry, error) {

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

func createScopeMap(ctx context.Context) (*armcontainerregistry.ScopeMap, error) {

	pollerResp, err := scopeMapsClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		scopeMapName,
		armcontainerregistry.ScopeMap{
			Properties: &armcontainerregistry.ScopeMapProperties{
				Actions: []*string{
					to.Ptr("repositories/foo/content/read"),
					to.Ptr("repositories/foo/content/delete"),
				},
				Description: to.Ptr("Developer Scopes"),
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
	return &resp.ScopeMap, nil
}

func createToken(ctx context.Context, scopeMapID string) (*armcontainerregistry.Token, error) {

	pollerResp, err := tokensClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		tokenName,
		armcontainerregistry.Token{
			Properties: &armcontainerregistry.TokenProperties{
				ScopeMapID: to.Ptr(scopeMapID),
				Status:     to.Ptr(armcontainerregistry.TokenStatusEnabled),
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
	return &resp.Token, nil
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
