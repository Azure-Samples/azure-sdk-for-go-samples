// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/eventhub/armeventhub"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

var (
	subscriptionID        string
	location              = "westus"
	resourceGroupName     = "sample-resource-group"
	storageAccountName    = "sample1storage"
	namespacesName        = "sample1namespace"
	eventHubName          = "sample-eventhub"
	authorizationRuleName = "sample-eventhub-authorization-rule"
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

func createNamespace(ctx context.Context, cred azcore.TokenCredential) (*armeventhub.EHNamespace, error) {
	namespacesClient := armeventhub.NewNamespacesClient(subscriptionID, cred, nil)

	pollerResp, err := namespacesClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		namespacesName,
		armeventhub.EHNamespace{
			Location: to.StringPtr(location),
			Tags: map[string]*string{
				"tag1": to.StringPtr("value1"),
				"tag2": to.StringPtr("value2"),
			},
			SKU: &armeventhub.SKU{
				Name: armeventhub.SKUNameStandard.ToPtr(),
				Tier: armeventhub.SKUTierStandard.ToPtr(),
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
	return &resp.EHNamespace, nil
}

func createEventHub(ctx context.Context, cred azcore.TokenCredential, storageAccountID string) (*armeventhub.Eventhub, error) {
	eventHubsClient := armeventhub.NewEventHubsClient(subscriptionID, cred, nil)

	resp, err := eventHubsClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		namespacesName,
		eventHubName,
		armeventhub.Eventhub{
			Properties: &armeventhub.Properties{
				MessageRetentionInDays: to.Int64Ptr(4),
				PartitionCount:         to.Int64Ptr(4),
				Status:                 armeventhub.EntityStatusActive.ToPtr(),
				CaptureDescription: &armeventhub.CaptureDescription{
					Enabled:           to.BoolPtr(true),
					Encoding:          armeventhub.EncodingCaptureDescriptionAvro.ToPtr(),
					IntervalInSeconds: to.Int32Ptr(120),
					SizeLimitInBytes:  to.Int32Ptr(10485763),
					Destination: &armeventhub.Destination{
						Name: to.StringPtr("EventHubArchive.AzureBlockBlob"),
						Properties: &armeventhub.DestinationProperties{
							ArchiveNameFormat:        to.StringPtr("{Namespace}/{EventHub}/{PartitionId}/{Year}/{Month}/{Day}/{Hour}/{Minute}/{Second}"),
							BlobContainer:            to.StringPtr("container"),
							StorageAccountResourceID: to.StringPtr(storageAccountID),
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
