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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
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

	containerItems := listBlobContainer(ctx, cred)
	log.Println("blob container list:")
	for _, item := range containerItems {
		log.Println("\t", *item.ID)
	}

	blobServices(ctx, cred)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createStorageAccount(ctx context.Context, cred azcore.TokenCredential) (*armstorage.StorageAccount, error) {
	storageAccountClient := armstorage.NewStorageAccountsClient(subscriptionID, cred, nil)

	pollerResp, err := storageAccountClient.BeginCreate(
		ctx,
		resourceGroupName,
		storageAccountName,
		armstorage.StorageAccountCreateParameters{
			Kind: armstorage.KindStorageV2.ToPtr(),
			SKU: &armstorage.SKU{
				Name: armstorage.SKUNameStandardLRS.ToPtr(),
			},
			Location: to.StringPtr(location),
			Properties: &armstorage.StorageAccountPropertiesCreateParameters{
				AccessTier: armstorage.AccessTierCool.ToPtr(),
				Encryption: &armstorage.Encryption{
					Services: &armstorage.EncryptionServices{
						File: &armstorage.EncryptionService{
							KeyType: armstorage.KeyTypeAccount.ToPtr(),
							Enabled: to.BoolPtr(true),
						},
						Blob: &armstorage.EncryptionService{
							KeyType: armstorage.KeyTypeAccount.ToPtr(),
							Enabled: to.BoolPtr(true),
						},
					},
					KeySource: armstorage.KeySourceMicrosoftStorage.ToPtr(),
				},
			},
		}, nil)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.StorageAccount, nil
}

func createBlobContainers(ctx context.Context, cred azcore.TokenCredential) (*armstorage.BlobContainer, error) {
	blobContainerClient := armstorage.NewBlobContainersClient(subscriptionID, cred, nil)

	blobContainerResp, err := blobContainerClient.Create(
		ctx,
		resourceGroupName,
		storageAccountName,
		containerName,
		armstorage.BlobContainer{
			ContainerProperties: &armstorage.ContainerProperties{
				PublicAccess: armstorage.PublicAccessNone.ToPtr(),
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
	blobContainerClient := armstorage.NewBlobContainersClient(subscriptionID, cred, nil)

	blobContainerResp, err := blobContainerClient.Get(ctx, resourceGroupName, storageAccountName, containerName, nil)
	if err != nil {
		return
	}

	blobContainer = &blobContainerResp.BlobContainer
	return
}

func listBlobContainer(ctx context.Context, cred azcore.TokenCredential) (listItems []*armstorage.ListContainerItem) {
	blobContainerClient := armstorage.NewBlobContainersClient(subscriptionID, cred, nil)

	containerItemsPager := blobContainerClient.List(resourceGroupName, storageAccountName, nil)

	listItems = make([]*armstorage.ListContainerItem, 0)
	for containerItemsPager.NextPage(ctx) {
		pageResp := containerItemsPager.PageResponse()
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
	blobServicesClient := armstorage.NewBlobServicesClient(subscriptionID, cred, nil)

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
	blobServicesClient := armstorage.NewBlobServicesClient(subscriptionID, cred, nil)

	blobServicesResp, err := blobServicesClient.GetServiceProperties(ctx, resourceGroupName, storageAccountName, nil)
	if err != nil {
		return nil, err
	}

	return &blobServicesResp.BlobServiceProperties, nil
}

func listBlobServices(ctx context.Context, cred azcore.TokenCredential) ([]*armstorage.BlobServiceProperties, error) {
	blobServicesClient := armstorage.NewBlobServicesClient(subscriptionID, cred, nil)

	blobServicesResp, err := blobServicesClient.List(ctx, resourceGroupName, storageAccountName, nil)
	if err != nil {
		return nil, err
	}

	return blobServicesResp.BlobServiceItems.Value, nil
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
