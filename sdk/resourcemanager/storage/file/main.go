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
	resourceGroupName  = "sample-resources-group"
	storageAccountName = "sample2storage2account"
	shareName          = "sample-file-share"
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

	fileShare, err := createFileShare(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("file share:", *fileShare.ID)

	fileShare, err = getFileShare(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get file share:", *fileShare.ID)

	fileShare, err = updateFileShare(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("update file share:", *fileShare.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createStorageAccount(ctx context.Context, cred azcore.TokenCredential) (*armstorage.Account, error) {
	storageAccountClient := armstorage.NewAccountsClient(subscriptionID, cred, nil)

	pollerResp, err := storageAccountClient.BeginCreate(
		ctx,
		resourceGroupName,
		storageAccountName,
		armstorage.AccountCreateParameters{
			Kind: armstorage.KindStorageV2.ToPtr(),
			SKU: &armstorage.SKU{
				Name: armstorage.SKUNameStandardLRS.ToPtr(),
			},
			Location: to.StringPtr(location),
			Properties: &armstorage.AccountPropertiesCreateParameters{
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
	return &resp.Account, nil
}

func createFileShare(ctx context.Context, cred azcore.TokenCredential) (*armstorage.FileShare, error) {
	fileSharesClient := armstorage.NewFileSharesClient(subscriptionID, cred, nil)

	resp, err := fileSharesClient.Create(
		ctx,
		resourceGroupName,
		storageAccountName,
		shareName,
		armstorage.FileShare{},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.FileShare, nil
}

func getFileShare(ctx context.Context, cred azcore.TokenCredential) (*armstorage.FileShare, error) {
	fileSharesClient := armstorage.NewFileSharesClient(subscriptionID, cred, nil)

	resp, err := fileSharesClient.Get(ctx, resourceGroupName, storageAccountName, shareName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.FileShare, nil
}

func updateFileShare(ctx context.Context, cred azcore.TokenCredential) (*armstorage.FileShare, error) {
	fileSharesClient := armstorage.NewFileSharesClient(subscriptionID, cred, nil)

	resp, err := fileSharesClient.Update(
		ctx,
		resourceGroupName,
		storageAccountName,
		shareName,
		armstorage.FileShare{
			FileShareProperties: &armstorage.FileShareProperties{
				Metadata: map[string]*string{
					"sample": to.StringPtr("value"),
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.FileShare, nil
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
