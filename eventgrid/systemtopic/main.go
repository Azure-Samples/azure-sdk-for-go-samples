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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventgrid/armeventgrid"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

var (
	subscriptionID     string
	location           = "westus"
	resourceGroupName  = "sample-resource-group"
	storageAccountName = "sample2storage2account"
	systemTopicName    = "sample-event-topic"
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

	systemTopic, err := createSystemTopic(ctx, conn, *storageAccount.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("system topic:", *systemTopic.ID)

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

func createSystemTopic(ctx context.Context, conn *arm.Connection, storageAccountID string) (*armeventgrid.SystemTopic, error) {
	systemTopicsClient := armeventgrid.NewSystemTopicsClient(conn, subscriptionID)

	pollerResp, err := systemTopicsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		systemTopicName,
		armeventgrid.SystemTopic{
			TrackedResource: armeventgrid.TrackedResource{
				Location: to.StringPtr(location),
			},
			Properties: &armeventgrid.SystemTopicProperties{
				Source:    to.StringPtr(storageAccountID),
				TopicType: to.StringPtr("microsoft.storage.storageaccounts"),
			},
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
	return &resp.SystemTopic, nil
}

func getSystemTopic(ctx context.Context, conn *arm.Connection, storageAccountID string) (*armeventgrid.SystemTopic, error) {
	systemTopicsClient := armeventgrid.NewSystemTopicsClient(conn, subscriptionID)

	resp, err := systemTopicsClient.Get(ctx, resourceGroupName, systemTopicName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.SystemTopic, nil
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
