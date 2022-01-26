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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
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
		_, err := cleanup(ctx, cred)
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
	resourceClient := armresources.NewClient(subscriptionID, cred, nil)

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
	resourceClient := armresources.NewClient(subscriptionID, cred, nil)

	pollerResp, err := resourceClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		resourceProviderNamespace,
		"/",
		resourceType,
		virtualNetworkName,
		apiVersion,
		armresources.GenericResource{
			Location: to.StringPtr(location),
			Properties: map[string]interface{}{
				"addressSpace": armnetwork.AddressSpace{
					AddressPrefixes: []*string{
						to.StringPtr("10.1.0.0/16"),
					},
				},
			},
		},
		nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return &resp.GenericResource, nil
}

func getResource(ctx context.Context, cred azcore.TokenCredential) (*armresources.GenericResource, error) {
	resourceClient := armresources.NewClient(subscriptionID, cred, nil)

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
