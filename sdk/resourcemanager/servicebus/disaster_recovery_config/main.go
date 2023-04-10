// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/network/armnetwork/v2"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/servicebus/armservicebus"
	"log"
	"os"
)

var (
	subscriptionID             string
	location                   = "westus"
	resourceGroupName          = "sample-resource-group"
	virtualNetworkName         = "sample-virtual-network"
	subnetName                 = "sample-subnet"
	namespaceName              = "sample-sb-namespace"
	namespacePrimaryName       = "sample-sb-namespace-primary"
	authorizationRuleName      = "sample-sb-authorization-rule"
	disasterRecoveryConfigName = "sample-disaster-recovery"
)

var (
	resourcesClientFactory  *armresources.ClientFactory
	networkClientFactory    *armnetwork.ClientFactory
	servicebusClientFactory *armservicebus.ClientFactory
)

var (
	resourceGroupClient           *armresources.ResourceGroupsClient
	virtualNetworksClient         *armnetwork.VirtualNetworksClient
	subnetsClient                 *armnetwork.SubnetsClient
	namespacesClient              *armservicebus.NamespacesClient
	disasterRecoveryConfigsClient *armservicebus.DisasterRecoveryConfigsClient
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

	servicebusClientFactory, err = armservicebus.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	namespacesClient = servicebusClientFactory.NewNamespacesClient()
	disasterRecoveryConfigsClient = servicebusClientFactory.NewDisasterRecoveryConfigsClient()

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

	namespace, err := createNamespace(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus namespace:", *namespace.ID)

	namespacePrimary, err := createNamespacePrimary(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus namespace primary:", *namespacePrimary.ID)

	namespaceAuthorizationRule, err := createNamespaceAuthorizationRule(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus namespace authorization rule:", *namespaceAuthorizationRule.ID)

	exist, err := checkNameAvailability(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if exist {
		log.Println("disaster recovery name existed.")
	}

	disasterRecoveryConfig, err := createDisasterRecoveryConfig(ctx, cred, *namespacePrimary.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus disaster recovery config:", *disasterRecoveryConfig.ID)

	disasterRecoveryConfig, err = getDisasterRecoveryConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get service bus disaster recovery config:", *disasterRecoveryConfig.ID)

	err = failOverDisasterRecoveryConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus fail over")

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

func createNamespace(ctx context.Context) (*armservicebus.SBNamespace, error) {

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

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.SBNamespace, nil
}

func createNamespacePrimary(ctx context.Context) (*armservicebus.SBNamespace, error) {

	pollerResp, err := namespacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		namespacePrimaryName,
		armservicebus.SBNamespace{
			Location: to.Ptr("eastus"),
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

	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.SBNamespace, nil
}

func createNamespaceAuthorizationRule(ctx context.Context) (*armservicebus.SBAuthorizationRule, error) {

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

func checkNameAvailability(ctx context.Context) (bool, error) {

	resp, err := disasterRecoveryConfigsClient.CheckNameAvailability(
		ctx,
		resourceGroupName,
		namespaceName,
		armservicebus.CheckNameAvailability{
			Name: to.Ptr(disasterRecoveryConfigName),
		},
		nil,
	)
	if err != nil {
		return false, nil
	}
	return *resp.NameAvailable, nil
}

func createDisasterRecoveryConfig(ctx context.Context, cred azcore.TokenCredential, secondNamespaceID string) (*armservicebus.ArmDisasterRecovery, error) {

	resp, err := disasterRecoveryConfigsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		namespaceName,
		disasterRecoveryConfigName,
		armservicebus.ArmDisasterRecovery{
			Properties: &armservicebus.ArmDisasterRecoveryProperties{
				PartnerNamespace: to.Ptr(secondNamespaceID),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.ArmDisasterRecovery, nil
}

func getDisasterRecoveryConfig(ctx context.Context) (*armservicebus.ArmDisasterRecovery, error) {

	resp, err := disasterRecoveryConfigsClient.Get(ctx, resourceGroupName, namespaceName, disasterRecoveryConfigName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.ArmDisasterRecovery, nil
}

func failOverDisasterRecoveryConfig(ctx context.Context) error {

	_, err := disasterRecoveryConfigsClient.FailOver(ctx, resourceGroupName, namespacePrimaryName, disasterRecoveryConfigName, nil)
	if err != nil {
		return err
	}

	return nil
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
