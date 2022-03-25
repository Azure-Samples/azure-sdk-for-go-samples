// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID            string
	resourceProviderNamespace = "Microsoft.Compute"
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

	provider, err := registerProvider(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("registered provider:", *provider.ID)

	provider, err = getProvider(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get provider:", *provider.ID)

	providers := listProvider(ctx, cred)
	log.Println("providers:", len(providers))
	l := math.Min(10, float64(len(providers)))
	for i := 0; i < int(l); i++ {
		log.Printf("Namespace: %s,RegistratonState: %s\n", *providers[i].Namespace, *providers[i].RegistrationState)
	}

	providerPermissionsResult, err := providerPermissions(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	data, err := json.Marshal(providerPermissionsResult)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(data))

	// Tenant
	providers = listAtTenantScopeProvider(ctx, cred)
	log.Println("list providers:", len(providers))
	l = math.Min(10, float64(len(providers)))
	for i := 0; i < int(l); i++ {
		log.Println("Namespace:", *providers[i].Namespace)
	}

	atTenant, err := getAtTenantScopeProvider(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get atTenant:", *atTenant.Namespace)
}

func registerProvider(ctx context.Context, cred azcore.TokenCredential) (*armresources.Provider, error) {
	providerClient := armresources.NewProvidersClient(subscriptionID, cred, nil)

	providerResp, err := providerClient.Register(ctx, resourceProviderNamespace, nil)
	if err != nil {
		return nil, err
	}

	return &providerResp.Provider, nil
}

func getProvider(ctx context.Context, cred azcore.TokenCredential) (*armresources.Provider, error) {
	providerClient := armresources.NewProvidersClient(subscriptionID, cred, nil)

	providerResp, err := providerClient.Get(ctx, resourceProviderNamespace, nil)
	if err != nil {
		return nil, err
	}

	return &providerResp.Provider, nil
}

func listProvider(ctx context.Context, cred azcore.TokenCredential) []*armresources.Provider {
	providerClient := armresources.NewProvidersClient(subscriptionID, cred, nil)

	providerList := providerClient.List(nil)

	var providers = make([]*armresources.Provider, 0)
	for providerList.NextPage(ctx) {
		page := providerList.PageResponse()
		providers = append(providers, page.ProviderListResult.Value...)
	}

	return providers
}

func getAtTenantScopeProvider(ctx context.Context, cred azcore.TokenCredential) (*armresources.Provider, error) {
	providerClient := armresources.NewProvidersClient(subscriptionID, cred, nil)

	providerResp, err := providerClient.GetAtTenantScope(ctx, resourceProviderNamespace, nil)
	if err != nil {
		return nil, err
	}

	return &providerResp.Provider, nil
}

func listAtTenantScopeProvider(ctx context.Context, cred azcore.TokenCredential) []*armresources.Provider {
	providerClient := armresources.NewProvidersClient(subscriptionID, cred, nil)

	providerList := providerClient.ListAtTenantScope(&armresources.ProvidersClientListAtTenantScopeOptions{})

	var providers = make([]*armresources.Provider, 0)
	for providerList.NextPage(ctx) {
		pageResp := providerList.PageResponse()

		providers = append(providers, pageResp.ProviderListResult.Value...)
	}

	return providers
}

func providerPermissions(ctx context.Context, cred azcore.TokenCredential) (*armresources.ProviderPermissionListResult, error) {
	providerClient := armresources.NewProvidersClient(subscriptionID, cred, nil)

	providerPermissionsResp, err := providerClient.ProviderPermissions(ctx, resourceProviderNamespace, nil)
	if err != nil {
		return nil, err
	}

	return &providerPermissionsResp.ProviderPermissionListResult, nil
}
