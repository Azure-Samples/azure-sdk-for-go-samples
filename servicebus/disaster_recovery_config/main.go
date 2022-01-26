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

	namespacePrimary, err := createNamespacePrimary(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus namespace primary:", *namespacePrimary.ID)

	namespaceAuthorizationRule, err := createNamespaceAuthorizationRule(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus namespace authorization rule:", *namespaceAuthorizationRule.ID)

	exist := checkNameAvailability(ctx, cred)
	if exist {
		log.Println("disaster recovery name existed.")
	}

	disasterRecoveryConfig, err := createDisasterRecoveryConfig(ctx, cred, *namespacePrimary.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus disaster recovery config:", *disasterRecoveryConfig.ID)

	disasterRecoveryConfig, err = getDisasterRecoveryConfig(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	count := 0
	for disasterRecoveryConfig.Properties.ProvisioningState != armservicebus.ProvisioningStateDRSucceeded.ToPtr() && count < 10 {
		time.Sleep(30)
		disasterRecoveryConfig, err = getDisasterRecoveryConfig(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		count++
	}
	log.Println("get service bus disaster recovery config:", *disasterRecoveryConfig.ID, count)

	resp, err := failOverDisasterRecoveryConfig(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus fail over:", resp)

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
			Location: to.StringPtr(location),
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

func createNamespace(ctx context.Context, cred azcore.TokenCredential) (*armservicebus.SBNamespace, error) {
	namespacesClient := armservicebus.NewNamespacesClient(subscriptionID, cred, nil)

	pollerResp, err := namespacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		namespaceName,
		armservicebus.SBNamespace{
			Location: to.StringPtr(location),
			SKU: &armservicebus.SBSKU{
				Name: armservicebus.SKUNamePremium.ToPtr(),
				Tier: armservicebus.SKUTierPremium.ToPtr(),
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

func createNamespacePrimary(ctx context.Context, cred azcore.TokenCredential) (*armservicebus.SBNamespace, error) {
	namespacesClient := armservicebus.NewNamespacesClient(subscriptionID, cred, nil)

	pollerResp, err := namespacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		namespacePrimaryName,
		armservicebus.SBNamespace{
			Location: to.StringPtr("eastus"),
			SKU: &armservicebus.SBSKU{
				Name: armservicebus.SKUNamePremium.ToPtr(),
				Tier: armservicebus.SKUTierPremium.ToPtr(),
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
		return nil, err
	}

	return &resp.SBAuthorizationRule, nil
}

func checkNameAvailability(ctx context.Context, cred azcore.TokenCredential) bool {
	namespacesClient := armservicebus.NewDisasterRecoveryConfigsClient(subscriptionID, cred, nil)

	resp, err := namespacesClient.CheckNameAvailability(
		ctx,
		resourceGroupName,
		namespaceName,
		armservicebus.CheckNameAvailability{
			Name: to.StringPtr(disasterRecoveryConfigName),
		},
		nil,
	)
	if err != nil {
		return false
	}
	return *resp.NameAvailable
}

func createDisasterRecoveryConfig(ctx context.Context, cred azcore.TokenCredential, secondNamespaceID string) (*armservicebus.ArmDisasterRecovery, error) {
	disasterRecoveryConfigsClient := armservicebus.NewDisasterRecoveryConfigsClient(subscriptionID, cred, nil)

	resp, err := disasterRecoveryConfigsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		namespaceName,
		disasterRecoveryConfigName,
		armservicebus.ArmDisasterRecovery{
			Properties: &armservicebus.ArmDisasterRecoveryProperties{
				PartnerNamespace: to.StringPtr(secondNamespaceID),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.ArmDisasterRecovery, nil
}

func getDisasterRecoveryConfig(ctx context.Context, cred azcore.TokenCredential) (*armservicebus.ArmDisasterRecovery, error) {
	disasterRecoveryConfigsClient := armservicebus.NewDisasterRecoveryConfigsClient(subscriptionID, cred, nil)

	resp, err := disasterRecoveryConfigsClient.Get(ctx, resourceGroupName, namespaceName, disasterRecoveryConfigName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.ArmDisasterRecovery, nil
}

func failOverDisasterRecoveryConfig(ctx context.Context, cred azcore.TokenCredential) (*http.Response, error) {
	disasterRecoveryConfigsClient := armservicebus.NewDisasterRecoveryConfigsClient(subscriptionID, cred, nil)

	resp, err := disasterRecoveryConfigsClient.FailOver(ctx, resourceGroupName, namespacePrimaryName, disasterRecoveryConfigName, nil)
	if err != nil {
		return nil, err
	}

	return resp.RawResponse, nil
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
