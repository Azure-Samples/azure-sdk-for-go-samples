package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/web/armweb"
)

var (
	subscriptionID     string
	location           = "eastus"
	resourceGroupName  = "sample-resource-group2"
	appServicePlanName = "sample-web-planx"
	webAppName         = "sample-web-app"
	slotName           = "sample-slot"
)

func main() {
	subscriptionID = os.Getenv("AZURE_SUBSCRIPTION_ID")
	if len(subscriptionID) == 0 {
		log.Fatal("AZURE_SUBSCRIPTION_ID is not set.")
	}

	cred, err := azidentity.NewDefaultAzureCredential(nil)
	if err != nil {
		log.Fatal(err)
	}
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	appServicePlan, err := createAppServicePlan(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("app service plan:", *appServicePlan.ID)

	// If encounter missing error information, it may be that webAppName has already been used.
	webApp, err := createWebApp(ctx, cred, *appServicePlan.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("web app:", *webApp.ID)

	webApp, err = getWebApp(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get web app:", *webApp.ID)

	webAppSlot, err := createWebAppSlot(ctx, cred, *appServicePlan.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("web app slot:", *webAppSlot.ID)

	webAppSlot, err = getWebAppSlot(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get web app slot:", *webAppSlot.ID)

	appConfiguration, err := getAppConfiguration(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("app configuration:", *appConfiguration.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createAppServicePlan(ctx context.Context, cred azcore.TokenCredential) (*armweb.AppServicePlan, error) {
	appServicePlansClient := armweb.NewAppServicePlansClient(subscriptionID, cred, nil)

	pollerResp, err := appServicePlansClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		appServicePlanName,
		armweb.AppServicePlan{
			Resource: armweb.Resource{
				Location: to.StringPtr(location),
			},
			SKU: &armweb.SKUDescription{
				Name:     to.StringPtr("B1"),
				Capacity: to.Int32Ptr(1),
				Tier:     to.StringPtr("BASIC"),
			},
			Properties: &armweb.AppServicePlanProperties{
				PerSiteScaling: to.BoolPtr(false),
				IsXenon:        to.BoolPtr(false),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return &resp.AppServicePlan, nil
}

func createWebApp(ctx context.Context, cred azcore.TokenCredential, appServicePlanID string) (*armweb.Site, error) {
	appsClient := armweb.NewWebAppsClient(subscriptionID, cred, nil)
	pollerResp, err := appsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		webAppName,
		armweb.Site{
			Resource: armweb.Resource{
				Location: to.StringPtr(location),
			},
			Properties: &armweb.SiteProperties{
				ServerFarmID: to.StringPtr(appServicePlanID),
				Reserved:     to.BoolPtr(false),
				IsXenon:      to.BoolPtr(false),
				HyperV:       to.BoolPtr(false),
				SiteConfig: &armweb.SiteConfig{
					NetFrameworkVersion: to.StringPtr("v4.6"),
					AppSettings: []*armweb.NameValuePair{
						{
							Name:  to.StringPtr("WEBSITE_NODE_DEFAULT_VERSION"),
							Value: to.StringPtr("10.14"),
						},
					},
					LocalMySQLEnabled: to.BoolPtr(false),
					Http20Enabled:     to.BoolPtr(true),
				},
				ScmSiteAlsoStopped: to.BoolPtr(false),
				HTTPSOnly:          to.BoolPtr(false),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return &resp.Site, nil
}

func createWebAppSlot(ctx context.Context, cred azcore.TokenCredential, appServicePlanID string) (*armweb.Site, error) {
	appsClient := armweb.NewWebAppsClient(subscriptionID, cred, nil)
	pollerResp, err := appsClient.BeginCreateOrUpdateSlot(
		ctx,
		resourceGroupName,
		webAppName,
		slotName,
		armweb.Site{
			Resource: armweb.Resource{
				Location: to.StringPtr(location),
			},
			Properties: &armweb.SiteProperties{
				ServerFarmID: to.StringPtr(appServicePlanID),
				Reserved:     to.BoolPtr(false),
				IsXenon:      to.BoolPtr(false),
				HyperV:       to.BoolPtr(false),
				SiteConfig: &armweb.SiteConfig{
					NetFrameworkVersion: to.StringPtr("v4.6"),
					LocalMySQLEnabled:   to.BoolPtr(false),
					Http20Enabled:       to.BoolPtr(true),
				},
				ScmSiteAlsoStopped: to.BoolPtr(false),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}
	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}

	return &resp.Site, nil
}

func getWebApp(ctx context.Context, cred azcore.TokenCredential) (*armweb.Site, error) {
	appsClient := armweb.NewWebAppsClient(subscriptionID, cred, nil)

	resp, err := appsClient.Get(ctx, resourceGroupName, webAppName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Site, nil
}

func getWebAppSlot(ctx context.Context, cred azcore.TokenCredential) (*armweb.Site, error) {
	appsClient := armweb.NewWebAppsClient(subscriptionID, cred, nil)

	resp, err := appsClient.GetSlot(ctx, resourceGroupName, webAppName, slotName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Site, nil
}

func getAppConfiguration(ctx context.Context, cred azcore.TokenCredential) (*armweb.SiteConfigResource, error) {
	appsClient := armweb.NewWebAppsClient(subscriptionID, cred, nil)

	resp, err := appsClient.GetConfiguration(ctx, resourceGroupName, webAppName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.SiteConfigResource, nil
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.StringPtr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, cred azcore.TokenCredential) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return nil, err
	}

	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return nil, err
	}
	return resp.RawResponse, nil
}
