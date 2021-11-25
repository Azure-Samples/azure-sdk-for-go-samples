package main

import (
	"context"
	"encoding/json"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/authorization/armauthorization"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID            string
	location                  = "westus"
	resourceGroupName         = "sample-resource-group"
	virtualNetworkName        = "sample-virtual-network"
	resourceProviderNamespace = "Microsoft.Network"
	resourceType              = "virtualNetworks"
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

	virtualNetwork, err := createVirtualNetwork(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("virtual network:", *virtualNetwork.ID)

	permissions := listPermissionForResource(ctx, cred)
	data, err := json.Marshal(permissions)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(data))

	permissions = listPermissionForResourceGroup(ctx, cred)
	data, err = json.Marshal(permissions)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(string(data))

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createVirtualNetwork(ctx context.Context, cred azcore.TokenCredential) (*armnetwork.VirtualNetwork, error) {
	virtualNetworkClient := armnetwork.NewVirtualNetworksClient(subscriptionID, cred, nil)

	pollerResp, err := virtualNetworkClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		virtualNetworkName,
		armnetwork.VirtualNetwork{
			Resource: armnetwork.Resource{
				Location: to.StringPtr(location),
			},
			Properties: &armnetwork.VirtualNetworkPropertiesFormat{
				AddressSpace: &armnetwork.AddressSpace{
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
	return &resp.VirtualNetwork, nil
}

func listPermissionForResource(ctx context.Context, cred azcore.TokenCredential) []*armauthorization.Permission {
	permissionsClient := armauthorization.NewPermissionsClient(subscriptionID, cred, nil)

	permissionPager := permissionsClient.ListForResource(
		resourceGroupName,
		resourceProviderNamespace,
		"",
		resourceType,
		virtualNetworkName,
		nil,
	)

	permissions := make([]*armauthorization.Permission, 0)
	for permissionPager.NextPage(ctx) {
		pageResp := permissionPager.PageResponse()
		permissions = append(permissions, pageResp.Value...)
	}
	return permissions
}

func listPermissionForResourceGroup(ctx context.Context, cred azcore.TokenCredential) []*armauthorization.Permission {
	permissionsClient := armauthorization.NewPermissionsClient(subscriptionID, cred, nil)

	permissionPager := permissionsClient.ListForResourceGroup(resourceGroupName, nil)

	permissions := make([]*armauthorization.Permission, 0)
	for permissionPager.NextPage(ctx) {
		pageResp := permissionPager.PageResponse()
		permissions = append(permissions, pageResp.Value...)
	}
	return permissions
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
