// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"log"
	"os"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/recoveryservices/armrecoveryservices"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "westus"
	resourceGroupName = "sample-resource-group"
	vaultName         = "sample-recoveryservice-vault"
)

var (
	resourcesClientFactory        *armresources.ClientFactory
	recoveryservicesClientFactory *armrecoveryservices.ClientFactory
)

var (
	resourceGroupClient     *armresources.ResourceGroupsClient
	vaultsClient            *armrecoveryservices.VaultsClient
	vaultExtendedInfoClient *armrecoveryservices.VaultExtendedInfoClient
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

	recoveryservicesClientFactory, err = armrecoveryservices.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	vaultsClient = recoveryservicesClientFactory.NewVaultsClient()
	vaultExtendedInfoClient = recoveryservicesClientFactory.NewVaultExtendedInfoClient()

	resourceGroup, err := createResourceGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	vault, err := createRecoveryServiceVault(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("recovery service vault:", *vault.ID)

	vaultExtendedInfo, err := createVaultExtendedInfo(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("recovery service vault extended info:", *vaultExtendedInfo.ID)

	vaultExtendedInfo, err = getVaultExtendedInfo(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get recovery service vault extended info:", *vaultExtendedInfo.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createRecoveryServiceVault(ctx context.Context, cred azcore.TokenCredential) (*armrecoveryservices.Vault, error) {

	pollerResp, err := vaultsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		vaultName,
		armrecoveryservices.Vault{
			Location: to.Ptr(location),
			SKU: &armrecoveryservices.SKU{
				Name: to.Ptr(armrecoveryservices.SKUNameStandard),
			},
			Properties: &armrecoveryservices.VaultProperties{},
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

	return &resp.Vault, err
}

func createVaultExtendedInfo(ctx context.Context, cred azcore.TokenCredential) (*armrecoveryservices.VaultExtendedInfoResource, error) {

	resp, err := vaultExtendedInfoClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		vaultName,
		armrecoveryservices.VaultExtendedInfoResource{
			Properties: &armrecoveryservices.VaultExtendedInfo{
				Algorithm: to.Ptr("None"),
				//IntegrityKey: to.StringPtr(),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.VaultExtendedInfoResource, err
}

func getVaultExtendedInfo(ctx context.Context, cred azcore.TokenCredential) (*armrecoveryservices.VaultExtendedInfoResource, error) {

	resp, err := vaultExtendedInfoClient.Get(ctx, resourceGroupName, vaultName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.VaultExtendedInfoResource, err
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {

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
