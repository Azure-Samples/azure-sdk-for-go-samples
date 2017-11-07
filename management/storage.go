package management

import (
	"github.com/Azure/azure-sdk-for-go/profiles/preview/storage/mgmt/storage"
	"github.com/joshgav/az-go/common"
	"github.com/subosito/gotenv"
	"log"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

var (
	accountName string
)

func init() {
	gotenv.Load() // read from .env file

	accountName = common.GetEnvVarOrFail("AZURE_STORAGEACCOUNT_NAME")
}

func getStorageAccountsClient() (storage.AccountsClient, error) {
	token, err := common.GetResourceManagementToken(common.OAuthGrantTypeServicePrincipal)
	if err != nil {
		log.Fatalf("%s: %v", "failed to get auth token", err)
	}

	storageAccountsClient := storage.NewAccountsClient(subscriptionId)
	storageAccountsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return storageAccountsClient, err
}

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
		resourceGroupName,
		accountName,
		storage.AccountCreateParameters{
			Sku: &storage.Sku{
				Name: storage.StandardLRS},
			Location: to.StringPtr(location),
			AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{}},
		nil /* cancel <-chan struct{} */)
}

func DeleteStorageAccount() (autorest.Response, error) {
	storageAccountsClient, _ := getStorageAccountsClient()
	return storageAccountsClient.Delete(resourceGroupName, accountName)
}
