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
	location           = "westus"
	resourceGroupName  = "sample-resource-group"
	appServicePlanName = "sample-web-plan"
	certificateName    = "sample-certificate"
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

	certificate, err := createCertificate(ctx, cred, *appServicePlan.ID)
	if err != nil {
		log.Fatal(err)
	}
	log.Println("web app:", *certificate.ID)

	//keepResource := os.Getenv("KEEP_RESOURCE")
	//if len(keepResource) == 0 {
	//	_, err := cleanup(ctx, conn)
	//	if err != nil {
	//		log.Fatal(err)
	//	}
	//	log.Println("cleaned up successfully.")
	//}
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
				Name:     to.StringPtr("P1V2"),
				Capacity: to.Int32Ptr(1),
			},
			Properties: &armweb.AppServicePlanProperties{
				PerSiteScaling:          to.BoolPtr(false),
				IsXenon:                 to.BoolPtr(false),
				FreeOfferExpirationTime: to.TimePtr(time.Now().AddDate(0, 0, 7)),
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

func createCertificate(ctx context.Context, cred azcore.TokenCredential, appServicePlanID string) (*armweb.Certificate, error) {
	certificatesClient := armweb.NewCertificatesClient(subscriptionID, cred, nil)
	500
	resp, err := certificatesClient.CreateOrUpdate(
		ctx,
		resourceGroupName,
		certificateName,
		armweb.Certificate{
			Resource: armweb.Resource{
				Location: to.StringPtr(location),
			},
			Properties: &armweb.CertificateProperties{
				HostNames: []*string{
					to.StringPtr("ServerCert"),
				},
				Password:     to.StringPtr("QWE!@#123$"),
				ServerFarmID: to.StringPtr(appServicePlanID),
			},
		},
		nil,
	)
	if err != nil {
		return nil, err
	}

	return &resp.Certificate, nil
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
