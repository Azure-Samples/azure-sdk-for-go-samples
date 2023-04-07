// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID            string
	resourceProviderNamespace = "Microsoft.Compute"
)

var (
	resourcesClientFactory *armresources.ClientFactory
)

var (
	providersClient *armresources.ProvidersClient
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
	providersClient = resourcesClientFactory.NewProvidersClient()

	provider, err := registerProvider(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("registered provider:", *provider.ID)

	provider, err = getProvider(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get provider:", *provider.ID)

	providers, err := listProvider(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("providers:", len(providers))
	l := math.Min(10, float64(len(providers)))
	for i := 0; i < int(l); i++ {
		log.Printf("Namespace: %s,RegistratonState: %s\n", *providers[i].Namespace, *providers[i].RegistrationState)
	}

	providerPermissionsResult, err := providerPermissions(ctx)
	if err != nil {
		log.Fatal(err)
	}
	data, err := json.Marshal(providerPermissionsResult)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(data))

	// Tenant
	providers, err = listAtTenantScopeProvider(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("list providers:", len(providers))
	l = math.Min(10, float64(len(providers)))
	for i := 0; i < int(l); i++ {
		log.Println("Namespace:", *providers[i].Namespace)
	}

	atTenant, err := getAtTenantScopeProvider(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get atTenant:", *atTenant.Namespace)
}

func registerProvider(ctx context.Context) (*armresources.Provider, error) {

	providerResp, err := providersClient.Register(ctx, resourceProviderNamespace, nil)
	if err != nil {
		return nil, err
	}

	return &providerResp.Provider, nil
}

func getProvider(ctx context.Context) (*armresources.Provider, error) {

	providerResp, err := providersClient.Get(ctx, resourceProviderNamespace, nil)
	if err != nil {
		return nil, err
	}

	return &providerResp.Provider, nil
}

func listProvider(ctx context.Context) ([]*armresources.Provider, error) {

	providerList := providersClient.NewListPager(nil)

	var providers = make([]*armresources.Provider, 0)
	for providerList.More() {
		page, err := providerList.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		providers = append(providers, page.ProviderListResult.Value...)
	}

	return providers, nil
}

func getAtTenantScopeProvider(ctx context.Context) (*armresources.Provider, error) {

	providerResp, err := providersClient.GetAtTenantScope(ctx, resourceProviderNamespace, nil)
	if err != nil {
		return nil, err
	}

	return &providerResp.Provider, nil
}

func listAtTenantScopeProvider(ctx context.Context) ([]*armresources.Provider, error) {

	providerList := providersClient.NewListAtTenantScopePager(&armresources.ProvidersClientListAtTenantScopeOptions{})
	var providers = make([]*armresources.Provider, 0)
	for providerList.More() {
		pageResp, err := providerList.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		providers = append(providers, pageResp.ProviderListResult.Value...)
	}

	return providers, nil
}

func providerPermissions(ctx context.Context) (*armresources.ProviderPermissionListResult, error) {

	providerPermissionsResp, err := providersClient.ProviderPermissions(ctx, resourceProviderNamespace, nil)
	if err != nil {
		return nil, err
	}

	return &providerPermissionsResp.ProviderPermissionListResult, nil
}
