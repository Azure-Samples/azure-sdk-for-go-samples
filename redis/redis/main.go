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

	redis, err := createRedis(ctx, conn, *subnet.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("redis:", *redis.ID)

	getRedis, err := getRedis(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get redis:", *getRedis.ID)

	updateRedis, err := updateRedis(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("update redis:", *updateRedis.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
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
						to.StringPtr("10.0.0.0/16"),
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
				AddressPrefix: to.StringPtr("10.0.0.0/24"),
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

func createRedis(ctx context.Context, conn *arm.Connection, subnetID string) (r armredis.RedisCreateResponse, err error) {
	redisClient := armredis.NewRedisClient(conn, subscriptionID)

	pollerResp, err := redisClient.BeginCreate(
		ctx,
		resourceGroupName,
		redisName,
		armredis.RedisCreateParameters{
			Location: to.StringPtr(location),
			Zones: []*string{
				to.StringPtr("1"),
			},
			Properties: &armredis.RedisCreateProperties{
				SKU: &armredis.SKU{
					Name:     armredis.SKUNamePremium.ToPtr(),
					Family:   armredis.SKUFamilyP.ToPtr(),
					Capacity: to.Int32Ptr(1),
				},
				RedisCommonProperties: armredis.RedisCommonProperties{
					EnableNonSSLPort: to.BoolPtr(true),
					ShardCount:       to.Int32Ptr(2),
					RedisConfiguration: &armredis.RedisCommonPropertiesRedisConfiguration{
						MaxmemoryPolicy: to.StringPtr("allkeys-lru"),
					},
					MinimumTLSVersion: armredis.TLSVersionOne2.ToPtr(),
				},
				SubnetID: to.StringPtr(subnetID),
				StaticIP: to.StringPtr("10.0.0.5"),
			},
		},
		nil,
	)
	if err != nil {
		return r, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return r, err
	}
	return resp, nil
}

func getRedis(ctx context.Context, conn *arm.Connection) (r armredis.RedisGetResponse, err error) {
	redisClient := armredis.NewRedisClient(conn, subscriptionID)

	resp, err := redisClient.Get(ctx, resourceGroupName, redisName, nil)
	if err != nil {
		return r, err
	}
	return resp, nil
}

func updateRedis(ctx context.Context, conn *arm.Connection) (r armredis.RedisUpdateResponse, err error) {
	redisClient := armredis.NewRedisClient(conn, subscriptionID)

	resp, err := redisClient.Update(ctx, resourceGroupName, redisName, armredis.RedisUpdateParameters{
		Properties: &armredis.RedisUpdateProperties{
			RedisCommonProperties: armredis.RedisCommonProperties{
				EnableNonSSLPort: to.BoolPtr(true),
			},
		},
	}, nil)
	if err != nil {
		return r, err
	}
	return resp, nil
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
