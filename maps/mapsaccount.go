package maps

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/Azure/azure-sdk-for-go/services/preview/maps/mgmt/2020-02-01-preview/maps"
	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/util"
	"github.com/marstr/randname"
)

func getAccountsClient() maps.AccountsClient {
	accountsClient := maps.NewAccountsClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	accountsClient.Authorizer = a
	accountsClient.AddToUserAgent(config.UserAgent())
	return accountsClient
}

func CreateMapsAccount(ctx context.Context, accountName string) (maps.Account, error) {
	accountClient := getAccountsClient()

	return accountClient.CreateOrUpdate(
		ctx,
		config.GroupName(),
		accountName,
		maps.AccountCreateParameters{
			Location: to.StringPtr(config.Location()),
			Sku: &maps.Sku{
				Name: to.StringPtr("S1"),
				Tier: to.StringPtr("Standard"),
			},
		},
	)
}

func DeleteMapsAccount(ctx context.Context, vaultName string) (autorest.Response, error) {
	accountsClient := getAccountsClient()
	return accountsClient.Delete(ctx, config.GroupName(), vaultName)
}

var (
	accountName = randname.GenerateWithPrefix("maps-sample-go-", 5)
)

func CreateResourceGroupWithMapAccount() maps.Account {
	var groupName = config.GenerateGroupName("Maps")
	config.SetGroupName(groupName)

	ctx := context.Background()

	_, err := resources.CreateGroup(ctx, config.GroupName())
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("resource group created")

	account, err := CreateMapsAccount(ctx, accountName)
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("account created")

	return account
}
