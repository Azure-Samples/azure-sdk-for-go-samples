// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/redis/armredis"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID     string
	location           = "eastus"
	resourceGroupName  = "sample-resource-group"
	virtualNetworkName = "sample-virtual-network"
	subnetName         = "sample-subnet"
	redisName          = "sample-redis"
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

	subnet, err := createSubnet(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("subnet:", *subnet.ID)

	redis, err := createRedis(ctx, cred, *subnet.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("redis:", *redis.ID)

	patch, err := createPatchSchedule(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("patch schedule:", *patch.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createVirtualNetwork(ctx context.Context, cred azcore.TokenCredential) (*armnetwork.VirtualNetwork, error) {
	virtualNetworkClient, err := armnetwork.NewVirtualNetworksClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := virtualNetworkClient.BeginCreateOrUpdate(
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

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.VirtualNetwork, nil
}

func createSubnet(ctx context.Context, cred azcore.TokenCredential) (*armnetwork.Subnet, error) {
	subnetsClient, err := armnetwork.NewSubnetsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.Subnet, nil
}

func createRedis(ctx context.Context, cred azcore.TokenCredential, subnetID string) (*armredis.ResourceInfo, error) {
	redisClient, err := armredis.NewClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.ResourceInfo, nil
}

func createPatchSchedule(ctx context.Context, cred azcore.TokenCredential) (*armredis.PatchSchedule, error) {
	patchSchedulesClient, err := armredis.NewPatchSchedulesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := patchSchedulesClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		redisName,
		armredis.DefaultNameDefault,
		armredis.PatchSchedule{
			Properties: &armredis.ScheduleEntries{
				ScheduleEntries: []*armredis.ScheduleEntry{
					{
						DayOfWeek:         to.Ptr(armredis.DayOfWeekMonday),
						StartHourUTC:      to.Ptr[int32](12),
						MaintenanceWindow: to.Ptr("PT5H"),
					},
					{
						DayOfWeek:    to.Ptr(armredis.DayOfWeekTuesday),
						StartHourUTC: to.Ptr[int32](12),
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.PatchSchedule, nil
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

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

func cleanup(ctx context.Context, cred azcore.TokenCredential) error {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return err
	}
	return nil
}
