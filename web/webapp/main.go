package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"time"

	"github.com/Azure/azure-sdk-for-go/sdk/azcore/arm"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/policy"
	"github.com/Azure/azure-sdk-for-go/sdk/azcore/to"
	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/resources/armresources"
	"github.com/Azure/azure-sdk-for-go/sdk/resourcemanager/web/armweb"
)

var (
	subscriptionID     string
	location           = "westus"
	resourceGroupName  = "sample-resource-group"
	appServicePlanName = "sample-web-plan"
	webAppName         = "sample-web-app"
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

	conn := arm.NewDefaultConnection(cred, &arm.ConnectionOptions{
		Logging: policy.LogOptions{
			IncludeBody: true,
		},
	})
	ctx := context.Background()

	resourceGroup, err := createResourceGroup(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("resources group:", *resourceGroup.ID)

	appServicePlan, err := getAppServicePlan(ctx, conn)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("app service plan:", *appServicePlan.ID)

	webApp, err := createWebApp(ctx, conn, *appServicePlan.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("web app:", *webApp.ID)

	keepResource := os.Getenv("KEEP_RESOURCE")
	if len(keepResource) == 0 {
		_, err := cleanup(ctx, conn)
		if err != nil {
			log.Fatal(err)
		}
		log.Println("cleaned up successfully.")
	}
}

//func createAppServicePlan(ctx context.Context, conn *arm.Connection) (*armweb.AppServicePlan, error) {
//	appServicePlansClient := armweb.NewAppServicePlansClient(conn, subscriptionID)
//
//	pollerResp, err := appServicePlansClient.BeginCreateOrUpdate(
//		ctx,
//		resourceGroupName,
//		appServicePlanName,
//		armweb.AppServicePlan{
//			Resource: armweb.Resource{
//				Location: to.StringPtr(location),
//			},
//			SKU: &armweb.SKUDescription{
//				Name:     to.StringPtr("P1V2"),
//				Capacity: to.Int32Ptr(1),
//			},
//			Properties: &armweb.AppServicePlanProperties{
//				//FreeOfferExpirationTime: to.TimePtr(time.Now()),
//				//PerSiteScaling:          to.BoolPtr(false),
//				//IsXenon:                 to.BoolPtr(false),
//			},
//		},
//		nil,
//	)
//	if err != nil {
//		return nil, err
//	}
//	resp, err := pollerResp.PollUntilDone(ctx, 10*time.Second)
//	if err != nil {
//		return nil, err
//	}
//	return &resp.AppServicePlan, nil
//}

func getAppServicePlan(ctx context.Context, conn *arm.Connection) (*armweb.AppServicePlan, error) {
	appServicePlansClient := armweb.NewAppServicePlansClient(conn, subscriptionID)

	resp, err := appServicePlansClient.Get(
		ctx,
		resourceGroupName,
		appServicePlanName,
		nil,
	)
	if err != nil {
		return nil, err
	}
	return &resp.AppServicePlan, nil
}

func createWebApp(ctx context.Context, conn *arm.Connection, appServicePlanID string) (*armweb.Site, error) {
	appsClient := armweb.NewWebAppsClient(conn, subscriptionID)

	pollerResp, err := appsClient.BeginCreateOrUpdate(
		ctx,
		resourceGroupName,
		webAppName,
		armweb.Site{
			Resource: armweb.Resource{
				Location: to.StringPtr(location),
			},
			Properties: &armweb.SiteProperties{
				SiteConfig: &armweb.SiteConfig{
					JavaVersion:          to.StringPtr("8"),
					JavaContainer:        to.StringPtr("tomcat"),
					JavaContainerVersion: to.StringPtr("8.5"),
					WindowsFxVersion:     to.StringPtr("10"),
				},
				ServerFarmID: to.StringPtr(appServicePlanID),
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

func getAppConfiguration(ctx context.Context, conn *arm.Connection) (*armweb.SiteConfigResource, error) {
	appsClient := armweb.NewWebAppsClient(conn, subscriptionID)

	resp, err := appsClient.GetConfiguration(ctx, resourceGroupName, webAppName, nil)
	if err != nil {
		return nil, err
	}
	return &resp.SiteConfigResource, nil
}

func createResourceGroup(ctx context.Context, conn *arm.Connection) (*armresources.ResourceGroup, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

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

func cleanup(ctx context.Context, conn *arm.Connection) (*http.Response, error) {
	resourceGroupClient := armresources.NewResourceGroupsClient(conn, subscriptionID)

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
