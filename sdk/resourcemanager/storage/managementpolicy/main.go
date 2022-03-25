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
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/storage/armstorage"
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

	managementPolicy, err := createManagementPolicy(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("management policy:", *managementPolicy.ID)

	managementPolicy, err = getManagementPolicy(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get management policy:", *managementPolicy.ID)

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

func createManagementPolicy(ctx context.Context, cred azcore.TokenCredential) (*armstorage.ManagementPolicy, error) {
	managementPoliciesClient := armstorage.NewManagementPoliciesClient(subscriptionID, cred, nil)

	resp, err := managementPoliciesClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		storageAccountName,
		armstorage.ManagementPolicyNameDefault,
		armstorage.ManagementPolicy{
			Properties: &armstorage.ManagementPolicyProperties{
				Policy: &armstorage.ManagementPolicySchema{
					Rules: []*armstorage.ManagementPolicyRule{
						{
							Enabled: to.BoolPtr(true),
							Name:    to.StringPtr("sampletest"),
							Type:    armstorage.RuleTypeLifecycle.ToPtr(),
							Definition: &armstorage.ManagementPolicyDefinition{
								Actions: &armstorage.ManagementPolicyAction{
									BaseBlob: &armstorage.ManagementPolicyBaseBlob{
										TierToCool: &armstorage.DateAfterModification{
											DaysAfterModificationGreaterThan: to.Float32Ptr(30),
										},
										TierToArchive: &armstorage.DateAfterModification{
											DaysAfterModificationGreaterThan: to.Float32Ptr(90),
										},
										Delete: &armstorage.DateAfterModification{
											DaysAfterModificationGreaterThan: to.Float32Ptr(1000),
										},
									},
									Snapshot: &armstorage.ManagementPolicySnapShot{
										Delete: &armstorage.DateAfterCreation{
											DaysAfterCreationGreaterThan: to.Float32Ptr(30),
										},
									},
								},
								Filters: &armstorage.ManagementPolicyFilter{
									BlobTypes: []*string{
										to.StringPtr("blockBlob"),
									},
									PrefixMatch: []*string{
										to.StringPtr("sampletestcontainer"),
									},
								},
							},
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
	return &resp.ManagementPolicy, nil
}

func getManagementPolicy(ctx context.Context, cred azcore.TokenCredential) (*armstorage.ManagementPolicy, error) {
	managementPoliciesClient := armstorage.NewManagementPoliciesClient(subscriptionID, cred, nil)

	resp, err := managementPoliciesClient.Get(ctx, resourceGroupName, storageAccountName, armstorage.ManagementPolicyNameDefault, nil)
	if err != nil {
		return nil, err
	}
	return &resp.ManagementPolicy, nil
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
