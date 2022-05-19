// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventhub/armeventhub"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
	"log"
	"os"
)

var (
	subscriptionID     string
	location           = "westus"
	resourceGroupName  = "sample-resource-group"
	storageAccountName = "sample1storage"
	namespacesName     = "sample1namespace"
	eventHubName       = "sample-eventhub"
	consumerGroupName  = "sample-consumer-group"
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

	namespace, err := createNamespace(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("eventhub namespace:", *namespace.ID)

	eventhub, err := createEventHub(ctx, cred, *storageAccount.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("eventhub:", *eventhub.ID)

	consumerGroup, err := createConsumerGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("consumer group:", *consumerGroup.ID)

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

func createNamespace(ctx context.Context, cred azcore.TokenCredential) (*armeventhub.EHNamespace, error) {
	namespacesClient, err := armeventhub.NewNamespacesClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := namespacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		namespacesName,
		armeventhub.EHNamespace{
			Location: to.Ptr(location),
			Tags: map[string]*string{
				"tag1": to.Ptr("value1"),
				"tag2": to.Ptr("value2"),
			},
			SKU: &armeventhub.SKU{
				Name: to.Ptr(armeventhub.SKUNameStandard),
				Tier: to.Ptr(armeventhub.SKUTierStandard),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, nil)
	if err != nil {
		return nil, err
	}
	return &resp.EHNamespace, nil
}

func createEventHub(ctx context.Context, cred azcore.TokenCredential, storageAccountID string) (*armeventhub.Eventhub, error) {
	eventHubsClient, err := armeventhub.NewEventHubsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := eventHubsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		namespacesName,
		eventHubName,
		armeventhub.Eventhub{
			Properties: &armeventhub.Properties{
				MessageRetentionInDays: to.Ptr[int64](4),
				PartitionCount:         to.Ptr[int64](4),
				Status:                 to.Ptr(armeventhub.EntityStatusActive),
				CaptureDescription: &armeventhub.CaptureDescription{
					Enabled:           to.Ptr(true),
					Encoding:          to.Ptr(armeventhub.EncodingCaptureDescriptionAvro),
					IntervalInSeconds: to.Ptr[int32](120),
					SizeLimitInBytes:  to.Ptr[int32](10485763),
					Destination: &armeventhub.Destination{
						Name: to.Ptr("EventHubArchive.AzureBlockBlob"),
						Properties: &armeventhub.DestinationProperties{
							ArchiveNameFormat:        to.Ptr("{Namespace}/{EventHub}/{PartitionId}/{Year}/{Month}/{Day}/{Hour}/{Minute}/{Second}"),
							BlobContainer:            to.Ptr("container"),
							StorageAccountResourceID: to.Ptr(storageAccountID),
						},
					},
				},
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.Eventhub, nil
}

func createConsumerGroup(ctx context.Context, cred azcore.TokenCredential) (*armeventhub.ConsumerGroup, error) {
	consumerGroupsClient, err := armeventhub.NewConsumerGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := consumerGroupsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		namespacesName,
		eventHubName,
		consumerGroupName,
		armeventhub.ConsumerGroup{
			Properties: &armeventhub.ConsumerGroupProperties{
				UserMetadata: to.Ptr("New consumer group"),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.ConsumerGroup, nil
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
