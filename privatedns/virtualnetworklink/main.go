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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/privatedns/armprivatedns"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID         string
	location               = "westus"
	resourceGroupName      = "sample-resource-group"
	virtualNetworkName     = "sample-virtual-network"
	subnetName             = "sample-subnet"
	privateZoneName        = "sample-private-zone"
	virtualNetworkLinkName = "sample-private-dns-link"
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

	privateZone, err := createPrivateZone(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("private zone:", *privateZone.ID)

	virtualNetwork, err := createVirtualNetwork(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("virtual network:", *virtualNetwork.ID)

	subnet, err := createSubnet(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("subnet:", *subnet.ID)

	vnetLink, err := createVirtualNetworkLink(ctx, conn, *subnet.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("virtual network link:", *vnetLink.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createPrivateZone(ctx context.Context, conn *arm.Connection) (*armprivatedns.PrivateZone, error) {
	privateZonesClient := armprivatedns.NewPrivateZonesClient(conn, subscriptionID)

	pollersResp, err := privateZonesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		privateZoneName,
		armprivatedns.PrivateZone{
			TrackedResource: armprivatedns.TrackedResource{
				Location: to.StringPtr(location),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollersResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.PrivateZone, nil
}

func createVirtualNetwork(ctx context.Context, conn *arm.Connection) (*armnetwork.VirtualNetwork, error) {
	virtualNetworkClient := armnetwork.NewVirtualNetworksClient(conn, subscriptionID)

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

func createSubnet(ctx context.Context, conn *arm.Connection) (*armnetwork.Subnet, error) {
	subnetsClient := armnetwork.NewSubnetsClient(conn, subscriptionID)

	pollerResp, err := subnetsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		virtualNetworkName,
		subnetName,
		armnetwork.Subnet{
			Properties: &armnetwork.SubnetPropertiesFormat{
				AddressPrefix: to.StringPtr("10.1.0.0/24"),
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
	return &resp.Subnet, nil
}

func createVirtualNetworkLink(ctx context.Context, conn *arm.Connection, subnetID string) (*armprivatedns.TrackedResource, error) {
	vnetLinksClient := armprivatedns.NewVirtualNetworkLinksClient(conn, subscriptionID)

	pollersResp, err := vnetLinksClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		privateZoneName,
		virtualNetworkLinkName,
		armprivatedns.VirtualNetworkLink{
			TrackedResource: armprivatedns.TrackedResource{
				Location: to.StringPtr(location),
			},
			Properties: &armprivatedns.VirtualNetworkLinkProperties{
				RegistrationEnabled: to.BoolPtr(true),
				VirtualNetwork: &armprivatedns.SubResource{
					ID: to.StringPtr(subnetID),
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollersResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.TrackedResource, nil
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
