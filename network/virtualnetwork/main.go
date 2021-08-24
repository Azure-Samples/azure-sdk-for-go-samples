package main

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/armcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/to"
	"log"
	"net/http"
	"os"
	"time"
)

var subscriptionID string
var location = "westus"
var resourceGroupName = "sample-resource-group"
var virtualNetworkName = "sample-virtual-network"

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}

	conn := armcore.NewDefaultConnection(cred, &armcore.ConnectionOptions{
		Logging: azcore.LogOptions{
			IncludeBody: true,
		},
	})
	ctx := context.Background()

	resourceGroup,err := createResourceGroup(ctx,conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resource group:",*resourceGroup.ID)

	virtualNetwork,err := createVirtualNetwork(ctx,conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("virtual network:",*virtualNetwork.ID)

	virtualNetwork,err = createVirtualNetworkAndSubnets(ctx,conn)
	if err != nil {
		log.Fatal(err)
	}
	subnets := virtualNetwork.Properties.Subnets
	log.Println("virtual network and subnets:")
	for _,sub:=range subnets {
		log.Println("\t",*sub.ID)
	}

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_,err := cleanup(ctx,conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createVirtualNetwork(ctx context.Context,conn *armcore.Connection) (*armnetwork.VirtualNetwork,error) {
	virtualNetworkClient := armnetwork.NewVirtualNetworksClient(conn,subscriptionID)

	pollerResp,err := virtualNetworkClient.BeginCreateOrUpdate(
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
		return nil,err
	}

	resp,err := pollerResp.PollUntilDone(ctx, 10 * time.Second)
	if err != nil {
		return nil,err
	}
	return resp.VirtualNetwork,nil
}

func createVirtualNetworkAndSubnets(ctx context.Context,conn *armcore.Connection) (*armnetwork.VirtualNetwork, error) {
	virtualNetworkClient := armnetwork.NewVirtualNetworksClient(conn,subscriptionID)

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
						to.StringPtr("10.0.0.0/8"),
					},
				},
				Subnets: []*armnetwork.Subnet{
					{
						Name: to.StringPtr("sample-subnet-0"),
						Properties: &armnetwork.SubnetPropertiesFormat{
							AddressPrefix: to.StringPtr("10.0.0.0/16"),
						},
					},
					{
						Name: to.StringPtr("sample-subnet-1"),
						Properties: &armnetwork.SubnetPropertiesFormat{
							AddressPrefix: to.StringPtr("10.1.0.0/16"),
						},
					},
				},
			},
		},
		nil)

	if err != nil {
		return nil, fmt.Errorf("cannot create virtual network: %v", err)
	}

	resp, err := pollerResp.PollUntilDone(ctx,10 * time.Second)
	if err != nil {
		return nil, err
	}

	return resp.VirtualNetwork,nil
}

func createResourceGroup(ctx context.Context,conn *armcore.Connection) (*armresources.ResourceGroup,error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn,subscriptionID)

	resourceGroupResp,err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.StringPtr(location),
		},
		nil)
	if err != nil {
		return nil,err
	}
	return resourceGroupResp.ResourceGroup,nil
}

func cleanup(ctx context.Context, conn *armcore.Connection) (*http.Response,error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn,subscriptionID)

	pollerResp,err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return nil,err
	}

	resp,err := pollerResp.PollUntilDone(ctx, 10 * time.Second)
	if err != nil {
		return nil, err
	}
	return resp,nil
}
