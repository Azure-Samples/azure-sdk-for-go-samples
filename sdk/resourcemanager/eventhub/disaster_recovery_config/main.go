// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventhub/armeventhub"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"log"
	"os"
)

var (
	subscriptionID             string
	location                   = "westus"
	resourceGroupName          = "sample-resource-group"
	namespacesName             = "sample1namespace"
	secondNamespacesName       = "sample1second1namespace"
	disasterRecoveryConfigName = "sample-disaster-recovery-config"
)

var (
	resourcesClientFactory *armresources.ClientFactory
	eventhubClientFactory  *armeventhub.ClientFactory
)

var (
	resourceGroupClient           *armresources.ResourceGroupsClient
	namespacesClient              *armeventhub.NamespacesClient
	disasterRecoveryConfigsClient *armeventhub.DisasterRecoveryConfigsClient
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

	eventhubClientFactory, err = armeventhub.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	namespacesClient = eventhubClientFactory.NewNamespacesClient()
	disasterRecoveryConfigsClient = eventhubClientFactory.NewDisasterRecoveryConfigsClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	namespace, err := createNamespace(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("eventhub namespace:", *namespace.ID)

	secondNamespace, err := createSecondNamespace(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("eventhub second namespace:", *secondNamespace.ID)

	ava, err := checkNameAva(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("check name availability:", *ava.NameAvailable)

	disasterRecoveryConfig, err := createDisasterRecoveryConfig(ctx, *secondNamespace.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("disaster recovery config:", *disasterRecoveryConfig.ID)

	disasterRecoveryConfig, err = getDisasterRecoveryConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get disaster recovery config:", *disasterRecoveryConfig.ID)

	// Only after breakPairing or failOVer can clean resource
	err = breakPairingDisasterRecoveryConfig(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("break pairing")

	//failOver, err := failOverDisasterRecoveryConfig(ctx, conn)
	//if err != nil {
	//	log.Fatal(err)
	//}
	//log.Println("fail over:", *failOver)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createNamespace(ctx context.Context) (*armeventhub.EHNamespace, error) {

	pollerResp, err := namespacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		namespacesName,
		armeventhub.EHNamespace{
			Location: to.Ptr(location),
			Tags: map[string]*string{
				"tag1": to.Ptr("value1"),
				"tag2": to.Ptr("value2"),
			},
			SKU: &armeventhub.SKU{
				Name: to.Ptr(armeventhub.SKUNameStandard),
				Tier: to.Ptr(armeventhub.SKUTierStandard),
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
	return &resp.EHNamespace, nil
}

func createSecondNamespace(ctx context.Context) (*armeventhub.EHNamespace, error) {

	pollerResp, err := namespacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		secondNamespacesName,
		armeventhub.EHNamespace{
			Location: to.Ptr("eastus"),
			Tags: map[string]*string{
				"tag1": to.Ptr("value1"),
				"tag2": to.Ptr("value2"),
			},
			SKU: &armeventhub.SKU{
				Name: to.Ptr(armeventhub.SKUNameStandard),
				Tier: to.Ptr(armeventhub.SKUTierStandard),
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
	return &resp.EHNamespace, nil
}

func checkNameAva(ctx context.Context) (*armeventhub.CheckNameAvailabilityResult, error) {

	resp, err := disasterRecoveryConfigsClient.CheckNameAvailability(
		ctx,
		resourceGroupName,
		namespacesName,
		armeventhub.CheckNameAvailabilityParameter{
			Name: to.Ptr(secondNamespacesName),
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.CheckNameAvailabilityResult, nil
}

func createDisasterRecoveryConfig(ctx context.Context, secondNamespaceID string) (*armeventhub.ArmDisasterRecovery, error) {

	resp, err := disasterRecoveryConfigsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		namespacesName,
		disasterRecoveryConfigName,
		armeventhub.ArmDisasterRecovery{
			Properties: &armeventhub.ArmDisasterRecoveryProperties{
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

func getDisasterRecoveryConfig(ctx context.Context) (*armeventhub.ArmDisasterRecovery, error) {

	resp, err := disasterRecoveryConfigsClient.Get(ctx, resourceGroupName, namespacesName, disasterRecoveryConfigName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ArmDisasterRecovery, nil
}

func breakPairingDisasterRecoveryConfig(ctx context.Context) error {

	_, err := disasterRecoveryConfigsClient.BreakPairing(ctx, resourceGroupName, namespacesName, disasterRecoveryConfigName, nil)
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
