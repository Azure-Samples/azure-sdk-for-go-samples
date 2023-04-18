// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID     string
	location           = "eastus"
	resourceGroupName  = "sample-resource-group"
	virtualNetworkName = "sample-virtual-network"
	subnetName         = "sample-subnet"
	redisName          = "sample-redis"
)

var (
	resourcesClientFactory *armresources.ClientFactory
	networkClientFactory   *armnetwork.ClientFactory
	redisClientFactory     *armredis.ClientFactory
)

var (
	resourceGroupClient   *armresources.ResourceGroupsClient
	virtualNetworksClient *armnetwork.VirtualNetworksClient
	subnetsClient         *armnetwork.SubnetsClient
	redisClient           *armredis.Client
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

	networkClientFactory, err = armnetwork.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	virtualNetworksClient = networkClientFactory.NewVirtualNetworksClient()
	subnetsClient = networkClientFactory.NewSubnetsClient()

	redisClientFactory, err = armredis.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	redisClient = redisClientFactory.NewClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	virtualNetwork, err := createVirtualNetwork(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("virtual network:", *virtualNetwork.ID)

	subnet, err := createSubnet(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("subnet:", *subnet.ID)

	redis, err := createRedis(ctx, *subnet.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("redis:", *redis.ID)

	getRedis, err := getRedis(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get redis:", *getRedis.ID)

	updateRedis, err := updateRedis(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("update redis:", *updateRedis.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createVirtualNetwork(ctx context.Context) (*armnetwork.VirtualNetwork, error) {

	pollerResp, err := virtualNetworksClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		virtualNetworkName,
		armnetwork.VirtualNetwork{
			Location: to.Ptr(location),
			Properties: &armnetwork.VirtualNetworkPropertiesFormat{
				AddressSpace: &armnetwork.AddressSpace{
					AddressPrefixes: []*string{
						to.Ptr("10.0.0.0/16"),
					},
				},
			},
		},
		nil)

	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.VirtualNetwork, nil
}

func createSubnet(ctx context.Context) (*armnetwork.Subnet, error) {

	pollerResp, err := subnetsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		virtualNetworkName,
		subnetName,
		armnetwork.Subnet{
			Properties: &armnetwork.SubnetPropertiesFormat{
				AddressPrefix: to.Ptr("10.0.0.0/24"),
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
	return &resp.Subnet, nil
}

func createRedis(ctx context.Context, subnetID string) (*armredis.ResourceInfo, error) {

	pollerResp, err := redisClient.BeginCreate(
		ctx,
		resourceGroupName,
		redisName,
		armredis.CreateParameters{
			Location: to.Ptr(location),
			Zones: []*string{
				to.Ptr("1"),
			},
			Properties: &armredis.CreateProperties{
				SKU: &armredis.SKU{
					Name:     to.Ptr(armredis.SKUNamePremium),
					Family:   to.Ptr(armredis.SKUFamilyP),
					Capacity: to.Ptr[int32](1),
				},
				EnableNonSSLPort: to.Ptr(true),
				ShardCount:       to.Ptr[int32](2),
				RedisConfiguration: &armredis.CommonPropertiesRedisConfiguration{
					MaxmemoryPolicy: to.Ptr("allkeys-lru"),
				},
				MinimumTLSVersion: to.Ptr(armredis.TLSVersionOne2),
				SubnetID:          to.Ptr(subnetID),
				StaticIP:          to.Ptr("10.0.0.5"),
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
	return &resp.ResourceInfo, nil
}

func getRedis(ctx context.Context) (*armredis.ResourceInfo, error) {

	resp, err := redisClient.Get(ctx, resourceGroupName, redisName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ResourceInfo, nil
}

func updateRedis(ctx context.Context) (*armredis.ResourceInfo, error) {

	pollerResp, err := redisClient.BeginUpdate(ctx, resourceGroupName, redisName, armredis.UpdateParameters{
		Properties: &armredis.UpdateProperties{
			EnableNonSSLPort: to.Ptr(true),
		},
	}, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	
	return &resp.ResourceInfo, nil
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
