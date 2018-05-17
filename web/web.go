package web

import (
	"context"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2016-09-01/web"
	"github.com/Azure/go-autorest/autorest/to"
)

func CreateAppServicePlan(ctx context.Context, name, kind string) (created web.AppServicePlan, err error) {
	client := web.NewAppServicePlansClient(helpers.SubscriptionID())
	client.Authorizer, err = iam.GetResourceManagementAuthorizer(iam.AuthGrantType())
	if err != nil {
		return
	}
	client.AddToUserAgent(helpers.UserAgent())

	future, err := client.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
		name,
		web.AppServicePlan{
			Location: to.StringPtr(helpers.Location()),
			Kind: &kind,
		},
	)
	if err != nil {
		return
	}

	err = future.WaitForCompletion(ctx, client.Client)
	if err != nil {
		return
	}

	created, err = future.Result(client)
	return
}
