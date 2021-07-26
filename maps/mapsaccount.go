package maps

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/resources"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/services/maps/mgmt/2021-02-01/maps"
	"github.com/Azure/azure-sdk-for-go/services/preview/maps/2.0/creator"
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
		maps.Account{
			Location: to.StringPtr(config.Location()),
			Sku: &maps.Sku{
				Name: maps.NameG2,
				Tier: to.StringPtr("Standard"),
			},
		},
	)
}

func DeleteMapsAccount(ctx context.Context, accountName string) (autorest.Response, error) {
	accountsClient := getAccountsClient()
	return accountsClient.Delete(ctx, config.GroupName(), accountName)
}

var (
	accountName = randname.GenerateWithPrefix("maps-sample-go-", 5)
)

func CreateResourceGroupWithMapAndCreatorAccount() (maps.Account, maps.Creator) {
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

	creator, err := CreateCreatorsAccount(ctx, accountName, accountName+"-creator")
	if err != nil {
		util.LogAndPanic(err)
	}
	util.PrintAndLog("creator account created")

	return account, creator
}

func Authenticate(account *maps.AccountsClient, ctx context.Context, accountName string, usesADAuth bool) (azcore.Credential, error) {
	var (
		cred    azcore.Credential
		credErr error
	)

	if usesADAuth {
		cred, credErr = azidentity.NewDefaultAzureCredential(nil)
	} else {
		keysResp, keysErr := account.ListKeys(ctx, config.GroupName(), accountName)
		if keysErr != nil {
			credErr = keysErr
		} else {
			cred = creator.SharedKeyCredential{SubscriptionKey: *keysResp.PrimaryKey}
		}
	}

	return cred, credErr
}
