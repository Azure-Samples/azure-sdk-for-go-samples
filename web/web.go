package web

import (
	"context"
	"fmt"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/helpers"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/iam"
	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2016-09-01/web"
	"github.com/Azure/go-autorest/autorest/to"
	"github.com/Azure/go-autorest/autorest"
)

func CreateContainerSite(ctx context.Context, name, image string) (createdConfig web.SiteConfigResource, err error) {
	client := web.NewAppsClient(helpers.SubscriptionID())
	client.Authorizer, err = iam.GetResourceManagementAuthorizer(iam.AuthGrantType())
	if err != nil {
		return
	}
	client.AddToUserAgent(helpers.UserAgent())

	future, err := client.CreateOrUpdate(
		ctx,
		helpers.ResourceGroupName(),
		name,
		web.Site{
			Location: to.StringPtr(helpers.Location()),
			SiteProperties: &web.SiteProperties{
				SiteConfig: &web.SiteConfig{
					LinuxFxVersion: to.StringPtr(fmt.Sprintf("DOCKER|%s", image)),
				},
			},
		})
	if err != nil {
		if cast, ok := err.(autorest.DetailedError); ok {
			fmt.Println("Status: ", cast.Response.Status)
			fmt.Println("Headers:")
			for name, value := range cast.Response.Header {
				fmt.Printf("\t%q -> %q\n", name, value)
			}

			fmt.Println("Decoded Body:\n\t", cast.Original)
		}
		return
	}

	err = future.WaitForCompletion(ctx, client.Client)
	if err != nil {
		return
	}

	createdConfig, err = client.GetConfiguration(ctx, helpers.ResourceGroupName(), name)
	return
}

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
