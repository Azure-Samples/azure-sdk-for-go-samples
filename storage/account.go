// Copyright (c) Microsoft and contributors.  All rights reserved.
//
// This source code is licensed under the MIT license found in the
// LICENSE file in the root directory of this source tree.

package storage

import (
	"context"
	"fmt"
	"log"
	"strings"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2017-06-01/storage"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

func getAccountName() string {
	accountName := "azuresamplesgo" + helpers.GetRandomLetterSequence(10)
	return strings.ToLower(accountName)
}

func getStorageAccountsClient() storage.AccountsClient {
	storageAccountsClient := storage.NewAccountsClient(helpers.SubscriptionID())
	auth, _ := iam.GetResourceManagementAuthorizer(iam.AuthGrantType())
	storageAccountsClient.Authorizer = auth
	storageAccountsClient.AddToUserAgent(helpers.UserAgent())
	return storageAccountsClient
}

func getUsageClient() storage.UsageClient {
	usageClient := storage.NewUsageClient(helpers.SubscriptionID())
	auth, _ := iam.GetResourceManagementAuthorizer(iam.AuthGrantType())
	usageClient.Authorizer = auth
	usageClient.AddToUserAgent(helpers.UserAgent())
	return usageClient
}

func getFirstKey(ctx context.Context, accountName string) string {
	res, err := GetAccountKeys(ctx, accountName)
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
			Kind:     storage.Storage,
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

// CheckAccountAvailability checkts if the storage account name is available.
// Storage aqccount names should be unique across all of Azure
func CheckAccountAvailability(ctx context.Context, accountName string) (bool, error) {
	storageAccountsClient := getStorageAccountsClient()
	result, err := storageAccountsClient.CheckNameAvailability(ctx, storage.AccountCheckNameAvailabilityParameters{
		Name: to.StringPtr(accountName),
		Type: to.StringPtr("Microsoft.Storage/storageAccounts"),
	})
	return *result.NameAvailable, err
}

// ListAccountsByResourceGroup lists storage accounts by resource group
func ListAccountsByResourceGroup(ctx context.Context) (storage.AccountListResult, error) {
	storageAccountsClient := getStorageAccountsClient()
	return storageAccountsClient.ListByResourceGroup(ctx, helpers.ResourceGroupName())
}

// ListAccountsBySubscription lists storage accounts by subscription
func ListAccountsBySubscription(ctx context.Context) (storage.AccountListResult, error) {
	storageAccountsClient := getStorageAccountsClient()
	return storageAccountsClient.List(ctx)
}

// GetAccountKeys gets the storage account keys
func GetAccountKeys(ctx context.Context, accountName string) (storage.AccountListKeysResult, error) {
	accountsClient := getStorageAccountsClient()
	return accountsClient.ListKeys(ctx, helpers.ResourceGroupName(), accountName)
}

// RegenerateAccountKey regenerates the selected storage account key
func RegenerateAccountKey(ctx context.Context, accountName string, key int) (list storage.AccountListKeysResult, err error) {
	oldKeys, err := GetAccountKeys(ctx, accountName)
	if err != nil {
		return list, err
	}
	accountsClient := getStorageAccountsClient()
	return accountsClient.RegenerateKey(
		ctx,
		helpers.ResourceGroupName(),
		accountName,
		storage.AccountRegenerateKeyParameters{
			KeyName: (*oldKeys.Keys)[key].KeyName,
		})
}

// UpdateAccount updates a storage account by adding tags
func UpdateAccount(ctx context.Context, accountName string) (storage.Account, error) {
	accountsClient := getStorageAccountsClient()
	return accountsClient.Update(
		ctx,
		helpers.ResourceGroupName(),
		accountName,
		storage.AccountUpdateParameters{
			Tags: map[string]*string{
				"who rocks": to.StringPtr("golang"),
				"where":     to.StringPtr("on azure")},
		})
}

// ListUsage gets the usage count and limits for the resources in the subscription
func ListUsage(ctx context.Context) (storage.UsageListResult, error) {
	usageClient := getUsageClient()
	return usageClient.List(ctx)
}
