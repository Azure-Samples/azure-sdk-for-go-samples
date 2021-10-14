package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	registryName      = "sample2registry"
	scopeMapName      = "sample-scope-map"
	tokenName         = "sample-token"
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

	resourceGroup, err := createResourceGroup(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	registry, err := createRegistry(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("registry:", *registry.ID)

	scopeMap, err := createScopeMap(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("scope map:", *scopeMap.ID)

	token, err := createToken(ctx, conn, *scopeMap.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("token:", *token.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createRegistry(ctx context.Context, conn *arm.Connection) (*armcontainerregistry.Registry, error) {
	registriesClient := armcontainerregistry.NewRegistriesClient(conn, subscriptionID)

	pollerResp, err := registriesClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		armcontainerregistry.Registry{
			Resource: armcontainerregistry.Resource{
				Location: to.StringPtr(location),
				Tags: map[string]*string{
					"key": to.StringPtr("value"),
				},
			},
			SKU: &armcontainerregistry.SKU{
				Name: armcontainerregistry.SKUNamePremium.ToPtr(),
			},
			Properties: &armcontainerregistry.RegistryProperties{
				AdminUserEnabled: to.BoolPtr(true),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.Registry, nil
}

func createScopeMap(ctx context.Context, conn *arm.Connection) (*armcontainerregistry.ScopeMap, error) {
	scopeMapsClient := armcontainerregistry.NewScopeMapsClient(conn, subscriptionID)

	pollerResp, err := scopeMapsClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		scopeMapName,
		armcontainerregistry.ScopeMap{
			Properties: &armcontainerregistry.ScopeMapProperties{
				Actions: []*string{
					to.StringPtr("repositories/foo/content/read"),
					to.StringPtr("repositories/foo/content/delete"),
				},
				Description: to.StringPtr("Developer Scopes"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.ScopeMap, nil
}

func createToken(ctx context.Context, conn *arm.Connection, scopeMapID string) (*armcontainerregistry.Token, error) {
	tokensClient := armcontainerregistry.NewTokensClient(conn, subscriptionID)

	pollerResp, err := tokensClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		tokenName,
		armcontainerregistry.Token{
			Properties: &armcontainerregistry.TokenProperties{
				ScopeMapID: to.StringPtr(scopeMapID),
				Status:     armcontainerregistry.TokenStatusEnabled.ToPtr(),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.Token, nil
}

func createResourceGroup(ctx context.Context, conn *arm.Connection) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

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

func cleanup(ctx context.Context, conn *arm.Connection) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

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
