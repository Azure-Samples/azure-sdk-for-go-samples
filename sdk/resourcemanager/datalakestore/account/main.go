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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/datalake-store/armdatalakestore"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID    string
	location          = "eastus2"
	resourceGroupName = "sample-resource-group"
	accountName       = "sample2datalake2account"
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

	account, err := createDataLakeStoreAccount(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("datalake store account:", *account.ID)

	account, err = getDataLakeStoreAccount(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get datalake store account:", *account.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createDataLakeStoreAccount(ctx context.Context, cred azcore.TokenCredential) (*armdatalakestore.Account, error) {
	accountsClient, err := armdatalakestore.NewAccountsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := accountsClient.BeginCreate(
		ctx,
		resourceGroupName,
		accountName,
		armdatalakestore.CreateDataLakeStoreAccountParameters{
			Location: to.Ptr(location),
			Identity: &armdatalakestore.EncryptionIdentity{
				Type: to.Ptr("SystemAssigned"),
			},
			Properties: &armdatalakestore.CreateDataLakeStoreAccountProperties{
				EncryptionConfig: &armdatalakestore.EncryptionConfig{
					Type: to.Ptr(armdatalakestore.EncryptionConfigTypeServiceManaged),
				},
				EncryptionState: to.Ptr(armdatalakestore.EncryptionStateEnabled),
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

	return &resp.Account, nil
}

func getDataLakeStoreAccount(ctx context.Context, cred azcore.TokenCredential) (*armdatalakestore.Account, error) {
	accountsClient, err := armdatalakestore.NewAccountsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := accountsClient.Get(ctx, resourceGroupName, accountName, nil)
	if err != nil {
		return nil, err
	}

	return &resp.Account, nil
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

	_, err = pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return err
	}

	return nil
}
