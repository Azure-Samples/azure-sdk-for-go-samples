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

	namespace, err := createNamespace(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus namespace:", *namespace.ID)

	namespacePrimary, err := createNamespacePrimary(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus namespace primary:", *namespacePrimary.ID)

	namespaceAuthorizationRule, err := createNamespaceAuthorizationRule(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus namespace authorization rule:", *namespaceAuthorizationRule.ID)

	exist := checkNameAvailability(ctx, conn)
	if exist {
		log.Println("disaster recovery name existed.")
	}

	disasterRecoveryConfig, err := createDisasterRecoveryConfig(ctx, conn, *namespacePrimary.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus disaster recovery config:", *disasterRecoveryConfig.ID)

	disasterRecoveryConfig, err = getDisasterRecoveryConfig(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	count := 0
	for disasterRecoveryConfig.Properties.ProvisioningState != armservicebus.ProvisioningStateDRSucceeded.ToPtr() && count < 10 {
		time.Sleep(30)
		disasterRecoveryConfig, err = getDisasterRecoveryConfig(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		count++
	}
	log.Println("get service bus disaster recovery config:", *disasterRecoveryConfig.ID, count)

	resp, err := failOverDisasterRecoveryConfig(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("service bus fail over:", resp)

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

func createNamespace(ctx context.Context, conn *arm.Connection) (np armservicebus.NamespacesCreateOrUpdateResponse, err error) {
	namespacesClient := armservicebus.NewNamespacesClient(conn, subscriptionID)

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

func createNamespacePrimary(ctx context.Context, conn *arm.Connection) (np armservicebus.NamespacesCreateOrUpdateResponse, err error) {
	namespacesClient := armservicebus.NewNamespacesClient(conn, subscriptionID)

	pollerResp, err := namespacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		namespacePrimaryName,
		armservicebus.SBNamespace{
			TrackedResource: armservicebus.TrackedResource{
				Location: to.StringPtr("eastus"),
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

func createNamespaceAuthorizationRule(ctx context.Context, conn *arm.Connection) (ar armservicebus.NamespacesCreateOrUpdateAuthorizationRuleResponse, err error) {
	namespacesClient := armservicebus.NewNamespacesClient(conn, subscriptionID)

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

func checkNameAvailability(ctx context.Context, conn *arm.Connection) bool {
	namespacesClient := armservicebus.NewDisasterRecoveryConfigsClient(conn, subscriptionID)

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

func createDisasterRecoveryConfig(ctx context.Context, conn *arm.Connection, secondNamespaceID string) (*armservicebus.ArmDisasterRecovery, error) {
	disasterRecoveryConfigsClient := armservicebus.NewDisasterRecoveryConfigsClient(conn, subscriptionID)

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

func getDisasterRecoveryConfig(ctx context.Context, conn *arm.Connection) (*armservicebus.ArmDisasterRecovery, error) {
	disasterRecoveryConfigsClient := armservicebus.NewDisasterRecoveryConfigsClient(conn, subscriptionID)

	resp, err := disasterRecoveryConfigsClient.Get(ctx, resourceGroupName, namespaceName, disasterRecoveryConfigName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.ArmDisasterRecovery, nil
}

func failOverDisasterRecoveryConfig(ctx context.Context, conn *arm.Connection) (*http.Response, error) {
	disasterRecoveryConfigsClient := armservicebus.NewDisasterRecoveryConfigsClient(conn, subscriptionID)

	resp, err := disasterRecoveryConfigsClient.FailOver(ctx, resourceGroupName, namespacePrimaryName, disasterRecoveryConfigName, nil)
	if err != nil {
		return nil, err
	}

	return resp.RawResponse, nil
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
