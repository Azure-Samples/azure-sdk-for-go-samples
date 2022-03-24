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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/containerregistry/armcontainerregistry"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "eastus"
	resourceGroupName = "sample-resource-group"
	registryName      = "sample2registry"
	agentPoolName     = "sample-agent-pool"
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

	agentPool, err := createAgentPool(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("agent pool:", *agentPool.ID)

	agentPool, err = getAgentPool(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get agent pool:", *agentPool.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createRegistry(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.Registry, error) {
	registriesClient := armcontainerregistry.NewRegistriesClient(subscriptionID, cred, nil)

	pollerResp, err := registriesClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		armcontainerregistry.Registry{
			Location: to.StringPtr(location),
			Tags: map[string]*string{
				"key": to.StringPtr("value"),
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

func createAgentPool(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.AgentPool, error) {
	agentPoolsClient := armcontainerregistry.NewAgentPoolsClient(subscriptionID, cred, nil)

	pollerResp, err := agentPoolsClient.BeginCreate(
		ctx,
		resourceGroupName,
		registryName,
		agentPoolName,
		armcontainerregistry.AgentPool{
			Location: to.StringPtr(location),
			Tags: map[string]*string{
				"key": to.StringPtr("value"),
			},
			Properties: &armcontainerregistry.AgentPoolProperties{
				Count: to.Int32Ptr(1),
				OS:    armcontainerregistry.OSLinux.ToPtr(),
				Tier:  to.StringPtr("S1"),
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
	return &resp.AgentPool, nil
}

func getAgentPool(ctx context.Context, cred azcore.TokenCredential) (*armcontainerregistry.AgentPool, error) {
	agentPoolsClient := armcontainerregistry.NewAgentPoolsClient(subscriptionID, cred, nil)

	resp, err := agentPoolsClient.Get(ctx, resourceGroupName, registryName, agentPoolName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.AgentPool, nil
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
