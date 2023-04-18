// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
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

var (
	resourcesClientFactory *armresources.ClientFactory
	storageClientFactory   *armstorage.ClientFactory
)

var (
	resourceGroupClient  *armresources.ResourceGroupsClient
	accountsClient       *armstorage.AccountsClient
	blobContainersClient *armstorage.BlobContainersClient
	blobServicesClient   *armstorage.BlobServicesClient
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
	blobContainersClient = storageClientFactory.NewBlobContainersClient()
	blobServicesClient = storageClientFactory.NewBlobServicesClient()

	resourceGroup, err := createResourceGroup(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	storageAccount, err := createStorageAccount(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("storage account:", *storageAccount.ID)

	blobContainer, err := createBlobContainers(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("blob container:", *blobContainer.ID)

	blobContainer, err = getBlobContainer(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("blob container ID:", *blobContainer.ID)

	containerItems, err := listBlobContainer(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("blob container list:")
	for _, item := range containerItems {
		log.Println("\t", *item.ID)
	}

	blobServices(ctx)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
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

func createBlobContainers(ctx context.Context) (*armstorage.BlobContainer, error) {

	blobContainerResp, err := blobContainersClient.Create(
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

func getBlobContainer(ctx context.Context) (blobContainer *armstorage.BlobContainer, err error) {

	blobContainerResp, err := blobContainersClient.Get(ctx, resourceGroupName, storageAccountName, containerName, nil)
	if err != nil {
		return
	}

	blobContainer = &blobContainerResp.BlobContainer
	return
}

func listBlobContainer(ctx context.Context) (listItems []*armstorage.ListContainerItem, err error) {

	containerItemsPager := blobContainersClient.NewListPager(resourceGroupName, storageAccountName, nil)

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

func blobServices(ctx context.Context) {
	blobServicesProperties, err := setBlobServices(ctx)
	if err != nil {
		log.Fatal(err)
	}
	if blobServicesProperties == nil {
		log.Fatal("what")
	}
	log.Println(*blobServicesProperties.ID)

	blobServicesProperties, err = getBlobServices(ctx)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(*blobServicesProperties.ID)

	listBlob, err := listBlobServices(ctx)
	if err != nil {
		log.Fatal(err)
	}
	for _, properties := range listBlob {
		log.Println(*properties.ID)
	}
}

func setBlobServices(ctx context.Context) (*armstorage.BlobServiceProperties, error) {

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

func getBlobServices(ctx context.Context) (*armstorage.BlobServiceProperties, error) {

	blobServicesResp, err := blobServicesClient.GetServiceProperties(ctx, resourceGroupName, storageAccountName, nil)
	if err != nil {
		return nil, err
	}

	return &blobServicesResp.BlobServiceProperties, nil
}

func listBlobServices(ctx context.Context) ([]*armstorage.BlobServiceProperties, error) {

	blobServicesResp := blobServicesClient.NewListPager(resourceGroupName, storageAccountName, nil)
	resp, err := blobServicesResp.NextPage(ctx)
	if err != nil {
		return nil, err
	}
	return resp.Value, nil
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
