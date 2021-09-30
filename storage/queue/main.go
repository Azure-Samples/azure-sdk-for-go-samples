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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
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

	queue, err := createQueue(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("file share:", *queue.ID)

	queue, err = getQueue(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get file share:", *queue.ID)

	queue, err = updateQueue(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("update file share:", *queue.ID)

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

func createQueue(ctx context.Context, conn *arm.Connection) (*armstorage.StorageQueue, error) {
	queueClient := armstorage.NewQueueClient(conn, subscriptionID)

	storageQueueResp, err := queueClient.Create(
		ctx,
		resourceGroupName,
		storageAccountName,
		queueName,
		armstorage.StorageQueue{},
		nil)
	if err != nil {
		return nil, err
	}
	return &storageQueueResp.StorageQueue, nil
}

func getQueue(ctx context.Context, conn *arm.Connection) (*armstorage.StorageQueue, error) {
	queueClient := armstorage.NewQueueClient(conn, subscriptionID)

	storageQueueResp, err := queueClient.Get(
		ctx,
		resourceGroupName,
		storageAccountName,
		queueName,
		nil)
	if err != nil {
		return nil, err
	}
	return &storageQueueResp.StorageQueue, nil
}

func updateQueue(ctx context.Context, conn *arm.Connection) (*armstorage.StorageQueue, error) {
	queueClient := armstorage.NewQueueClient(conn, subscriptionID)

	storageQueueResp, err := queueClient.Update(
		ctx,
		resourceGroupName,
		storageAccountName,
		queueName,
		armstorage.StorageQueue{
			QueueProperties: &armstorage.QueueProperties{
				Metadata: map[string]*string{
					"sample1": to.StringPtr("value1"),
					"sample2": to.StringPtr("value2"),
				},
			},
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &storageQueueResp.StorageQueue, nil
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
