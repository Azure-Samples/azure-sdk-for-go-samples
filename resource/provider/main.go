package main

import (
	"context"
	"encoding/json"
	"log"
	"math"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
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

	conn := arm.NewDefaultConnection(cred, &arm.ConnectionOptions{
		Logging: policy.LogOptions{
			IncludeBody: true,
		},
	})
	ctx := context.Background()

	provider, err := registerProvider(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("registered provider:", *provider.ID)

	provider, err = getProvider(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get provider:", *provider.ID)

	providers := listProvider(ctx, conn)
	log.Println("providers:", len(providers))
	l := math.Min(10, float64(len(providers)))
	for i := 0; i < int(l); i++ {
		log.Printf("Namespace: %s,RegistratonState: %s\n", *providers[i].Namespace, *providers[i].RegistrationState)
	}

	providerPermissionsResult, err := providerPermissions(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	data, err := json.Marshal(providerPermissionsResult)
	log.Println(string(data))

	// Tenant
	providers = listAtTenantScopeProvider(ctx, conn)
	log.Println("list providers:", len(providers))
	l = math.Min(10, float64(len(providers)))
	for i := 0; i < int(l); i++ {
		log.Println("Namespace:", *providers[i].Namespace)
	}

	atTenant, err := getAtTenantScopeProvider(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get atTenant:", *atTenant.Namespace)
}

func registerProvider(ctx context.Context, conn *arm.Connection) (*armresources.Provider, error) {
	providerClient := armresources.NewProvidersClient(conn, subscriptionID)

	providerResp, err := providerClient.Register(ctx, resourceProviderNamespace, nil)
	if err != nil {
		return nil, err
	}

	return &providerResp.Provider, nil
}

func getProvider(ctx context.Context, conn *arm.Connection) (*armresources.Provider, error) {
	providerClient := armresources.NewProvidersClient(conn, subscriptionID)

	providerResp, err := providerClient.Get(ctx, resourceProviderNamespace, nil)
	if err != nil {
		return nil, err
	}

	return &providerResp.Provider, nil
}

func listProvider(ctx context.Context, conn *arm.Connection) []*armresources.Provider {
	providerClient := armresources.NewProvidersClient(conn, subscriptionID)

	providerList := providerClient.List(nil)

	var providers = make([]*armresources.Provider, 0)
	for providerList.NextPage(ctx) {
		page := providerList.PageResponse()
		providers = append(providers, page.ProviderListResult.Value...)
	}

	return providers
}

func getAtTenantScopeProvider(ctx context.Context, conn *arm.Connection) (*armresources.Provider, error) {
	providerClient := armresources.NewProvidersClient(conn, subscriptionID)

	providerResp, err := providerClient.GetAtTenantScope(ctx, resourceProviderNamespace, nil)
	if err != nil {
		return nil, err
	}

	return &providerResp.Provider, nil
}

func listAtTenantScopeProvider(ctx context.Context, conn *arm.Connection) []*armresources.Provider {
	providerClient := armresources.NewProvidersClient(conn, subscriptionID)

	providerList := providerClient.ListAtTenantScope(&armresources.ProvidersListAtTenantScopeOptions{})

	var providers = make([]*armresources.Provider, 0)
	for providerList.NextPage(ctx) {
		pageResp := providerList.PageResponse()

		providers = append(providers, pageResp.ProviderListResult.Value...)
	}

	return providers
}

func providerPermissions(ctx context.Context, conn *arm.Connection) (*armresources.ProviderPermissionListResult, error) {
	providerClient := armresources.NewProvidersClient(conn, subscriptionID)

	providerPermissionsResp, err := providerClient.ProviderPermissions(ctx, resourceProviderNamespace, nil)
	if err != nil {
		return nil, err
	}

	return &providerPermissionsResp.ProviderPermissionListResult, nil
}
