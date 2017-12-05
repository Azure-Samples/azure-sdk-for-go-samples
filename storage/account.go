package storage

import (
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/storage/mgmt/2017-06-01/storage"
	"github.com/subosito/gotenv"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

func getStorageAccountsClient() storage.AccountsClient {
	storageAccountsClient := storage.NewAccountsClient(subscriptionId)
	storageAccountsClient.Authorizer = token
	return storageAccountsClient
}

func CreateStorageAccount(accountName string) (<-chan storage.Account, <-chan error) {
	storageAccountsClient := getStorageAccountsClient()
var (
	accountName string
	accountKey  string
)

func init() {
	gotenv.Load() // read from .env file
	accountName = helpers.GetEnvVarOrFail("AZURE_STORAGE_ACCOUNTNAME")
}

func getStorageAccountsClient() (storage.AccountsClient, error) {
	token, err := iam.GetResourceManagementToken(iam.OAuthGrantTypeServicePrincipal)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get auth token", err)
	}

	storageAccountsClient := storage.NewAccountsClient(helpers.SubscriptionID)
	storageAccountsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return storageAccountsClient, err
}

func loadKey() {
	storageAccountsClient, _ := getStorageAccountsClient()
	res, err := storageAccountsClient.ListKeys(helpers.ResourceGroupName, accountName)
	if err != nil {
		log.Fatalf("failed to list keys: %#v", err)
	}
	accountKey = *(((*res.Keys)[0]).Value)
}

// CreateStorageAccount creates a new storage account.
func CreateStorageAccount() (<-chan storage.Account, <-chan error) {
	storageAccountsClient, _ := getStorageAccountsClient()

	result, err := storageAccountsClient.CheckNameAvailability(
		storage.AccountCheckNameAvailabilityParameters{
			Name: to.StringPtr(accountName),
			Type: to.StringPtr("Microsoft.Storage/storageAccounts")})
	if err != nil {
		log.Fatalf("%s: %v", "storage account creation failed", err)
	}
	if *result.NameAvailable != true {
		log.Fatalf("%s: %v", "storage account name not available", err)
	}

	return storageAccountsClient.Create(
		helpers.ResourceGroupName,
		accountName,
		storage.AccountCreateParameters{
			Sku: &storage.Sku{
				Name: storage.StandardLRS},
			Location: to.StringPtr(helpers.Location),
			AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{}},
		nil /* cancel <-chan struct{} */)
}

func DeleteStorageAccount(accountName string) (autorest.Response, error) {
	storageAccountsClient := getStorageAccountsClient()
	return storageAccountsClient.Delete(resourceGroupName, accountName)
}
