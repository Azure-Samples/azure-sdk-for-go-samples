package storage

import (
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2017-06-01/storage"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

func getStorageAccountsClient() storage.AccountsClient {
	token, _ := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	storageAccountsClient := storage.NewAccountsClient(helpers.SubscriptionID())
	storageAccountsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return storageAccountsClient
}

func getFirstKey(accountName string) string {
	accountsClient := getStorageAccountsClient()
	res, err := accountsClient.ListKeys(helpers.ResourceGroupName(), accountName)
	if err != nil {
		log.Fatalf("failed to list keys: %v", err)
	}
	return *(((*res.Keys)[0]).Value)
}

// CreateStorageAccount creates a new storage account.
func CreateStorageAccount(accountName string) (<-chan storage.Account, <-chan error) {
	storageAccountsClient := getStorageAccountsClient()

	result, err := storageAccountsClient.CheckNameAvailability(
		storage.AccountCheckNameAvailabilityParameters{
			Name: to.StringPtr(accountName),
			Type: to.StringPtr("Microsoft.Storage/storageAccounts"),
		})
	if err != nil {
		log.Fatalf("%s: %v", "storage account creation failed", err)
	}
	if *result.NameAvailable != true {
		log.Fatalf("%s [%s]: %v", "storage account name not available", accountName, err)
	}

	return storageAccountsClient.Create(
		helpers.ResourceGroupName(),
		accountName,
		storage.AccountCreateParameters{
			Sku: &storage.Sku{
				Name: storage.StandardLRS},
			Location: to.StringPtr(helpers.Location()),
			AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{},
		},
		nil /* cancel <-chan struct{} */)
}

func DeleteStorageAccount(accountName string) (autorest.Response, error) {
	storageAccountsClient := getStorageAccountsClient()
	return storageAccountsClient.Delete(helpers.ResourceGroupName(), accountName)
}
