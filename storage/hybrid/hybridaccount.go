package hybridstorage

import (
	"context"
	"fmt"
	"log"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/profiles/2017-03-09/storage/mgmt/storage"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

var (
	location = "local"
)

func getStorageAccountsClient() storage.AccountsClient {
	token, err := iam.GetResourceManagementTokenHybrid(helpers.ActiveDirectoryEndpoint(), helpers.TenantID(), helpers.ClientID(), helpers.ClientSecret(), helpers.ActiveDirectoryResourceID())
	if err != nil {
		log.Fatal(fmt.Sprintf("Cannot generate token. Error details: %s.", err.Error()))
	}
	storageAccountsClient := storage.NewAccountsClientWithBaseURI(helpers.ArmEndpoint(), helpers.SubscriptionID())
	storageAccountsClient.Authorizer = autorest.NewBearerAuthorizer(token)
	return storageAccountsClient
}

// CreateStorageAccount creates a new storage account.
func CreateStorageAccount(cntx context.Context, accountName string) (s storage.Account, err error) {
	storageAccountsClient := getStorageAccountsClient()
	result, err := storageAccountsClient.CheckNameAvailability(
		cntx,
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
	future, err := storageAccountsClient.Create(
		cntx,
		helpers.ResourceGroupName(),
		accountName,
		storage.AccountCreateParameters{
			Sku: &storage.Sku{
				Name: storage.StandardLRS},
			Location: to.StringPtr(location),
			AccountPropertiesCreateParameters: &storage.AccountPropertiesCreateParameters{},
		})
	if err != nil {
		return s, fmt.Errorf("cannot create storage account: %v", err)
	}
	err = future.WaitForCompletion(cntx, storageAccountsClient.Client)
	if err != nil {
		return s, fmt.Errorf("cannot get the storage account create future response: %v", err)
	}
	return future.Result(storageAccountsClient)
}
