package storage

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/storage/mgmt/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/azure"
	"github.com/Azure/go-autorest/autorest/to"
)

const (
	errorPrefix = "Cannot create storage account, reason: %v"
)

func getStorageAccountsClient(activeDirectoryEndpoint, tokenAudience string) storage.AccountsClient {
	token, err := iam.GetResourceManagementTokenHybrid(activeDirectoryEndpoint, tokenAudience)
	if err != nil {
		log.Fatal(fmt.Sprintf(errorPrefix, fmt.Sprintf("Cannot generate token. Error details: %v.", err)))
	}
	storageAccountsClient := storage.NewAccountsClientWithBaseURI(helpers.ArmEndpoint(), helpers.SubscriptionID())
	storageAccountsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	storageAccountsClient.AddToUserAgent(helpers.UserAgent())
	return storageAccountsClient
}

// CreateStorageAccount creates a new storage account.
func CreateStorageAccount(cntx context.Context, accountName string) (s storage.Account, err error) {
	environment, _ := azure.EnvironmentFromURL(helpers.ArmEndpoint())
	storageAccountsClient := getStorageAccountsClient(environment.ActiveDirectoryEndpoint, environment.TokenAudience)
	result, err := storageAccountsClient.CheckNameAvailability(
		cntx,
		storage.AccountCheckNameAvailabilityParameters{
			Name: to.StringPtr(accountName),
			Type: to.StringPtr("Microsoft.Storage/storageAccounts"),
		})
	if err != nil {
		return s, fmt.Errorf(errorPrefix, err)
	}
	if *result.NameAvailable != true {
		return s, fmt.Errorf(errorPrefix, fmt.Sprintf("storage account name [%v] not available", accountName))
	}
	future, err := storageAccountsClient.Create(
		cntx,
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
		return s, fmt.Errorf(fmt.Sprintf(errorPrefix, err))
	}
	err = future.WaitForCompletion(cntx, storageAccountsClient.Client)
	if err != nil {
		return s, fmt.Errorf(fmt.Sprintf(errorPrefix, fmt.Sprintf("cannot get the storage account create future response: %v", err)))
	}
	return future.Result(storageAccountsClient)
}
