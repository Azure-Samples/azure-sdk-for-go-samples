package maps

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/maps/mgmt/2021-02-01/maps"

	"github.com/Azure/go-autorest/autorest"
	"github.com/Azure/go-autorest/autorest/to"
)

func getCreatorsAccount() maps.CreatorsClient {
	creatorClient := maps.NewCreatorsClient(config.SubscriptionID())
	a, _ := iam.GetResourceManagementAuthorizer()
	creatorClient.Authorizer = a
	creatorClient.AddToUserAgent(config.UserAgent())
	return creatorClient
}

func CreateCreatorsAccount(ctx context.Context, accountName string, creatorName string) (maps.Creator, error) {
	creatorsAccount := getCreatorsAccount()
	return creatorsAccount.CreateOrUpdate(
		ctx,
		config.GroupName(),
		accountName,
		creatorName,
		maps.Creator{
			Location: to.StringPtr(config.Location()),
			Properties: &maps.CreatorProperties{
				StorageUnits: to.Int32Ptr(1),
			},
		},
	)
}

func DeleteCreatorsAccount(ctx context.Context, accountName string, creatorName string) (autorest.Response, error) {
	creatorsAccount := getCreatorsAccount()
	return creatorsAccount.Delete(ctx, config.GroupName(), accountName, creatorName)
}
