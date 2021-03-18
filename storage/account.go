// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2019-06-01/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

var (
	// globals used by tests
	testAccountName      string
	testAccountGroupName string
)

func getStorageAccountsClient() storage.AccountsClient {
	storageAccountsClient := storage.NewAccountsClient(config.SubscriptionID())
	auth, _ := iam.GetResourceManagementAuthorizer()
	storageAccountsClient.Authorizer = auth
	storageAccountsClient.AddToUserAgent(config.UserAgent())
	return storageAccountsClient
}

func getUsageClient() storage.UsagesClient {
	usageClient := storage.NewUsagesClient(config.SubscriptionID())
	auth, _ := iam.GetResourceManagementAuthorizer()
	usageClient.Authorizer = auth
	usageClient.AddToUserAgent(config.UserAgent())
	return usageClient
}

func getAccountPrimaryKey(ctx context.Context, accountName, accountGroupName string) string {
	response, err := GetAccountKeys(ctx, accountName, accountGroupName)
	if err != nil {
		log.Fatalf("failed to list keys: %v", err)
	}
	return *(((*response.Keys)[0]).Value)
}

// CreateStorageAccount starts creation of a new storage account and waits for
// the account to be created.
func CreateStorageAccount(ctx context.Context, accountName, accountGroupName string) (storage.Account, error) {
	var s storage.Account
	storageAccountsClient := getStorageAccountsClient()

	result, err := storageAccountsClient.CheckNameAvailability(
		ctx,
		storage.AccountCheckNameAvailabilityParameters{
			Name: to.StringPtr(accountName),
			Type: to.StringPtr("Microsoft.Storage/storageAccounts"),
		})
	if err != nil {
		return s, fmt.Errorf("storage account check-name-availability failed: %+v", err)
	}

	if !*result.NameAvailable {
		return s, fmt.Errorf(
			"storage account name [%s] not available: %v\nserver message: %v",
			accountName, err, *result.Message)
	}

	future, err := storageAccountsClient.Create(
		ctx,
		accountGroupName,
		accountName,
		storage.AccountCreateParameters{
			Sku: &storage.Sku{
				Name: storage.StandardLRS},
			Kind:                              storage.Storage,
			Location:                          to.StringPtr(config.DefaultLocation()),
			AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{},
		})

	if err != nil {
		return s, fmt.Errorf("failed to start creating storage account: %+v", err)
	}

	err = future.WaitForCompletionRef(ctx, storageAccountsClient.Client)
	if err != nil {
		return s, fmt.Errorf("failed to finish creating storage account: %+v", err)
	}

	return future.Result(storageAccountsClient)
}

// GetStorageAccount gets details on the specified storage account
func GetStorageAccount(ctx context.Context, accountName, accountGroupName string) (storage.Account, error) {
	storageAccountsClient := getStorageAccountsClient()
	return storageAccountsClient.GetProperties(ctx, accountGroupName, accountName, storage.AccountExpandBlobRestoreStatus)
}

// DeleteStorageAccount deletes an existing storate account
func DeleteStorageAccount(ctx context.Context, accountName, accountGroupName string) (autorest.Response, error) {
	storageAccountsClient := getStorageAccountsClient()
	return storageAccountsClient.Delete(ctx, accountGroupName, accountName)
}

// CheckAccountNameAvailability checks if the storage account name is available.
// Storage account names must be unique across Azure and meet other requirements.
func CheckAccountNameAvailability(ctx context.Context, accountName string) (storage.CheckNameAvailabilityResult, error) {
	storageAccountsClient := getStorageAccountsClient()
	result, err := storageAccountsClient.CheckNameAvailability(
		ctx,
		storage.AccountCheckNameAvailabilityParameters{
			Name: to.StringPtr(accountName),
			Type: to.StringPtr("Microsoft.Storage/storageAccounts"),
		})
	return result, err
}

// ListAccountsByResourceGroup lists storage accounts by resource group.
func ListAccountsByResourceGroup(ctx context.Context, groupName string) (storage.AccountListResult, error) {
	storageAccountsClient := getStorageAccountsClient()
	return storageAccountsClient.ListByResourceGroup(ctx, groupName)
}

// ListAccountsBySubscription lists storage accounts by subscription.
func ListAccountsBySubscription(ctx context.Context) (storage.AccountListResultIterator, error) {
	storageAccountsClient := getStorageAccountsClient()
	return storageAccountsClient.ListComplete(ctx)
}

// GetAccountKeys gets the storage account keys
func GetAccountKeys(ctx context.Context, accountName, accountGroupName string) (storage.AccountListKeysResult, error) {
	accountsClient := getStorageAccountsClient()
	return accountsClient.ListKeys(ctx, accountGroupName, accountName, storage.Kerb)
}

// RegenerateAccountKey regenerates the selected storage account key. `key` can be 0 or 1.
func RegenerateAccountKey(ctx context.Context, accountName, accountGroupName string, key int) (storage.AccountListKeysResult, error) {
	var list storage.AccountListKeysResult
	oldKeys, err := GetAccountKeys(ctx, accountName, accountGroupName)
	if err != nil {
		return list, err
	}
	accountsClient := getStorageAccountsClient()
	return accountsClient.RegenerateKey(
		ctx,
		accountGroupName,
		accountName,
		storage.AccountRegenerateKeyParameters{
			KeyName: (*oldKeys.Keys)[key].KeyName,
		})
}

// UpdateAccount updates a storage account by adding tags
func UpdateAccount(ctx context.Context, accountName, accountGroupName string) (storage.Account, error) {
	accountsClient := getStorageAccountsClient()
	return accountsClient.Update(
		ctx,
		accountGroupName,
		accountName,
		storage.AccountUpdateParameters{
			Tags: map[string]*string{
				"who rocks": to.StringPtr("golang"),
				"where":     to.StringPtr("on azure")},
		})
}

// ListUsage gets the usage count and limits for the resources in the subscription based on location
func ListUsage(ctx context.Context, location string) (storage.UsageListResult, error) {
	usageClient := getUsageClient()
	return usageClient.ListByLocation(ctx, location)
}
