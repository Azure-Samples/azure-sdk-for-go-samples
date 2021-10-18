package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/operationalinsights/armoperationalinsights"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

var (
	subscriptionID     string
	location           = "westus"
	resourceGroupName  = "sample-resource-group"
	workspaceName      = "sample-workspace"
	storageInsightName = "sample-storage-insight"
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

	storageAccount, err := createStorageAccount(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("storage account:", *storageAccount.ID)

	keys := regenerateKeyStorageAccount(ctx, conn)
	for _, v := range keys {
		if *v.KeyName == "key1" {
			log.Println("regenerate key:", *v.KeyName, *v.Value, *v.CreationTime, *v.Permissions)
		}
	}

	workspace, err := createWorkspace(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("operational insights workspace:", *workspace.ID)

	storageInsight, err := createStorageInsight(ctx, conn, *storageAccount.ID, *keys[0].Value)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("storage insight:", *storageInsight.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
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

func createWorkspace(ctx context.Context, conn *arm.Connection) (*armoperationalinsights.Workspace, error) {
	workspacesClient := armoperationalinsights.NewWorkspacesClient(conn, subscriptionID)

	pollerResp, err := workspacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		workspaceName,
		armoperationalinsights.Workspace{
			TrackedResource: armoperationalinsights.TrackedResource{
				Location: to.StringPtr(location),
			},
			Properties: &armoperationalinsights.WorkspaceProperties{},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.Workspace, nil
}

func createStorageInsight(ctx context.Context, conn *arm.Connection, storageAccountID, storageKeyID string) (*armoperationalinsights.StorageInsight, error) {
	storageInsightConfigsClient := armoperationalinsights.NewStorageInsightConfigsClient(conn, subscriptionID)

	resp, err := storageInsightConfigsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		workspaceName,
		storageInsightName,
		armoperationalinsights.StorageInsight{
			Properties: &armoperationalinsights.StorageInsightProperties{
				StorageAccount: &armoperationalinsights.StorageAccount{
					ID:  to.StringPtr(storageAccountID),
					Key: to.StringPtr(storageKeyID),
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.StorageInsight, nil
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
