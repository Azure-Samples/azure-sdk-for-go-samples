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
	ruleName           = "sample2rule"
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

	firewallRule, err := createFireWallRule(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("firewall rule:", *firewallRule.ID)

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

func createSubnet(ctx context.Context, cred azcore.TokenCredential) (*armnetwork.Subnet, error) {
	subnetsClient := armnetwork.NewSubnetsClient(subscriptionID, cred, nil)

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

func createRedis(ctx context.Context, cred azcore.TokenCredential, subnetID string) (r armredis.RedisCreateResponse, err error) {
	redisClient := armredis.NewRedisClient(subscriptionID, cred, nil)

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

func createFireWallRule(ctx context.Context, cred azcore.TokenCredential) (r armredis.FirewallRulesCreateOrUpdateResponse, err error) {
	firewallRulesClient := armredis.NewFirewallRulesClient(subscriptionID, cred, nil)

	resp, err := firewallRulesClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		redisName,
		ruleName,
		armredis.RedisFirewallRule{
			Properties: &armredis.RedisFirewallRuleProperties{
				StartIP: to.StringPtr("10.0.1.1"),
				EndIP:   to.StringPtr("10.0.1.4"),
			},
		},
		nil,
	)
	if err != nil {
		return r, err
	}

	return resp, nil
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
