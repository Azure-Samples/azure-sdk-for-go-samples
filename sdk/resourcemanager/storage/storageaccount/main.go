// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"fmt"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"log"
	"os"
)

var (
	subscriptionID     string
	location           = "westus"
	resourceGroupName  = "sample-resource-group"
	storageAccountName = "sample2storage2account"
)

var (
	resourcesClientFactory *armresources.ClientFactory
	storageClientFactory   *armstorage.ClientFactory
)

var (
	resourceGroupClient *armresources.ResourceGroupsClient
	accountsClient      *armstorage.AccountsClient
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

	storageClientFactory, err = armstorage.NewClientFactory(subscriptionID, cred, nil)
	if err != nil {
		log.Fatal(err)
	}
	accountsClient = storageClientFactory.NewAccountsClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	availability, err := checkNameAvailability(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if !*availability.NameAvailable {
		log.Fatal(*availability.Message)
	}

	storageAccount, err := createStorageAccount(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("storage account:", *storageAccount.ID)

	properties, err := storageAccountProperties(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(*properties.ID)

	listByResourceGroup, err := listByResourceGroupStorageAccount(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, sa := range listByResourceGroup {
		log.Println(*sa.ID)
	}

	list, err := listStorageAccount(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("Storage Accounts:")
	for _, sa := range list {
		log.Println("\t" + *sa.ID)
	}

	keys, err := regenerateKeyStorageAccount(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, v := range keys {
		if *v.KeyName == "key1" {
			log.Println("regenerate key:", *v.KeyName, *v.Value, *v.CreationTime, *v.Permissions)
		}
	}

	keys2, err := listKeysStorageAccount(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("list keys:")
	for i, v := range keys2 {
		log.Println("\t", i, *v.KeyName, *v.Value, *v.CreationTime, *v.Permissions)
	}

	update, err := updateStorageAccount(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Printf("updated storage account:%s,sample tag:%s\n", *update.ID, *update.Tags["sample"])

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func storageAccountProperties(ctx context.Context) (*armstorage.Account, error) {

	storageAccountResponse, err := accountsClient.GetProperties(
		ctx,
		resourceGroupName,
		storageAccountName,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &storageAccountResponse.Account, nil
}

func checkNameAvailability(ctx context.Context) (*armstorage.CheckNameAvailabilityResult, error) {

	result, err := accountsClient.CheckNameAvailability(
		ctx,
		armstorage.AccountCheckNameAvailabilityParameters{
			Name: to.Ptr(storageAccountName),
			Type: to.Ptr("Microsoft.Storage/storageAccounts"),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &result.CheckNameAvailabilityResult, nil
}

func createStorageAccount(ctx context.Context) (*armstorage.Account, error) {

	pollerResp, err := accountsClient.BeginCreate(
		ctx,
		resourceGroupName,
		storageAccountName,
		armstorage.AccountCreateParameters{
			Kind: to.Ptr(armstorage.KindStorageV2),
			SKU: &armstorage.SKU{
				Name: to.Ptr(armstorage.SKUNameStandardLRS),
			},
			Location: to.Ptr(location),
			Properties: &armstorage.AccountPropertiesCreateParameters{
				AccessTier: to.Ptr(armstorage.AccessTierCool),
				Encryption: &armstorage.Encryption{
					Services: &armstorage.EncryptionServices{
						File: &armstorage.EncryptionService{
							KeyType: to.Ptr(armstorage.KeyTypeAccount),
							Enabled: to.Ptr(true),
						},
						Blob: &armstorage.EncryptionService{
							KeyType: to.Ptr(armstorage.KeyTypeAccount),
							Enabled: to.Ptr(true),
						},
						Queue: &armstorage.EncryptionService{
							KeyType: to.Ptr(armstorage.KeyTypeAccount),
							Enabled: to.Ptr(true),
						},
						Table: &armstorage.EncryptionService{
							KeyType: to.Ptr(armstorage.KeyTypeAccount),
							Enabled: to.Ptr(true),
						},
					},
					KeySource: to.Ptr(armstorage.KeySourceMicrosoftStorage),
				},
			},
		}, nil)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Account, nil
}

func listByResourceGroupStorageAccount(ctx context.Context) ([]*armstorage.Account, error) {

	listAccounts := accountsClient.NewListByResourceGroupPager(resourceGroupName, nil)

	list := make([]*armstorage.Account, 0)
	for listAccounts.More() {
		pageResponse, err := listAccounts.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		list = append(list, pageResponse.AccountListResult.Value...)
	}
	return list, nil
}

func listStorageAccount(ctx context.Context) ([]*armstorage.Account, error) {

	listAccounts := accountsClient.NewListPager(nil)

	list := make([]*armstorage.Account, 0)
	for listAccounts.More() {
		pageResponse, err := listAccounts.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		list = append(list, pageResponse.AccountListResult.Value...)
	}

	return list, nil
}

func listKeysStorageAccount(ctx context.Context) ([]*armstorage.AccountKey, error) {

	listKeys, err := accountsClient.ListKeys(ctx, resourceGroupName, storageAccountName, nil)
	if err != nil {
		return nil, err
	}

	return listKeys.AccountListKeysResult.Keys, nil
}

func regenerateKeyStorageAccount(ctx context.Context) ([]*armstorage.AccountKey, error) {

	regenerateKeyResp, err := accountsClient.RegenerateKey(
		ctx,
		resourceGroupName,
		storageAccountName,
		armstorage.AccountRegenerateKeyParameters{
			KeyName: to.Ptr("key1"),
		},
		nil)

	if err != nil {
		return nil, err
	}

	return regenerateKeyResp.AccountListKeysResult.Keys, nil
}

func updateStorageAccount(ctx context.Context) (*armstorage.Account, error) {

	updateResp, err := accountsClient.Update(
		ctx,
		resourceGroupName,
		storageAccountName,
		armstorage.AccountUpdateParameters{
			Tags: map[string]*string{
				"sample": to.Ptr("golang"),
			},
		},
		nil)
	if err != nil {
		return nil, fmt.Errorf("update storage account err:%s", err)
	}

	return &updateResp.Account, nil
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
