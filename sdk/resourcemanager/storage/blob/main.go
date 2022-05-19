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
	resourceGroupName  = "sample-resource-group"
	storageAccountName = "sample2storage2account"
	containerName      = "blob2container"
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

	blobContainer, err := createBlobContainers(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("blob container:", *blobContainer.ID)

	blobContainer, err = getBlobContainer(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("blob container ID:", *blobContainer.ID)

	containerItems, err := listBlobContainer(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("blob container list:")
	for _, item := range containerItems {
		log.Println("\t", *item.ID)
	}

	blobServices(ctx, cred)

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

func createBlobContainers(ctx context.Context, cred azcore.TokenCredential) (*armstorage.BlobContainer, error) {
	blobContainerClient, err := armstorage.NewBlobContainersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	blobContainerResp, err := blobContainerClient.Create(
		ctx,
		resourceGroupName,
		storageAccountName,
		containerName,
		armstorage.BlobContainer{
			ContainerProperties: &armstorage.ContainerProperties{
				PublicAccess: to.Ptr(armstorage.PublicAccessNone),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &blobContainerResp.BlobContainer, nil
}

func getBlobContainer(ctx context.Context, cred azcore.TokenCredential) (blobContainer *armstorage.BlobContainer, err error) {
	blobContainerClient, err := armstorage.NewBlobContainersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	blobContainerResp, err := blobContainerClient.Get(ctx, resourceGroupName, storageAccountName, containerName, nil)
	if err != nil {
		return
	}

	blobContainer = &blobContainerResp.BlobContainer
	return
}

func listBlobContainer(ctx context.Context, cred azcore.TokenCredential) (listItems []*armstorage.ListContainerItem, err error) {
	blobContainerClient, err := armstorage.NewBlobContainersClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	containerItemsPager := blobContainerClient.NewListPager(resourceGroupName, storageAccountName, nil)

	listItems = make([]*armstorage.ListContainerItem, 0)
	for containerItemsPager.More() {
		pageResp, err := containerItemsPager.NextPage(ctx)
		if err != nil {
			return nil, err
		}
		listItems = append(listItems, pageResp.ListContainerItems.Value...)
	}
	return
}

func blobServices(ctx context.Context, cred azcore.TokenCredential) {
	blobServicesProperties, err := setBlobServices(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	if blobServicesProperties == nil {
		log.Fatal("what")
	}
	log.Println(*blobServicesProperties.ID)

	blobServicesProperties, err = getBlobServices(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(*blobServicesProperties.ID)

	listBlob, err := listBlobServices(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	for _, properties := range listBlob {
		log.Println(*properties.ID)
	}
}

func setBlobServices(ctx context.Context, cred azcore.TokenCredential) (*armstorage.BlobServiceProperties, error) {
	blobServicesClient, err := armstorage.NewBlobServicesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	blobServicesPropertiesResp, err := blobServicesClient.SetServiceProperties(
		ctx,
		resourceGroupName,
		storageAccountName,
		armstorage.BlobServiceProperties{
			BlobServiceProperties: &armstorage.BlobServicePropertiesProperties{},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &blobServicesPropertiesResp.BlobServiceProperties, nil
}

func getBlobServices(ctx context.Context, cred azcore.TokenCredential) (*armstorage.BlobServiceProperties, error) {
	blobServicesClient, err := armstorage.NewBlobServicesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	blobServicesResp, err := blobServicesClient.GetServiceProperties(ctx, resourceGroupName, storageAccountName, nil)
	if err != nil {
		return nil, err
	}

	return &blobServicesResp.BlobServiceProperties, nil
}

func listBlobServices(ctx context.Context, cred azcore.TokenCredential) ([]*armstorage.BlobServiceProperties, error) {
	blobServicesClient, err := armstorage.NewBlobServicesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	blobServicesResp := blobServicesClient.NewListPager(resourceGroupName, storageAccountName, nil)
	resp, err := blobServicesResp.NextPage(ctx)
	if err != nil {
		return nil, err
	}
	return resp.Value, nil
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
