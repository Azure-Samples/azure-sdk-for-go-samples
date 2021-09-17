package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/storage/armstorage"
)

var (
	subscriptionID     string
	location           = "westus"
	resourceGroupName  = "sample-resource-group"
	storageAccountName = "sample2storage2account"
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

	availability, err := checkNameAvailability(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	if !*availability.NameAvailable {
		log.Fatal(*availability.Message)
	}

	storageAccount, err := createStorageAccount(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("storage account:", *storageAccount.ID)

	properties, err := storageAccountProperties(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println(*properties.ID)

	listByResourceGroup := listByResourceGroupStorageAccount(ctx, conn)
	for _, sa := range listByResourceGroup {
		log.Println(*sa.ID)
	}

	list := listStorageAccount(ctx, conn)
	log.Println("Storage Accounts:")
	for _, sa := range list {
		log.Println("\t" + *sa.ID)
	}

	keys := regenerateKeyStorageAccount(ctx, conn)
	for _, v := range keys {
		if *v.KeyName == "key1" {
			log.Println("regenerate key:", *v.KeyName, *v.Value, *v.CreationTime, *v.Permissions)
		}
	}

	keys2 := listKeysStorageAccount(ctx, conn)
	log.Println("list keys:")
	for i, v := range keys2 {
		log.Println("\t", i, *v.KeyName, *v.Value, *v.CreationTime, *v.Permissions)
	}

	update, err := updateStorageAccount(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	fmt.Printf("updated storage account:%s,sample tag:%s\n", *update.ID, *update.Tags["sample"])

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func storageAccountProperties(ctx context.Context, conn *arm.Connection) (*armstorage.StorageAccount, error) {
	storageAccountClient := armstorage.NewStorageAccountsClient(conn, subscriptionID)

	storageAccountResponse, err := storageAccountClient.GetProperties(
		ctx,
		resourceGroupName,
		storageAccountName,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &storageAccountResponse.StorageAccount, nil
}

func checkNameAvailability(ctx context.Context, conn *arm.Connection) (*armstorage.CheckNameAvailabilityResult, error) {
	storageAccountClient := armstorage.NewStorageAccountsClient(conn, subscriptionID)
	result, err := storageAccountClient.CheckNameAvailability(
		ctx,
		armstorage.StorageAccountCheckNameAvailabilityParameters{
			Name: to.StringPtr(storageAccountName),
			Type: to.StringPtr("Microsoft.Storage/storageAccounts"),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &result.CheckNameAvailabilityResult, nil
}

func createStorageAccount(ctx context.Context, conn *arm.Connection) (*armstorage.StorageAccount, error) {
	storageAccountClient := armstorage.NewStorageAccountsClient(conn, subscriptionID)

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
						Queue: &armstorage.EncryptionService{
							KeyType: armstorage.KeyTypeAccount.ToPtr(),
							Enabled: to.BoolPtr(true),
						},
						Table: &armstorage.EncryptionService{
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

func listByResourceGroupStorageAccount(ctx context.Context, conn *arm.Connection) []*armstorage.StorageAccount {
	storageAccountClient := armstorage.NewStorageAccountsClient(conn, subscriptionID)

	listAccounts := storageAccountClient.ListByResourceGroup(resourceGroupName, nil)

	list := make([]*armstorage.StorageAccount, 0)
	for listAccounts.NextPage(ctx) {
		pageResponse := listAccounts.PageResponse()
		list = append(list, pageResponse.StorageAccountListResult.Value...)
	}
	return list
}

func listStorageAccount(ctx context.Context, conn *arm.Connection) []*armstorage.StorageAccount {
	storageAccountClient := armstorage.NewStorageAccountsClient(conn, subscriptionID)

	listAccounts := storageAccountClient.List(nil)

	list := make([]*armstorage.StorageAccount, 0)
	for listAccounts.NextPage(ctx) {
		pageResponse := listAccounts.PageResponse()
		list = append(list, pageResponse.StorageAccountListResult.Value...)
	}

	return list
}

func listKeysStorageAccount(ctx context.Context, conn *arm.Connection) []*armstorage.StorageAccountKey {
	storageAccountClient := armstorage.NewStorageAccountsClient(conn, subscriptionID)

	listKeys, err := storageAccountClient.ListKeys(ctx, resourceGroupName, storageAccountName, nil)
	if err != nil {
		log.Fatal(err)
	}

	return listKeys.StorageAccountListKeysResult.Keys
}

func regenerateKeyStorageAccount(ctx context.Context, conn *arm.Connection) []*armstorage.StorageAccountKey {
	storageAccountClient := armstorage.NewStorageAccountsClient(conn, subscriptionID)

	regenerateKeyResp, err := storageAccountClient.RegenerateKey(
		ctx,
		resourceGroupName,
		storageAccountName,
		armstorage.StorageAccountRegenerateKeyParameters{
			KeyName: to.StringPtr("key1"),
		},
		nil)

	if err != nil {
		log.Fatal(err)
	}

	return regenerateKeyResp.StorageAccountListKeysResult.Keys
}

func updateStorageAccount(ctx context.Context, conn *arm.Connection) (*armstorage.StorageAccount, error) {
	storageAccountClient := armstorage.NewStorageAccountsClient(conn, subscriptionID)

	updateResp, err := storageAccountClient.Update(
		ctx,
		resourceGroupName,
		storageAccountName,
		armstorage.StorageAccountUpdateParameters{
			Tags: map[string]*string{
				"sample": to.StringPtr("golang"),
			},
		},
		nil)
	if err != nil {
		return nil, fmt.Errorf("update storage account err:%s", err)
	}

	return &updateResp.StorageAccount, nil
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
