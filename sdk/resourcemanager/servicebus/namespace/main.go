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

func createNamespace(ctx context.Context, cred azcore.TokenCredential) (*armservicebus.SBNamespace, error) {
	namespacesClient, err := armservicebus.NewNamespacesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := namespacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		namespaceName,
		armservicebus.SBNamespace{
			Location: to.Ptr(location),
			SKU: &armservicebus.SBSKU{
				Name: to.Ptr(armservicebus.SKUNamePremium),
				Tier: to.Ptr(armservicebus.SKUTierPremium),
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
	return &resp.SBNamespace, nil
}

func createNamespaceAuthorizationRule(ctx context.Context, cred azcore.TokenCredential) (*armservicebus.SBAuthorizationRule, error) {
	namespacesClient, err := armservicebus.NewNamespacesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := namespacesClient.CreateOrUpdateAuthorizationRule(
		ctx,
		resourceGroupName,
		namespaceName,
		authorizationRuleName,
		armservicebus.SBAuthorizationRule{
			Properties: &armservicebus.SBAuthorizationRuleProperties{
				Rights: []*armservicebus.AccessRights{
					to.Ptr(armservicebus.AccessRightsListen),
					to.Ptr(armservicebus.AccessRightsSend),
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.SBAuthorizationRule, nil
}

func createNamespaceNetworkRuleSet(ctx context.Context, cred azcore.TokenCredential, subnetID string) (*armservicebus.NetworkRuleSet, error) {
	namespacesClient, err := armservicebus.NewNamespacesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := namespacesClient.CreateOrUpdateNetworkRuleSet(
		ctx,
		resourceGroupName,
		namespaceName,
		armservicebus.NetworkRuleSet{
			Properties: &armservicebus.NetworkRuleSetProperties{
				DefaultAction: to.Ptr(armservicebus.DefaultActionDeny),
				VirtualNetworkRules: []*armservicebus.NWRuleSetVirtualNetworkRules{
					{
						Subnet: &armservicebus.Subnet{
							ID: to.Ptr(subnetID),
						},
						IgnoreMissingVnetServiceEndpoint: to.Ptr(true),
					},
				},
				IPRules: []*armservicebus.NWRuleSetIPRules{
					{
						Action: to.Ptr(armservicebus.NetworkRuleIPActionAllow),
						IPMask: to.Ptr("1.1.1.1"),
					},
					{
						Action: to.Ptr(armservicebus.NetworkRuleIPActionAllow),
						IPMask: to.Ptr("1.1.1.2"),
					}, {
						Action: to.Ptr(armservicebus.NetworkRuleIPActionAllow),
						IPMask: to.Ptr("1.1.1.3"),
					}, {
						Action: to.Ptr(armservicebus.NetworkRuleIPActionAllow),
						IPMask: to.Ptr("1.1.1.4"),
					}, {
						Action: to.Ptr(armservicebus.NetworkRuleIPActionAllow),
						IPMask: to.Ptr("1.1.1.5"),
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
