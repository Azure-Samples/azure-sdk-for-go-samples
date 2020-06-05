package web

import (
	"context"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2016-09-01/web"
	"github.com/Azure/go-autorest/autorest/to"
)

func getWebAppsClient() (client web.AppsClient, err error) {
	client = web.NewAppsClient(config.SubscriptionID())
	client.Authorizer, err = iam.GetResourceManagementAuthorizer()
	if err != nil {
		return
	}
	client.AddToUserAgent(config.UserAgent())
	return
}

// CreateWebApp creates a blank web app with specified name
func CreateWebApp(ctx context.Context, name string) (webSite web.Site, err error) {
	client, err := getWebAppsClient()
	if err != nil {
		return
	}
	future, err := client.CreateOrUpdate(
		ctx,
		config.GroupName(),
		name,
		web.Site{
			Location:       to.StringPtr(config.Location()),
			SiteProperties: &web.SiteProperties{},
		})
	if err != nil {
		return
	}

	err = future.WaitForCompletionRef(ctx, client.Client)
	if err != nil {
		return
	}
	webSite, err = future.Result(client)
	return
}

// GetAppConfiguration returns web app configuration info.
func GetAppConfiguration(ctx context.Context, name string) (createdConfig web.SiteConfigResource, err error) {
	client, err := getWebAppsClient()
	if err != nil {
		return
	}
	createdConfig, err = client.GetConfiguration(ctx, config.GroupName(), name)
	return
}
