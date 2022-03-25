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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/monitor/armmonitor"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
)

var (
	subscriptionID     string
	location           = "westus"
	resourceGroupName  = "sample-resource-group"
	storageAccountName = "sample2storage2account"
	logProfileName     = "sample-log-profile"
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

	logProfile, err := createLogProfile(ctx, cred, *storageAccount.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("log profile:", *logProfile.ID)

	logProfile, err = getLogProfile(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get log profile:", *logProfile.ID)

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
				EnableHTTPSTrafficOnly: to.BoolPtr(true),
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

func createLogProfile(ctx context.Context, cred azcore.TokenCredential, storageAccountID string) (*armmonitor.LogProfileResource, error) {
	logProfilesClient := armmonitor.NewLogProfilesClient(subscriptionID, cred, nil)

	resp, err := logProfilesClient.CreateOrUpdate(
		ctx,
		logProfileName,
		armmonitor.LogProfileResource{
			Location: to.StringPtr(location),
			Properties: &armmonitor.LogProfileProperties{
				Categories: []*string{
					to.StringPtr("Write"),
					to.StringPtr("Delete"),
					to.StringPtr("Action"),
				},
				Locations: []*string{
					to.StringPtr("global"),
				},
				RetentionPolicy: &armmonitor.RetentionPolicy{
					Enabled: to.BoolPtr(true),
					Days:    to.Int32Ptr(3),
				},
				StorageAccountID: to.StringPtr(storageAccountID),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.LogProfileResource, nil
}

func getLogProfile(ctx context.Context, cred azcore.TokenCredential) (*armmonitor.LogProfileResource, error) {
	logProfilesClient := armmonitor.NewLogProfilesClient(subscriptionID, cred, nil)

	resp, err := logProfilesClient.Get(ctx, logProfileName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.LogProfileResource, nil
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
