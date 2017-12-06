package storage

import (
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/management"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2017-06-01/storage"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

func getStorageAccountsClient() storage.AccountsClient {
	storageAccountsClient := storage.NewAccountsClient(management.GetSubID())
	storageAccountsClient.Authorizer = management.GetToken()
	return storageAccountsClient
}

func loadKey(accountName string) string {
	storageAccClient := getStorageAccountsClient()
	res, err := storageAccClient.ListKeys(management.GetResourceGroup(), accountName)
	if err != nil {
		log.Fatalf("failed to list keys: %#v", err)
	}
	return *(((*res.Keys)[0]).Value)
}

// CreateStorageAccount creates a new storage account.
func CreateStorageAccount(accountName string) (<-chan storage.Account, <-chan error) {
	storageAccClient := getStorageAccountsClient()

	result, err := storageAccClient.CheckNameAvailability(
		storage.AccountCheckNameAvailabilityParameters{
			Name: to.StringPtr(accountName),
			Type: to.StringPtr("Microsoft.Storage/storageAccounts"),
		})
	if err != nil {
		log.Fatalf("%s: %v", "storage account creation failed", err)
	}
	if *result.NameAvailable != true {
		log.Fatalf("%s: %v", "storage account name not available", err)
	}

	return storageAccClient.Create(
		management.GetResourceGroup(),
		accountName,
		storage.AccountCreateParameters{
			Sku: &storage.Sku{
				Name: storage.StandardLRS},
			Location: to.StringPtr(management.GetLocation()),
			AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{}},
		nil /* cancel <-chan struct{} */)
}

func DeleteStorageAccount(accountName string) (autorest.Response, error) {
	storageAccountsClient := getStorageAccountsClient()
	return storageAccountsClient.Delete(management.GetResourceGroup(), accountName)
}
