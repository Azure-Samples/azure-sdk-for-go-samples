// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
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
	resourceGroupName  = "sample-resources-group"
	storageAccountName = "sample2storage2account"
	queueName          = "sample-storage-queue"
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

	storageAccount, err := createStorageAccount(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("storage account:", *storageAccount.ID)

	queue, err := createQueue(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("file share:", *queue.ID)

	queue, err = getQueue(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get file share:", *queue.ID)

	queue, err = updateQueue(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("update file share:", *queue.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createStorageAccount(ctx context.Context, cred azcore.TokenCredential) (*armstorage.Account, error) {
	storageAccountClient, err := armstorage.NewAccountsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := storageAccountClient.BeginCreate(
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

func createQueue(ctx context.Context, cred azcore.TokenCredential) (*armstorage.Queue, error) {
	queueClient, err := armstorage.NewQueueClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	storageQueueResp, err := queueClient.Create(
		ctx,
		resourceGroupName,
		storageAccountName,
		queueName,
		armstorage.Queue{},
		nil)
	if err != nil {
		return nil, err
	}
	return &storageQueueResp.Queue, nil
}

func getQueue(ctx context.Context, cred azcore.TokenCredential) (*armstorage.Queue, error) {
	queueClient, err := armstorage.NewQueueClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	storageQueueResp, err := queueClient.Get(
		ctx,
		resourceGroupName,
		storageAccountName,
		queueName,
		nil)
	if err != nil {
		return nil, err
	}
	return &storageQueueResp.Queue, nil
}

func updateQueue(ctx context.Context, cred azcore.TokenCredential) (*armstorage.Queue, error) {
	queueClient, err := armstorage.NewQueueClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	storageQueueResp, err := queueClient.Update(
		ctx,
		resourceGroupName,
		storageAccountName,
		queueName,
		armstorage.Queue{
			QueueProperties: &armstorage.QueueProperties{
				Metadata: map[string]*string{
					"sample1": to.Ptr("value1"),
					"sample2": to.Ptr("value2"),
				},
			},
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &storageQueueResp.Queue, nil
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
