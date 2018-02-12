// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2017-06-01/storage"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

func getStorageAccountsClient() storage.AccountsClient {
	token, _ := iam.GetResourceManagementToken(iam.AuthGrantType())
	storageAccountsClient := storage.NewAccountsClient(helpers.SubscriptionID())
	storageAccountsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	storageAccountsClient.AddToUserAgent(helpers.UserAgent())
	return storageAccountsClient
}

func getFirstKey(ctx context.Context, accountName string) string {
	accountsClient := getStorageAccountsClient()
	res, err := accountsClient.ListKeys(ctx, helpers.ResourceGroupName(), accountName)
	if err != nil {
		log.Fatalf("failed to list keys: %v", err)
	}
	return *(((*res.Keys)[0]).Value)
}

// CreateStorageAccount creates a new storage account.
func CreateStorageAccount(ctx context.Context, accountName string) (s storage.Account, err error) {
	storageAccountsClient := getStorageAccountsClient()

	result, err := storageAccountsClient.CheckNameAvailability(
		ctx,
		storage.AccountCheckNameAvailabilityParameters{
			Name: to.StringPtr(accountName),
			Type: to.StringPtr("Microsoft.Storage/storageAccounts"),
		})
	if err != nil {
		log.Fatalf("%s: %v", "storage account creation failed", err)
	}
	if *result.NameAvailable != true {
		log.Fatalf("%s [%s]: %v: %v", "storage account name not available", accountName, err, *result.Message)
	}

	future, err := storageAccountsClient.Create(
		ctx,
		helpers.ResourceGroupName(),
		accountName,
		storage.AccountCreateParameters{
			Sku: &storage.Sku{
				Name: storage.StandardLRS},
			Location: to.StringPtr(helpers.Location()),
			AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{},
		})

	if err != nil {
		return s, fmt.Errorf("cannot create storage account: %v", err)
	}

	err = future.WaitForCompletion(ctx, storageAccountsClient.Client)
	if err != nil {
		return s, fmt.Errorf("cannot get the storage account create future response: %v", err)
	}

	return future.Result(storageAccountsClient)
}

// GetStorageAccount gets details on the specified storage account
func GetStorageAccount(ctx context.Context, accountName string) (storage.Account, error) {
	storageAccountsClient := getStorageAccountsClient()
	return storageAccountsClient.GetProperties(ctx, helpers.ResourceGroupName(), accountName)
}

// DeleteStorageAccount deletes an existing storate account
func DeleteStorageAccount(ctx context.Context, accountName string) (autorest.Response, error) {
	storageAccountsClient := getStorageAccountsClient()
	return storageAccountsClient.Delete(ctx, helpers.ResourceGroupName(), accountName)
}
