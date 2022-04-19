// Copyright (c) Microsoft Corporation. All rights reserved.
// Licensed under the MIT License. See License.txt in the project root for license information.

package main

import (
	"context"
	"log"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/appservice/armappservice"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
)

var (
	subscriptionID     string
	location           = "eastus"
	resourceGroupName  = "sample-resource-group"
	appServicePlanName = "sample-appservice-planx"
	appServiceName     = "sample-appservice-appxyz"
	slotName           = "sample-slotxyz"
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

	// If encounter missing error information, it may be that appServiceName has already been used.
	appService, err := createWebApp(ctx, cred, *appServicePlan.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("appservice app:", *appService.ID)

	appService, err = getWebApp(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get appservice app:", *appService.ID)

	appServiceSlot, err := createWebAppSlot(ctx, cred, *appServicePlan.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("appservice app slot:", *appServiceSlot.ID)

	appServiceSlot, err = getWebAppSlot(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("get appservice app slot:", *appServiceSlot.ID)

	appConfiguration, err := getAppConfiguration(ctx, cred)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("app configuration:", *appConfiguration.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		err = cleanup(ctx, cred)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

func createAppServicePlan(ctx context.Context, cred azcore.TokenCredential) (*armappservice.Plan, error) {
	appServicePlansClient, err := armappservice.NewPlansClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := appServicePlansClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		appServicePlanName,
		armappservice.Plan{
			Location: to.Ptr(location),
			SKU: &armappservice.SKUDescription{
				Name:     to.Ptr("S1"),
				Capacity: to.Ptr[int32](1),
				Tier:     to.Ptr("STANDARD"),
			},
			Properties: &armappservice.PlanProperties{
				PerSiteScaling: to.Ptr(false),
				IsXenon:        to.Ptr(false),
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
	return &resp.Plan, nil
}

func createWebApp(ctx context.Context, cred azcore.TokenCredential, appServicePlanID string) (*armappservice.Site, error) {
	appsClient, err := armappservice.NewWebAppsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := appsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		appServiceName,
		armappservice.Site{
			Location: to.Ptr(location),
			Properties: &armappservice.SiteProperties{
				ServerFarmID: to.Ptr(appServicePlanID),
				Reserved:     to.Ptr(false),
				IsXenon:      to.Ptr(false),
				HyperV:       to.Ptr(false),
				SiteConfig: &armappservice.SiteConfig{
					NetFrameworkVersion: to.Ptr("v4.6"),
					AppSettings: []*armappservice.NameValuePair{
						{
							Name:  to.Ptr("WEBSITE_NODE_DEFAULT_VERSION"),
							Value: to.Ptr("10.14"),
						},
					},
					LocalMySQLEnabled: to.Ptr(false),
					Http20Enabled:     to.Ptr(true),
				},
				ScmSiteAlsoStopped: to.Ptr(false),
				HTTPSOnly:          to.Ptr(false),
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

func createWebAppSlot(ctx context.Context, cred azcore.TokenCredential, appServicePlanID string) (*armappservice.Site, error) {
	appsClient, err := armappservice.NewWebAppsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	pollerResp, err := appsClient.BeginCreateOrUpdateSlot(
		ctx,
		resourceGroupName,
		appServiceName,
		slotName,
		armappservice.Site{
			Location: to.Ptr(location),
			Properties: &armappservice.SiteProperties{
				ServerFarmID: to.Ptr(appServicePlanID),
				Reserved:     to.Ptr(false),
				IsXenon:      to.Ptr(false),
				HyperV:       to.Ptr(false),
				SiteConfig: &armappservice.SiteConfig{
					NetFrameworkVersion: to.Ptr("v4.6"),
					LocalMySQLEnabled:   to.Ptr(false),
					Http20Enabled:       to.Ptr(true),
				},
				ScmSiteAlsoStopped: to.Ptr(false),
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

func getWebApp(ctx context.Context, cred azcore.TokenCredential) (*armappservice.Site, error) {
	appsClient, err := armappservice.NewWebAppsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := appsClient.Get(ctx, resourceGroupName, appServiceName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Site, nil
}

func getWebAppSlot(ctx context.Context, cred azcore.TokenCredential) (*armappservice.Site, error) {
	appsClient, err := armappservice.NewWebAppsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := appsClient.GetSlot(ctx, resourceGroupName, appServiceName, slotName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.Site, nil
}

func getAppConfiguration(ctx context.Context, cred azcore.TokenCredential) (*armappservice.SiteConfigResource, error) {
	appsClient, err := armappservice.NewWebAppsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resp, err := appsClient.GetConfiguration(ctx, resourceGroupName, appServiceName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.SiteConfigResource, nil
}

func createResourceGroup(ctx context.Context, cred azcore.TokenCredential) (*armresources.ResourceGroup, error) {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return nil, err
	}

	resourceGroupResp, err := resourceGroupClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		armresources.ResourceGroup{
			Location: to.Ptr(location),
		},
		nil)
	if err != nil {
		return nil, err
	}
	return &resourceGroupResp.ResourceGroup, nil
}

func cleanup(ctx context.Context, cred azcore.TokenCredential) error {
	resourceGroupClient, err := armresources.NewResourceGroupsClient(subscriptionID, cred, nil)
	if err != nil {
		return err
	}

	pollerResp, err := resourceGroupClient.BeginDelete(ctx, resourceGroupName, nil)
	if err != nil {
		return err
	}

	_, err = pollerResp.PollUntilDone(ctx, 10*time.Second)
	if err != nil {
		return err
	}
	return nil
}
