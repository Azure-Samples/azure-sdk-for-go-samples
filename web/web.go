package web

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/services/web/mgmt/2016-09-01/web"

	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/config"
	"github.com/Azure-Samples/azure-sdk-for-go-samples/internal/iam"
	"github.com/Azure/go-autorest/autorest/to"
)

// CreateContainerSite provisions all infrastructure needed to run an Azure Web App for Containers.
func CreateContainerSite(ctx context.Context, name, image string) (createdConfig web.SiteConfigResource, err error) {
	client := web.NewAppsClient(config.SubscriptionID())
	client.Authorizer, err = iam.GetResourceManagementAuthorizer()
	if err != nil {
		return
	}
	client.AddToUserAgent(config.UserAgent())

	future, err := client.CreateOrUpdate(
		ctx,
		config.GroupName(),
		name,
		web.Site{
			Location: to.StringPtr(config.Location()),
			SiteProperties: &web.SiteProperties{
				SiteConfig: &web.SiteConfig{
					LinuxFxVersion: to.StringPtr(fmt.Sprintf("DOCKER|%s", image)),
				},
			},
		})
	if err != nil {
		return
	}

	err = future.WaitForCompletion(ctx, client.Client)
	if err != nil {
		return
	}

	createdConfig, err = client.GetConfiguration(ctx, config.GroupName(), name)
	return
}
