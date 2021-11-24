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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
)

var (
	subscriptionID        string
	location              = "westus"
	resourceGroupName     = "sample-resource-group"
	virtualNetworkName    = "sample-virtual-network"
	subnetName            = "sample-subnet"
	namespaceName         = "sample-sb-namespace"
	authorizationRuleName = "sample-sb-authorization-rule"
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

	namespace, err := createNamespace(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus namespace:", *namespace.ID)

	namespaceAuthorizationRule, err := createNamespaceAuthorizationRule(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus namespace authorization rule:", *namespaceAuthorizationRule.ID)

	namespaceNetworkRuleSet, err := createNamespaceNetworkRuleSet(ctx, cred, *subnet.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus namespace network rule set:", *namespaceNetworkRuleSet.ID)

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

func createNamespace(ctx context.Context, cred azcore.TokenCredential) (np armservicebus.NamespacesCreateOrUpdateResponse, err error) {
	namespacesClient := armservicebus.NewNamespacesClient(subscriptionID, cred, nil)

	pollerResp, err := namespacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		namespaceName,
		armservicebus.SBNamespace{
			TrackedResource: armservicebus.TrackedResource{
				Location: to.StringPtr(location),
			},
			SKU: &armservicebus.SBSKU{
				Name: armservicebus.SKUNamePremium.ToPtr(),
				Tier: armservicebus.SKUTierPremium.ToPtr(),
			},
		},
		nil,
	)
	if err != nil {
		return np, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return np, err
	}
	return resp, nil
}

func createNamespaceAuthorizationRule(ctx context.Context, cred azcore.TokenCredential) (ar armservicebus.NamespacesCreateOrUpdateAuthorizationRuleResponse, err error) {
	namespacesClient := armservicebus.NewNamespacesClient(subscriptionID, cred, nil)

	resp, err := namespacesClient.CreateOrUpdateAuthorizationRule(
		ctx,
		resourceGroupName,
		namespaceName,
		authorizationRuleName,
		armservicebus.SBAuthorizationRule{
			Properties: &armservicebus.SBAuthorizationRuleProperties{
				Rights: []*armservicebus.AccessRights{
					armservicebus.AccessRightsListen.ToPtr(),
					armservicebus.AccessRightsSend.ToPtr(),
				},
			},
		},
		nil,
	)
	if err != nil {
		return ar, err
	}

	return resp, nil
}

func createNamespaceNetworkRuleSet(ctx context.Context, cred azcore.TokenCredential, subnetID string) (*armservicebus.NetworkRuleSet, error) {
	namespacesClient := armservicebus.NewNamespacesClient(subscriptionID, cred, nil)

	resp, err := namespacesClient.CreateOrUpdateNetworkRuleSet(
		ctx,
		resourceGroupName,
		namespaceName,
		armservicebus.NetworkRuleSet{
			Properties: &armservicebus.NetworkRuleSetProperties{
				DefaultAction: armservicebus.DefaultActionDeny.ToPtr(),
				VirtualNetworkRules: []*armservicebus.NWRuleSetVirtualNetworkRules{
					{
						Subnet: &armservicebus.Subnet{
							ID: to.StringPtr(subnetID),
						},
						IgnoreMissingVnetServiceEndpoint: to.BoolPtr(true),
					},
				},
				IPRules: []*armservicebus.NWRuleSetIPRules{
					{
						Action: armservicebus.NetworkRuleIPActionAllow.ToPtr(),
						IPMask: to.StringPtr("1.1.1.1"),
					},
					{
						Action: armservicebus.NetworkRuleIPActionAllow.ToPtr(),
						IPMask: to.StringPtr("1.1.1.2"),
					}, {
						Action: armservicebus.NetworkRuleIPActionAllow.ToPtr(),
						IPMask: to.StringPtr("1.1.1.3"),
					}, {
						Action: armservicebus.NetworkRuleIPActionAllow.ToPtr(),
						IPMask: to.StringPtr("1.1.1.4"),
					}, {
						Action: armservicebus.NetworkRuleIPActionAllow.ToPtr(),
						IPMask: to.StringPtr("1.1.1.5"),
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.NetworkRuleSet, nil
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
